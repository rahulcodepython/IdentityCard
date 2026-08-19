package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
	qrcode "github.com/skip2/go-qrcode"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/entities"
	"identitycard-server/internal/generic"
	"identitycard-server/internal/pkg/jwt"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/oauth"
	"identitycard-server/internal/pkg/postgres"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
)

const (
	refreshKeyPrefix = "refresh_token:"
	totpIssuer       = "IdentityCard"
	otpTTL           = 5 * time.Minute

	// oauthExchangeKeyPrefix/oauthExchangeTTL back HandleGoogleCallback's
	// one-time code handoff to the Next.js callback route — see there.
	oauthExchangeKeyPrefix = "oauth_exchange:"
	oauthExchangeTTL       = 30 * time.Second
)

// oauthExchangePayload is what HandleGoogleCallback stashes in Redis behind
// a one-time code and ExchangeOAuthCode reads back out.
type oauthExchangePayload struct {
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	AccessTokenTTLSeconds  int    `json:"access_ttl_seconds"`
	RefreshTokenTTLSeconds int    `json:"refresh_ttl_seconds"`
}

type AuthService struct {
	cfg     *config.Config
	users   *repositories.AuthRepository
	members *repositories.MembersRepository
	orgs    *OrganizationsService
	rdb     *redis.Client
	google  *oauth.Client
	mail    *mailer.Mailer

	// Used only by Register and CreateOrganization, which write across more
	// than one table atomically — everything else in this service works
	// through the repositories above.
	pool        *pgxpool.Pool
	baseQueries *dbgen.Queries
}

func NewAuthService(
	cfg *config.Config,
	users *repositories.AuthRepository,
	members *repositories.MembersRepository,
	orgs *OrganizationsService,
	pool *pgxpool.Pool,
	baseQueries *dbgen.Queries,
	rdb *redis.Client,
	google *oauth.Client,
	mail *mailer.Mailer,
) *AuthService {
	return &AuthService{
		cfg: cfg, users: users, members: members, orgs: orgs,
		pool: pool, baseQueries: baseQueries, rdb: rdb, google: google, mail: mail,
	}
}

// Register creates the organization (unique slug) and its first,
// unverified user with a freshly generated TOTP secret — all in one
// transaction. No plan is chosen and no subscription created; see
// PlansService, done later from the dashboard. No session is issued yet —
// the account can't sign in until VerifyOTP or VerifyTOTP succeeds.
func (s *AuthService) Register(ctx context.Context, req entities.RegisterRequest) (entities.RegisterResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: totpIssuer, AccountName: req.Email})
	if err != nil {
		return entities.RegisterResponse{}, utils.ErrInternal()
	}

	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.baseQueries.WithTx(tx)

		orgRepo := repositories.NewOrganizationsRepository(q)
		slug, txErr := GenerateUniqueSlug(ctx, orgRepo, req.OrganizationName)
		if txErr != nil {
			return utils.ErrInternal()
		}
		org, txErr := orgRepo.Create(ctx, req.OrganizationName, slug)
		if txErr != nil {
			return utils.ErrInternal()
		}

		user, txErr := repositories.NewAuthRepository(q).CreateUser(ctx, req.Email, req.Name, key.Secret())
		if txErr != nil {
			if postgres.IsUniqueViolation(txErr, "users_email_key") {
				return utils.NewError(http.StatusConflict, "email_taken", "an account with this email already exists")
			}
			return utils.ErrInternal()
		}

		if _, txErr := repositories.NewMembersRepository(q).Create(ctx, org.ID, user.ID, []string{string(generic.RoleSuperAdmin)}); txErr != nil {
			return utils.ErrInternal()
		}
		return nil
	})
	if txErr != nil {
		return entities.RegisterResponse{}, txErr
	}

	qrImage, err := qrDataURI(key.String())
	if err != nil {
		return entities.RegisterResponse{}, utils.ErrInternal()
	}
	return entities.RegisterResponse{Email: req.Email, TOTPQRImage: qrImage, TOTPSecret: key.Secret()}, nil
}

