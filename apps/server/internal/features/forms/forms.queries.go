package forms

const (
    // ListFormsQuery retrieves paginated forms in a single CTE round-trip.
    ListFormsQuery = `
        WITH filtered_forms AS (
            SELECT
                f.id,
                f.name,
                f.fields,
                f.created_at,
                f.updated_at
            FROM forms f
            WHERE ($1 = '' OR f.name ILIKE '%' || $1 || '%')
        ),
        total_count AS (
            SELECT count(*) AS count FROM filtered_forms
        ),
        paginated_data AS (
            SELECT jsonb_build_object(
                'id', ff.id,
                'name', ff.name,
                'fields', ff.fields,
                'created_at', ff.created_at,
                'updated_at', ff.updated_at
            ) AS item
            FROM filtered_forms ff
            ORDER BY ff.created_at DESC
            LIMIT $2 OFFSET $3
        )
        SELECT jsonb_build_object(
            'data', COALESCE((SELECT jsonb_agg(item) FROM paginated_data), '[]'::jsonb),
            'total', COALESCE((SELECT count FROM total_count), 0),
            'page', $4::int,
            'limit', $2::int
        );
    `

    // GetFormQuery fetches a single form by UUID.
    GetFormQuery = `
        SELECT jsonb_build_object(
            'id', f.id,
            'name', f.name,
            'fields', f.fields,
            'created_at', f.created_at,
            'updated_at', f.updated_at
        )
        FROM forms f
        WHERE f.id = $1;
    `

    // CreateFormQuery creates a new form with default seeded fields.
    CreateFormQuery = `
        INSERT INTO forms (name, fields)
        VALUES ($1, $2::jsonb)
        RETURNING jsonb_build_object(
            'id', id,
            'name', name,
            'fields', fields,
            'created_at', created_at,
            'updated_at', updated_at
        );
    `

    // UpdateFormQuery updates form metadata (name).
    UpdateFormQuery = `
        UPDATE forms
        SET
            name = COALESCE($2, name),
            updated_at = now()
        WHERE id = $1
        RETURNING jsonb_build_object(
            'id', id,
            'name', name,
            'fields', fields,
            'created_at', created_at,
            'updated_at', updated_at
        );
    `

    // UpdateFormFieldsQuery atomic update for the entire fields array.
    UpdateFormFieldsQuery = `
        UPDATE forms
        SET
            fields = $2::jsonb,
            updated_at = now()
        WHERE id = $1
        RETURNING jsonb_build_object(
            'id', id,
            'name', name,
            'fields', fields,
            'created_at', created_at,
            'updated_at', updated_at
        );
    `

    // DeleteFormQuery removes a form.
    DeleteFormQuery = `
        DELETE FROM forms WHERE id = $1;
    `
)
