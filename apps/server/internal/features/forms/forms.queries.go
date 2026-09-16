package forms

const (
    // ListFormsQuery retrieves paginated form templates in a single CTE round-trip.
    ListFormsQuery = `
        WITH filtered_templates AS (
            SELECT
                ft.id,
                ft.name,
                ft.fields,
                ft.created_at,
                ft.updated_at
            FROM form_templates ft
            WHERE ($1 = '' OR ft.name ILIKE '%' || $1 || '%')
        ),
        total_count AS (
            SELECT count(*) AS count FROM filtered_templates
        ),
        paginated_data AS (
            SELECT jsonb_build_object(
                'id', ft.id,
                'name', ft.name,
                'fields', ft.fields,
                'created_at', ft.created_at,
                'updated_at', ft.updated_at
            ) AS item
            FROM filtered_templates ft
            ORDER BY ft.created_at DESC
            LIMIT $2 OFFSET $3
        )
        SELECT jsonb_build_object(
            'data', COALESCE((SELECT jsonb_agg(item) FROM paginated_data), '[]'::jsonb),
            'total', COALESCE((SELECT count FROM total_count), 0),
            'page', $4::int,
            'limit', $2::int
        );
    `

    // GetFormQuery fetches a single form template by UUID.
    GetFormQuery = `
        SELECT jsonb_build_object(
            'id', ft.id,
            'name', ft.name,
            'fields', ft.fields,
            'created_at', ft.created_at,
            'updated_at', ft.updated_at
        )
        FROM form_templates ft
        WHERE ft.id = $1;
    `

    // CreateFormQuery creates a new form template with default seeded fields.
    CreateFormQuery = `
        INSERT INTO form_templates (name, fields)
        VALUES ($1, $2::jsonb)
        RETURNING jsonb_build_object(
            'id', id,
            'name', name,
            'fields', fields,
            'created_at', created_at,
            'updated_at', updated_at
        );
    `

    // UpdateFormQuery updates form template metadata (name).
    UpdateFormQuery = `
        UPDATE form_templates
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

    // UpdateFormFieldsQuery updates the entire fields array for a template.
    UpdateFormFieldsQuery = `
        UPDATE form_templates
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

    // DeleteFormQuery removes a form template.
    DeleteFormQuery = `
        DELETE FROM form_templates WHERE id = $1;
    `
)
