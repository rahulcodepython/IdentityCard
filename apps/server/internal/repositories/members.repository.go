package repositories

import (
	"context"

	"github.com/google/uuid"

	dbgen "identitycard-server/internal/db/sqlc/generated"
)

// MembersRepository owns an organization's roster: who belongs to an
// organization and which role(s) they hold there. It has no matching
// controller/service — it's consumed directly by AuthService, the only
// caller that needs it.
type MembersRepository struct {
	q *dbgen.Queries
}

func NewMembersRepository(q *dbgen.Queries) *MembersRepository {
	return &MembersRepository{q: q}
}

func (r *MembersRepository) Create(ctx context.Context, orgID, userID uuid.UUID, roles []string) (dbgen.OrganizationMember, error) {
	return r.q.CreateOrganizationMember(ctx, dbgen.CreateOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
		Roles:          roles,
	})
}

func (r *MembersRepository) Get(ctx context.Context, orgID, userID uuid.UUID) (dbgen.OrganizationMember, error) {
	return r.q.GetOrganizationMember(ctx, dbgen.GetOrganizationMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
}

// ListForUser returns every organization the user belongs to. Phase 0 has
// no "switch organization" UI yet, so callers use the first result as the
// user's active org (see AuthService.issueSessionForUser) — a multi-org
// member picking a different one is roadmap item 1.
func (r *MembersRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]dbgen.ListOrganizationMembershipsForUserRow, error) {
	return r.q.ListOrganizationMembershipsForUser(ctx, userID)
}

func (r *MembersRepository) ListOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]dbgen.ListOrganizationMembersRow, error) {
	return r.q.ListOrganizationMembers(ctx, orgID)
}

func (r *MembersRepository) Delete(ctx context.Context, orgID, memberID uuid.UUID) error {
	return r.q.DeleteOrganizationMember(ctx, dbgen.DeleteOrganizationMemberParams{
		OrganizationID: orgID,
		ID:             memberID,
	})
}