// SendOTP emails a fresh 6-digit code, good for otpTTL. The same code
// completes either a pending registration or a login — VerifyOTP treats
// both the same way, see there.
func (s *AuthService) SendOTP(ctx context.Context, email string) error {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return utils.ErrNotFound("account")
	}

	code, err := generateAuthOTP()
	if err != nil {
		return utils.ErrInternal()
	}
	if _, err := s.users.SetOTP(ctx, user.ID, code, time.Now().Add(otpTTL)); err != nil {
		return utils.ErrInternal()
	}

	body := fmt.Sprintf("Your IdentityCard verification code is %s. It expires in %d minutes.", code, int(otpTTL.Minutes()))
	if err := s.mail.Send(email, "Your IdentityCard verification code", body); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

// VerifyOTP validates a code sent by SendOTP and issues a session. Whether
// this completes a pending registration or logs an already-verified user
// in is decided by the DB, not the caller — VerifyAndConsumeOTP only ever
// sets email_verified_at if it was still null.
func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	user, err := s.users.VerifyAndConsumeOTP(ctx, email, code)
	if err != nil {
		return "", "", 0, 0, utils.NewError(http.StatusBadRequest, "invalid_otp", "code is invalid or expired")
	}
	return s.issueSessionForUser(ctx, user)
}

// VerifyTOTP validates a code from the user's authenticator app. For a
// signup still in progress, the QR Register returned is what the app
// scanned to get this secret; at login, it must have been scanned during
// that user's own signup, or every code fails and they fall back to OTP.
func (s *AuthService) VerifyTOTP(ctx context.Context, email, code string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, utils.NewError(http.StatusBadRequest, "invalid_totp", "code is invalid")
	}
	if !user.TotpSecret.Valid {
		return "", "", 0, 0, utils.NewError(http.StatusBadRequest, "totp_not_configured", "authenticator sign-in isn't set up for this account — use email code instead")
	}
	if !totp.Validate(code, user.TotpSecret.String) {
		return "", "", 0, 0, utils.NewError(http.StatusBadRequest, "invalid_totp", "code is invalid")
	}

	if !user.EmailVerifiedAt.Valid {
		verified, err := s.users.MarkEmailVerified(ctx, user.ID)
		if err != nil {
			return "", "", 0, 0, utils.ErrInternal()
		}
		user = verified
	}
	return s.issueSessionForUser(ctx, user)
}

// CreateOrganization finishes onboarding for a signed-in user who doesn't
// have an organization yet — a Google signup, which (unlike email
// registration) never collects an org name up front since the OAuth round
// trip is a plain redirect. Mints a fresh token pair carrying the new
// org_id so the session updates immediately, no re-login required.
func (s *AuthService) CreateOrganization(ctx context.Context, claims *jwt.AccessClaims, req entities.CreateOrganizationRequest) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	if claims.OrganizationID != uuid.Nil {
		return "", "", 0, 0, utils.NewError(http.StatusConflict, "organization_already_exists", "this account already belongs to an organization")
	}

	var orgID uuid.UUID
	txErr := postgres.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.baseQueries.WithTx(tx)

		orgRepo := repositories.NewOrganizationsRepository(q)
		slug, txErr := GenerateUniqueSlug(ctx, orgRepo, req.OrganizationName)
		if txErr != nil {
			return utils.ErrInternal()
		}
		org, txErr := orgRepo.Create(ctx, req.OrganizationName, slug)
		if txErr != nil {
			return utils.ErrInternal()
		}
		if _, txErr := repositories.NewMembersRepository(q).Create(ctx, org.ID, claims.UserID, []string{string(generic.RoleSuperAdmin)}); txErr != nil {
			return utils.ErrInternal()
		}
		orgID = org.ID
		return nil
	})
	if txErr != nil {
		return "", "", 0, 0, txErr
	}

	return s.issueTokenPair(ctx, claims.UserID, orgID, []string{string(generic.RoleSuperAdmin)})
}

// BuildGoogleAuthURL returns the Google consent-screen URL, with intent
// folded into a signed, short-lived state token that survives the
// redirect round trip.
func (s *AuthService) BuildGoogleAuthURL(intent string) (string, error) {
	if intent != "login" && intent != "register" {
		return "", utils.NewError(http.StatusBadRequest, "bad_request", "invalid intent")
	}

	state, err := signOAuthState(s.cfg.JWTSecret, intent)
	if err != nil {
		return "", utils.ErrInternal()
	}

	url, err := s.google.AuthURL(state)
	if err != nil {
		if errors.Is(err, oauth.ErrNotConfigured) {
			return "", utils.NewError(http.StatusServiceUnavailable, "oauth_not_configured", "Google sign-in is not configured")
		}
		return "", utils.ErrInternal()
	}
	return url, nil
}

