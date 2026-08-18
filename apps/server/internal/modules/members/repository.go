// Package members owns an organization's roster: who belongs to an
// organization and which role(s) they hold there.
package members

import (
	"context"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

type Repository struct {
	q *dbgen.Queries
}

func NewRepository(q *dbgen.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) Create(ctx context.Context, orgID, userID uuid.UUID, roles []string) (dbgen.OrganizationMember, error) {
	return r.q.CreateOrganizationMember(ctx, dbgen.CreateOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
		Roles:          roles,
	})
}

func (r *Repository) Get(ctx context.Context, orgID, userID uuid.UUID) (dbgen.OrganizationMember, error) {
	return r.q.GetOrganizationMember(ctx, dbgen.GetOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
}

// ListForUser returns every organization the user belongs to. Phase 0 has
// no "switch organization" UI yet, so callers use the first result as the
// user's active org (see internal/modules/auth Service.Login) — a
// multi-org member picking a different one is roadmap item 1.
func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]dbgen.ListOrganizationMembershipsForUserRow, error) {
	return r.q.ListOrganizationMembershipsForUser(ctx, userID)
}
