package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

// Repository wraps the user-credential queries. Membership/role lookups
// live in internal/modules/members since they're a separate concern
// (an org's roster) reused outside login too.
type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) CreateUser(ctx context.Context, email, name, totpSecret string) (dbgen.User, error) {
	return r.q.CreateUser(ctx, dbgen.CreateUserParams{
		Email:      email,
		Name:       name,
		TotpSecret: pgtype.Text{String: totpSecret, Valid: true},
	})
}

// CreateOAuthUser creates a Google-only account: no totp_secret, and
// already email-verified since Google verified it.
func (r *Repository) CreateOAuthUser(ctx context.Context, email, name, googleID, avatarURL string) (dbgen.User, error) {
	return r.q.CreateOAuthUser(ctx, dbgen.CreateOAuthUserParams{
		Email:     email,
		Name:      name,
		GoogleID:  pgtype.Text{String: googleID, Valid: true},
		AvatarUrl: pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	})
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (dbgen.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (dbgen.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *Repository) GetUserByGoogleID(ctx context.Context, googleID string) (dbgen.User, error) {
	return r.q.GetUserByGoogleID(ctx, pgtype.Text{String: googleID, Valid: true})
}

// LinkGoogleID sets google_id on an existing account the first time they
// authenticate via Google with the same (Google-verified) email.
func (r *Repository) LinkGoogleID(ctx context.Context, userID uuid.UUID, googleID string) (dbgen.User, error) {
	return r.q.LinkGoogleID(ctx, dbgen.LinkGoogleIDParams{
		ID:       userID,
		GoogleID: pgtype.Text{String: googleID, Valid: true},
	})
}

func (r *Repository) SetOTP(ctx context.Context, userID uuid.UUID, code string, expiresAt time.Time) (dbgen.User, error) {
	return r.q.SetUserOTP(ctx, dbgen.SetUserOTPParams{
		ID:           userID,
		OtpCode:      pgtype.Text{String: code, Valid: true},
		OtpExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

// VerifyAndConsumeOTP matches and clears the code in one statement, so it
// can never be replayed — see the query's WHERE clause.
func (r *Repository) VerifyAndConsumeOTP(ctx context.Context, email, code string) (dbgen.User, error) {
	return r.q.VerifyAndConsumeOTP(ctx, dbgen.VerifyAndConsumeOTPParams{
		Email:   email,
		OtpCode: pgtype.Text{String: code, Valid: true},
	})
}

func (r *Repository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) (dbgen.User, error) {
	return r.q.MarkEmailVerified(ctx, userID)
}
