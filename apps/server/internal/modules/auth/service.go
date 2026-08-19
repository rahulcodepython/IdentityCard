package auth

import (
	"context"
	"encoding/base64"
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

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/mailer"
	"identitycard-server/internal/modules/members"
	"identitycard-server/internal/modules/organizations"
	"identitycard-server/internal/oauth"
)

const (
	refreshKeyPrefix = "refresh_token:"
	totpIssuer       = "IdentityCard"
	otpTTL           = 5 * time.Minute
)

type Service struct {
	cfg     *config.Config
	users   *Repository
	members *members.Repository
	orgs    *organizations.Service
	rdb     *redis.Client
	google  *oauth.Client
	mail    *mailer.Mailer

	// Used only by Register and CreateOrganization, which write across more
	// than one table atomically — everything else in this service works
	// through the repositories above.
	pool        *pgxpool.Pool
	baseQueries *dbgen.Queries
}

func NewService(
	cfg *config.Config,
	users *Repository,
	members *members.Repository,
	orgs *organizations.Service,
	pool *pgxpool.Pool,
	baseQueries *dbgen.Queries,
	rdb *redis.Client,
	google *oauth.Client,
	mail *mailer.Mailer,
) *Service {
	return &Service{
		cfg: cfg, users: users, members: members, orgs: orgs,
		pool: pool, baseQueries: baseQueries, rdb: rdb, google: google, mail: mail,
	}
}

// Register creates the organization (unique slug) and its first,
// unverified user with a freshly generated TOTP secret — all in one
// transaction. No plan is chosen and no subscription created; see
// internal/modules/plans, done later from the dashboard. No session is
// issued yet — the account can't sign in until VerifyOTP or VerifyTOTP
// succeeds.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: totpIssuer, AccountName: req.Email})
	if err != nil {
		return RegisterResponse{}, httpx.ErrInternal()
	}

	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.baseQueries.WithTx(tx)

		orgRepo := organizations.NewRepository(q)
		slug, txErr := organizations.GenerateUniqueSlug(ctx, orgRepo, req.OrganizationName)
		if txErr != nil {
			return httpx.ErrInternal()
		}
		org, txErr := orgRepo.Create(ctx, req.OrganizationName, slug)
		if txErr != nil {
			return httpx.ErrInternal()
		}

		user, txErr := NewRepository(q).CreateUser(ctx, req.Email, req.Name, key.Secret())
		if txErr != nil {
			if db.IsUniqueViolation(txErr, "users_email_key") {
				return httpx.NewError(http.StatusConflict, "email_taken", "an account with this email already exists")
			}
			return httpx.ErrInternal()
		}

		if _, txErr := members.NewRepository(q).Create(ctx, org.ID, user.ID, []string{string(coreauth.RoleSuperAdmin)}); txErr != nil {
			return httpx.ErrInternal()
		}
		return nil
	})
	if txErr != nil {
		return RegisterResponse{}, txErr
	}

	qrImage, err := qrDataURI(key.String())
	if err != nil {
		return RegisterResponse{}, httpx.ErrInternal()
	}
	return RegisterResponse{Email: req.Email, TOTPQRImage: qrImage, TOTPSecret: key.Secret()}, nil
}

// SendOTP emails a fresh 6-digit code, good for otpTTL. The same code
// completes either a pending registration or a login — VerifyOTP treats
// both the same way, see there.
func (s *Service) SendOTP(ctx context.Context, email string) error {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return httpx.ErrNotFound("account")
	}

	code, err := generateOTP()
	if err != nil {
		return httpx.ErrInternal()
	}
	if _, err := s.users.SetOTP(ctx, user.ID, code, time.Now().Add(otpTTL)); err != nil {
		return httpx.ErrInternal()
	}

	body := fmt.Sprintf("Your IdentityCard verification code is %s. It expires in %d minutes.", code, int(otpTTL.Minutes()))
	if err := s.mail.Send(email, "Your IdentityCard verification code", body); err != nil {
		return httpx.ErrInternal()
	}
	return nil
}

// VerifyOTP validates a code sent by SendOTP and issues a session. Whether
// this completes a pending registration or logs an already-verified user
// in is decided by the DB, not the caller — VerifyAndConsumeOTP only ever
// sets email_verified_at if it was still null.
func (s *Service) VerifyOTP(ctx context.Context, email, code string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	user, err := s.users.VerifyAndConsumeOTP(ctx, email, code)
	if err != nil {
		return "", "", 0, 0, httpx.NewError(http.StatusBadRequest, "invalid_otp", "code is invalid or expired")
	}
	return s.issueSessionForUser(ctx, user)
}

