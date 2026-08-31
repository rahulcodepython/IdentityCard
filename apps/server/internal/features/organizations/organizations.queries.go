package organizations

const (
	GetOrganizationByIDQuery = `
        SELECT jsonb_build_object(
            'id', o.id,
            'name', o.name,
            'slug', o.slug,
            'logo_object_key', o."logoObjectKey",
            'has_logo', (o."logoObjectKey" IS NOT NULL AND o."logoObjectKey" <> '')
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
            'slug', u.slug,
            'logo_object_key', u."logoObjectKey",
            'has_logo', (u."logoObjectKey" IS NOT NULL AND u."logoObjectKey" <> '')
        )
        FROM updated u;
    `

	UpdateOrganizationLogoQuery = `
        WITH updated AS (
            UPDATE "organization"
            SET "logoObjectKey" = $2
            WHERE id = $1
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', u.id,
            'name', u.name,
            'slug', u.slug,
            'logo_object_key', u."logoObjectKey",
            'has_logo', (u."logoObjectKey" IS NOT NULL AND u."logoObjectKey" <> '')
        )
        FROM updated u;
    `

	DeleteOrganizationLogoQuery = `
        UPDATE "organization"
        SET "logoObjectKey" = NULL
        WHERE id = $1;
    `

	DeleteOrganizationQuery = `
        DELETE FROM "organization"
        WHERE id = $1;
    `
)
