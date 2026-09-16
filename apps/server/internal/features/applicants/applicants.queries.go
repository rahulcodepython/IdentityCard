package applicants

import "fmt"

const (
    // GetEventAssignedFormQuery retrieves assigned form schema for an event.
    GetEventAssignedFormQuery = `
        SELECT jsonb_build_object(
            'id', f.id,
            'name', f.name,
            'fields', f.fields
        )
        FROM event_forms ef
        JOIN forms f ON f.id = ef.form_id
        WHERE ef.event_id = $1::uuid;
    `
)

// BuildFilteredApplicantsQuery constructs the dynamic CTE query for filtering and paginating event applicants.
func BuildFilteredApplicantsQuery(whereClause string, limitIdx, offsetIdx, pageIdx, limitIdx2 int) string {
    return fmt.Sprintf(`
        WITH filtered_applicants AS (
            SELECT
                a.id AS user_id,
                a.name,
                a.email,
                a.data,
                ea.created_at
            FROM event_applicants ea
            JOIN applicants a ON a.id = ea.user_id
            WHERE %s
        ),
        total_count AS (
            SELECT count(*) AS count FROM filtered_applicants
        ),
        paginated_data AS (
            SELECT jsonb_build_object(
                'user_id', fa.user_id,
                'name', fa.name,
                'email', fa.email,
                'data', fa.data,
                'created_at', fa.created_at
            ) AS item
            FROM filtered_applicants fa
            ORDER BY fa.created_at DESC
            LIMIT $%d OFFSET $%d
        )
        SELECT jsonb_build_object(
            'data', COALESCE((SELECT jsonb_agg(item) FROM paginated_data), '[]'::jsonb),
            'total', COALESCE((SELECT count FROM total_count), 0),
            'page', $%d::int,
            'limit', $%d::int
        );
    `, whereClause, limitIdx, offsetIdx, pageIdx, limitIdx2)
}