// HandleGoogleCallback exchanges the authorization code, fetches the
// Google profile, and either logs an existing user in or creates a new,
// org-less user pending onboarding (see CreateOrganization). Rather than
// setting a cookie itself (there is no browser-visible Go domain to set
// one on anymore — see internal/middlewares.RequireAuth), it mints a
// short-lived, single-use exchange code, stashes the real token pair
// behind it in Redis, and returns just the code — the Next.js callback
// route trades it for the tokens server-to-server via ExchangeOAuthCode
// and sets the session cookie on its own domain. intent is returned even
// on error so the handler can send the user back to the right page
// (/login vs /register).
func (s *AuthService) HandleGoogleCallback(ctx context.Context, code, state string) (exchangeCode, intent string, err error) {
	claims, err := parseOAuthState(s.cfg.JWTSecret, state)
	if err != nil {
		return "", "", utils.NewError(http.StatusBadRequest, "invalid_state", "invalid or expired oauth state")
	}
	intent = claims.Intent

	info, err := s.google.ExchangeAndFetchUserInfo(ctx, code)
	if err != nil || info.Email == "" {
		return "", intent, utils.NewError(http.StatusBadGateway, "oauth_failed", "failed to fetch google profile")
	}

	var access, refresh string
	var accessTTL, refreshTTL time.Duration

	if intent == "register" {
		if _, existsErr := s.users.GetUserByEmail(ctx, info.Email); existsErr == nil {
			return "", intent, utils.NewError(http.StatusConflict, "account_exists", "an account with this email already exists")
		}

		user, createErr := s.users.CreateOAuthUser(ctx, info.Email, info.Name, info.Sub, info.Picture)
		if createErr != nil {
			return "", intent, utils.ErrInternal()
		}
		access, refresh, accessTTL, refreshTTL, err = s.issueTokenPair(ctx, user.ID, uuid.Nil, nil)
		if err != nil {
			return "", intent, err
		}
	} else {
		user, lookupErr := s.users.GetUserByGoogleID(ctx, info.Sub)
		if lookupErr != nil {
			// Not linked yet — fall back to email match and auto-link, since
			// Google has already verified this email belongs to this person.
			user, lookupErr = s.users.GetUserByEmail(ctx, info.Email)
			if lookupErr != nil {
				return "", intent, utils.NewError(http.StatusNotFound, "google_no_account", "no account found for this Google email")
			}
			if !user.GoogleID.Valid {
				user, lookupErr = s.users.LinkGoogleID(ctx, user.ID, info.Sub)
				if lookupErr != nil {
					return "", intent, utils.ErrInternal()
				}
			}
		}

		access, refresh, accessTTL, refreshTTL, err = s.issueSessionForUser(ctx, user)
		if err != nil {
			return "", intent, err
		}
	}

	payload, err := json.Marshal(oauthExchangePayload{
		AccessToken:            access,
		RefreshToken:           refresh,
		AccessTokenTTLSeconds:  int(accessTTL.Seconds()),
		RefreshTokenTTLSeconds: int(refreshTTL.Seconds()),
	})
	if err != nil {
		return "", intent, utils.ErrInternal()
	}

	exchangeCode = uuid.NewString()
	if err := s.rdb.Set(ctx, oauthExchangeKeyPrefix+exchangeCode, payload, oauthExchangeTTL).Err(); err != nil {
		return "", intent, utils.ErrInternal()
	}
	return exchangeCode, intent, nil
}

// ExchangeOAuthCode trades a one-time code minted by HandleGoogleCallback
// for the real token pair. Single-use: the Redis key is deleted on read,
// whether or not the payload turns out to be valid.
func (s *AuthService) ExchangeOAuthCode(ctx context.Context, code string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	key := oauthExchangeKeyPrefix + code
	data, getErr := s.rdb.Get(ctx, key).Result()
	if getErr != nil {
		return "", "", 0, 0, utils.NewError(http.StatusBadRequest, "invalid_or_expired_code", "code is invalid or expired")
	}
	s.rdb.Del(ctx, key)

	var payload oauthExchangePayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}
	return payload.AccessToken, payload.RefreshToken,
		time.Duration(payload.AccessTokenTTLSeconds) * time.Second,
		time.Duration(payload.RefreshTokenTTLSeconds) * time.Second,
		nil
}

