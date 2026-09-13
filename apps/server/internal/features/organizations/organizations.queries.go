package organizations

const (
    ListOrganizationsQuery = `
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', o.id,
                'name', o.name,
                'slug', o.slug,
                'logo', o.logo,
                'role', m.role
            )
        )
        FROM "organization" o
        JOIN "member" m ON m."organizationId" = o.id
        WHERE m."userId" = $1;
    `

    GetOrganizationByIDQuery = `
        SELECT jsonb_build_object(
            'id', o.id,
            'name', o.name,
            'slug', o.slug
        )
        FROM "organization" o
        WHERE o.id = $1;
    `

    UpdateOrganizationNameQuery = `
        WITH updated AS (
            UPDATE "organization"
            SET name = $2
            WHERE id = $1
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', u.id,
            'name', u.name,
            'slug', u.slug
        )
        FROM updated u;
    `

    DeleteOrganizationQuery = `
        WITH user_org_count AS (
            SELECT COUNT(*)::int AS total
            FROM "member"
            WHERE "userId" = $2
        ),
        deleted AS (
            DELETE FROM "organization"
            WHERE id = $1
              AND (SELECT total FROM user_org_count) > 1
        )
        SELECT jsonb_build_object(
            'user_org_count', (SELECT total FROM user_org_count),
            'deleted', EXISTS(SELECT 1 FROM deleted)
        );
    `
)