// VerifyTOTP validates a code from the user's authenticator app. For a
// signup still in progress, the QR Register returned is what the app
// scanned to get this secret; at login, it must have been scanned during
// that user's own signup, or every code fails and they fall back to OTP.
func (s *Service) VerifyTOTP(ctx context.Context, email, code string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, httpx.NewError(http.StatusBadRequest, "invalid_totp", "code is invalid")
	}
	if !user.TotpSecret.Valid {
		return "", "", 0, 0, httpx.NewError(http.StatusBadRequest, "totp_not_configured", "authenticator sign-in isn't set up for this account — use email code instead")
	}
	if !totp.Validate(code, user.TotpSecret.String) {
		return "", "", 0, 0, httpx.NewError(http.StatusBadRequest, "invalid_totp", "code is invalid")
	}

	if !user.EmailVerifiedAt.Valid {
		verified, err := s.users.MarkEmailVerified(ctx, user.ID)
		if err != nil {
			return "", "", 0, 0, httpx.ErrInternal()
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
func (s *Service) CreateOrganization(ctx context.Context, claims *coreauth.AccessClaims, req CreateOrganizationRequest) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	if claims.OrganizationID != uuid.Nil {
		return "", "", 0, 0, httpx.NewError(http.StatusConflict, "organization_already_exists", "this account already belongs to an organization")
	}

	var orgID uuid.UUID
	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.baseQueries.WithTx(tx)

		orgRepo := organizations.NewRepository(q)
		slug, txErr := organizations.GenerateUniqueSlug(ctx, orgRepo, req.OrganizationName)
		if txErr != nil {
			return httpx.ErrInternal()
		}
		org, txErr := orgRepo.Create(ctx, req.OrganizationName, slug)
		if txErr != nil {
			return httpx.ErrInternal()
		}
		if _, txErr := members.NewRepository(q).Create(ctx, org.ID, claims.UserID, []string{string(coreauth.RoleSuperAdmin)}); txErr != nil {
			return httpx.ErrInternal()
		}
		orgID = org.ID
		return nil
	})
	if txErr != nil {
		return "", "", 0, 0, txErr
	}

	return s.issueTokenPair(ctx, claims.UserID, orgID, []string{string(coreauth.RoleSuperAdmin)})
}

// BuildGoogleAuthURL returns the Google consent-screen URL, with intent
// folded into a signed, short-lived state token that survives the
// redirect round trip.
func (s *Service) BuildGoogleAuthURL(intent string) (string, error) {
	if intent != "login" && intent != "register" {
		return "", httpx.NewError(http.StatusBadRequest, "bad_request", "invalid intent")
	}

	state, err := signOAuthState(s.cfg.JWTSecret, intent)
	if err != nil {
		return "", httpx.ErrInternal()
	}

	url, err := s.google.AuthURL(state)
	if err != nil {
		if errors.Is(err, oauth.ErrNotConfigured) {
			return "", httpx.NewError(http.StatusServiceUnavailable, "oauth_not_configured", "Google sign-in is not configured")
		}
		return "", httpx.ErrInternal()
	}
	return url, nil
}

// HandleGoogleCallback exchanges the authorization code, fetches the
// Google profile, and either logs an existing user in or creates a new,
// org-less user pending onboarding (see CreateOrganization) —
// redirectPath is always "/dashboard": the dashboard layout itself bounces
// an org-less session to /onboarding. intent is returned even on error so
// the handler can send the user back to the right page (/login vs /register).
func (s *Service) HandleGoogleCallback(ctx context.Context, code, state string) (accessToken, refreshToken, redirectPath, intent string, accessTTL, refreshTTL time.Duration, err error) {
	claims, err := parseOAuthState(s.cfg.JWTSecret, state)
	if err != nil {
		return "", "", "", "", 0, 0, httpx.NewError(http.StatusBadRequest, "invalid_state", "invalid or expired oauth state")
	}
	intent = claims.Intent

	token, err := s.google.Exchange(ctx, code)
	if err != nil {
		return "", "", "", intent, 0, 0, httpx.NewError(http.StatusBadGateway, "oauth_failed", "google authorization failed")
	}
	info, err := s.google.FetchUserInfo(ctx, token)
	if err != nil || info.Email == "" {
		return "", "", "", intent, 0, 0, httpx.NewError(http.StatusBadGateway, "oauth_failed", "failed to fetch google profile")
	}

	if intent == "register" {
		if _, existsErr := s.users.GetUserByEmail(ctx, info.Email); existsErr == nil {
			return "", "", "", intent, 0, 0, httpx.NewError(http.StatusConflict, "account_exists", "an account with this email already exists")
		}

		user, createErr := s.users.CreateOAuthUser(ctx, info.Email, info.Name, info.Sub, info.Picture)
		if createErr != nil {
			return "", "", "", intent, 0, 0, httpx.ErrInternal()
		}
		access, refresh, accessTTL, refreshTTL, tokErr := s.issueTokenPair(ctx, user.ID, uuid.Nil, nil)
		return access, refresh, "/dashboard", intent, accessTTL, refreshTTL, tokErr
	}

	user, lookupErr := s.users.GetUserByGoogleID(ctx, info.Sub)
	if lookupErr != nil {
		// Not linked yet — fall back to email match and auto-link, since
		// Google has already verified this email belongs to this person.
		user, lookupErr = s.users.GetUserByEmail(ctx, info.Email)
		if lookupErr != nil {
			return "", "", "", intent, 0, 0, httpx.NewError(http.StatusNotFound, "google_no_account", "no account found for this Google email")
		}
		if !user.GoogleID.Valid {
			user, lookupErr = s.users.LinkGoogleID(ctx, user.ID, info.Sub)
			if lookupErr != nil {
				return "", "", "", intent, 0, 0, httpx.ErrInternal()
			}
		}
	}

	access, refresh, accessTTL, refreshTTL, tokErr := s.issueSessionForUser(ctx, user)
	return access, refresh, "/dashboard", intent, accessTTL, refreshTTL, tokErr
}