// Refresh rotates the refresh token: the presented one is invalidated
// immediately, whether or not this call succeeds, so a stolen-and-reused
// token fails closed.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (newAccess, newRefresh string, accessTTL, refreshTTL time.Duration, err error) {
	claims, err := jwt.ParseRefreshToken(s.cfg.JWTSecret, refreshToken)
	if err != nil {
		return "", "", 0, 0, utils.ErrUnauthorized("")
	}

	key := refreshKeyPrefix + claims.ID
	deleted, err := s.rdb.Del(ctx, key).Result()
	if err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}
	if deleted == 0 {
		// Already used or never issued by us (revoked/expired) — reject.
		return "", "", 0, 0, utils.ErrUnauthorized("")
	}

	memberships, err := s.members.ListForUser(ctx, claims.UserID)
	if err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}
	if len(memberships) == 0 {
		return s.issueTokenPair(ctx, claims.UserID, uuid.Nil, nil)
	}
	membership := memberships[0]

	return s.issueTokenPair(ctx, claims.UserID, membership.OrganizationID, membership.Roles)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := jwt.ParseRefreshToken(s.cfg.JWTSecret, refreshToken)
	if err != nil {
		return nil // already invalid/expired — nothing to revoke
	}
	s.rdb.Del(ctx, refreshKeyPrefix+claims.ID)
	return nil
}

func (s *AuthService) Me(ctx context.Context, claims *jwt.AccessClaims) (entities.MeResponse, error) {
	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return entities.MeResponse{}, utils.ErrNotFound("user")
	}

	resp := entities.MeResponse{UserID: user.ID, Email: user.Email, Name: user.Name}
	if claims.OrganizationID == uuid.Nil {
		return resp, nil
	}

	membership, err := s.members.Get(ctx, claims.OrganizationID, claims.UserID)
	if err != nil {
		return entities.MeResponse{}, utils.ErrNotFound("organization membership")
	}
	org, err := s.orgs.GetSettings(ctx, claims.OrganizationID)
	if err != nil {
		return entities.MeResponse{}, err
	}

	orgID := claims.OrganizationID
	resp.HasOrganization = true
	resp.OrganizationID = &orgID
	resp.OrganizationName = &org.Name
	resp.Roles = membership.Roles
	return resp, nil
}

// issueSessionForUser looks up the user's org membership, if any — a
// Google signup can reach here with none yet, pending /onboarding — and
// issues a token pair scoped accordingly.
func (s *AuthService) issueSessionForUser(ctx context.Context, user dbgen.User) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	memberships, err := s.members.ListForUser(ctx, user.ID)
	if err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}
	if len(memberships) == 0 {
		return s.issueTokenPair(ctx, user.ID, uuid.Nil, nil)
	}
	membership := memberships[0]
	return s.issueTokenPair(ctx, user.ID, membership.OrganizationID, membership.Roles)
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID, orgID uuid.UUID, roleStrings []string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	roles := make([]generic.Role, len(roleStrings))
	for i, r := range roleStrings {
		roles[i] = generic.Role(r)
	}

	accessToken, err = jwt.IssueAccessToken(s.cfg.JWTSecret, s.cfg.AccessTokenTTL, userID, orgID, roles)
	if err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}

	refreshToken, jti, err := jwt.IssueRefreshToken(s.cfg.JWTSecret, s.cfg.RefreshTokenTTL, userID)
	if err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}
	if err := s.rdb.Set(ctx, refreshKeyPrefix+jti, userID.String(), s.cfg.RefreshTokenTTL).Err(); err != nil {
		return "", "", 0, 0, utils.ErrInternal()
	}

	return accessToken, refreshToken, s.cfg.AccessTokenTTL, s.cfg.RefreshTokenTTL, nil
}

func qrDataURI(content string) (string, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(png), nil
}
