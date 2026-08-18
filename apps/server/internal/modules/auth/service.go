package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	coreauth "identitycard-server/internal/auth"
	"identitycard-server/internal/config"
	"identitycard-server/internal/db"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/httpx"
	"identitycard-server/internal/modules/members"
	"identitycard-server/internal/modules/organizations"
	"identitycard-server/internal/modules/plans"
)

const refreshKeyPrefix = "refresh_token:"

type Service struct {
	cfg     *config.Config
	users   *Repository
	members *members.Repository
	orgs    *organizations.Service
	rdb     *redis.Client

	// Used only by Register, which writes across four tables (users,
	// organizations, subscriptions, organization_members) atomically —
	// everything else in this service works through the repositories above.
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
) *Service {
	return &Service{
		cfg: cfg, users: users, members: members, orgs: orgs,
		pool: pool, baseQueries: baseQueries, rdb: rdb,
	}
}

var errInvalidCredentials = httpx.NewError(http.StatusUnauthorized, "invalid_credentials", "invalid email or password")

// Login verifies credentials and issues a token pair scoped to the user's
// first organization membership. Phase 0 has no "choose your organization"
// step yet (see members.Repository.ListForUser) — a user who belongs to
// more than one org always lands in the first, alphabetically by org name.
func (s *Service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", 0, 0, errInvalidCredentials
	}
	if !coreauth.VerifyPassword(user.PasswordHash, password) {
		return "", "", 0, 0, errInvalidCredentials
	}

	memberships, err := s.members.ListForUser(ctx, user.ID)
	if err != nil || len(memberships) == 0 {
		return "", "", 0, 0, httpx.ErrForbidden("this account does not belong to an organization yet")
	}
	membership := memberships[0]

	return s.issueTokenPair(ctx, user.ID, membership.OrganizationID, membership.Roles)
}

// Register creates a brand-new organization end to end: the user, the
// organization (with a generated unique slug), a subscription to the
// chosen plan (stubbed 'active' — no real payment call yet, see
// plans.Repository.CreateSubscription), and a super_admin membership
// linking them — all in one transaction, then logs the user in.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (accessToken, refreshToken string, accessTTL, refreshTTL time.Duration, err error) {
	var userID, orgID uuid.UUID

	txErr := db.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.baseQueries.WithTx(tx)

		plan, txErr := plans.NewRepository(q).GetByCode(ctx, req.PlanCode)
		if txErr != nil {
			return httpx.NewError(http.StatusBadRequest, "invalid_plan", "unknown plan")
		}

		orgRepo := organizations.NewRepository(q)
		slug, txErr := organizations.GenerateUniqueSlug(ctx, orgRepo, req.OrganizationName)
		if txErr != nil {
			return httpx.ErrInternal()
		}
		org, txErr := orgRepo.Create(ctx, req.OrganizationName, slug)
		if txErr != nil {
			return httpx.ErrInternal()
		}

		hash, txErr := coreauth.HashPassword(req.Password)
		if txErr != nil {
			return httpx.ErrInternal()
		}
		user, txErr := NewRepository(q).CreateUser(ctx, req.Email, hash, req.Name)
		if txErr != nil {
			if db.IsUniqueViolation(txErr, "users_email_key") {
				return httpx.NewError(http.StatusConflict, "email_taken", "an account with this email already exists")
			}
			return httpx.ErrInternal()
		}

		if _, txErr := plans.NewRepository(q).CreateSubscription(ctx, org.ID, plan.ID); txErr != nil {
			return httpx.ErrInternal()
		}
		if _, txErr := members.NewRepository(q).Create(ctx, org.ID, user.ID, []string{string(coreauth.RoleSuperAdmin)}); txErr != nil {
			return httpx.ErrInternal()
		}

		userID, orgID = user.ID, org.ID
		return nil
	})
	if txErr != nil {
		return "", "", 0, 0, txErr
	}

	return s.issueTokenPair(ctx, userID, orgID, []string{string(coreauth.RoleSuperAdmin)})
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
	if err != nil || len(memberships) == 0 {
		return "", "", 0, 0, httpx.ErrForbidden("this account does not belong to an organization yet")
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
	membership, err := s.members.Get(ctx, claims.OrganizationID, claims.UserID)
	if err != nil {
		return MeResponse{}, httpx.ErrNotFound("organization membership")
	}
	org, err := s.orgs.GetSettings(ctx, claims.OrganizationID)
	if err != nil {
		return MeResponse{}, err
	}

	return MeResponse{
		UserID:           user.ID,
		Email:            user.Email,
		Name:             user.Name,
		OrganizationID:   claims.OrganizationID,
		OrganizationName: org.Name,
		Roles:            membership.Roles,
	}, nil
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