// Refresh rotates the refresh token: the presented one is invalidated
// immediately, whether or not this call succeeds, so a stolen-and-reused
// token fails closed.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (newAccess, newRefresh string, accessTTL, refreshTTL time.Duration, err error) {
	claims, err := coreauth.ParseRefreshToken(s.cfg.JWTSecret, refreshToken)
	if err != nil {
		return "", "", 0, 0, httpx.ErrUnauthorized("")
	}

	key := refreshKeyPrefix + claims.ID
	deleted, err := s.rdb.Del(ctx, key).Result()
	if err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
	}
	if deleted == 0 {
		// Already used or never issued by us (revoked/expired) — reject.
		return "", "", 0, 0, httpx.ErrUnauthorized("")
	}

	memberships, err := s.members.ListForUser(ctx, claims.UserID)
	if err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
	}
	if len(memberships) == 0 {
		return s.issueTokenPair(ctx, claims.UserID, uuid.Nil, nil)
	}
	membership := memberships[0]

	return s.issueTokenPair(ctx, claims.UserID, membership.OrganizationID, membership.Roles)
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := coreauth.ParseRefreshToken(s.cfg.JWTSecret, refreshToken)
	if err != nil {
		return nil // already invalid/expired — nothing to revoke
	}
	s.rdb.Del(ctx, refreshKeyPrefix+claims.ID)
	return nil
}

func (s *Service) Me(ctx context.Context, claims *coreauth.AccessClaims) (MeResponse, error) {
	user, err := s.users.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return MeResponse{}, httpx.ErrNotFound("user")
	}

	resp := MeResponse{UserID: user.ID, Email: user.Email, Name: user.Name}
	if claims.OrganizationID == uuid.Nil {
		return resp, nil
	}

	membership, err := s.members.Get(ctx, claims.OrganizationID, claims.UserID)
	if err != nil {
		return MeResponse{}, httpx.ErrNotFound("organization membership")
	}
	org, err := s.orgs.GetSettings(ctx, claims.OrganizationID)
	if err != nil {
		return MeResponse{}, err
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
func (s *Service) issueSessionForUser(ctx context.Context, user dbgen.User) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	memberships, err := s.members.ListForUser(ctx, user.ID)
	if err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
	}
	if len(memberships) == 0 {
		return s.issueTokenPair(ctx, user.ID, uuid.Nil, nil)
	}
	membership := memberships[0]
	return s.issueTokenPair(ctx, user.ID, membership.OrganizationID, membership.Roles)
}

func (s *Service) issueTokenPair(ctx context.Context, userID, orgID uuid.UUID, roleStrings []string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	roles := make([]coreauth.Role, len(roleStrings))
	for i, r := range roleStrings {
		roles[i] = coreauth.Role(r)
	}

	accessToken, err = coreauth.IssueAccessToken(s.cfg.JWTSecret, s.cfg.AccessTokenTTL, userID, orgID, roles)
	if err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
	}

	refreshToken, jti, err := coreauth.IssueRefreshToken(s.cfg.JWTSecret, s.cfg.RefreshTokenTTL, userID)
	if err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
	}
	if err := s.rdb.Set(ctx, refreshKeyPrefix+jti, userID.String(), s.cfg.RefreshTokenTTL).Err(); err != nil {
		return "", "", 0, 0, httpx.ErrInternal()
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
