-- name: CreateOrganizationMember :one
INSERT INTO organization_members (organization_id, user_id, roles)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrganizationMember :one
SELECT * FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: ListOrganizationMembershipsForUser :many
-- Every organization a user belongs to, with the org's name/slug attached
-- so a "pick your organization" login step doesn't need a second query.
SELECT
    organization_members.*,
    organizations.name AS organization_name,
    organizations.slug AS organization_slug
FROM organization_members
JOIN organizations ON organizations.id = organization_members.organization_id
WHERE organization_members.user_id = $1
ORDER BY organizations.name;
