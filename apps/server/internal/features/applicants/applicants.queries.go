package applicants

import "fmt"

const (
    // GetEventAssignedFormQuery retrieves assigned form schema for an event.
    GetEventAssignedFormQuery = `
        SELECT jsonb_build_object(
            'id', ef.id,
            'name', ef.name,
            'fields', ef.fields
        )
        FROM event_forms ef
        WHERE ef.event_id = $1::uuid;
    `

    // CreateApplicantAtomicQuery checks for event existence, validates duplicate email,
    // and registers the applicant atomically in one single database roundtrip.
    CreateApplicantAtomicQuery = `
        WITH event_check AS (
            SELECT id FROM events WHERE id = $1::uuid
        ),
        duplicate_check AS (
            SELECT EXISTS (
                SELECT 1 FROM event_applicants WHERE event_id = $1::uuid AND email = $4
            ) AS email_exists
        ),
        inserted_applicant AS (
            INSERT INTO applicants (id, name, email, phone, data, created_at, updated_at)
            SELECT $2, $3, $4, $5, $6::jsonb, now(), now()
            FROM event_check ec, duplicate_check dc
            WHERE dc.email_exists = false
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email, phone = EXCLUDED.phone, data = EXCLUDED.data, updated_at = now()
            RETURNING id, name, email, phone, data, created_at
        ),
        inserted_event_applicant AS (
            INSERT INTO event_applicants (event_id, user_id, email, created_at)
            SELECT $1::uuid, ia.id, ia.email, now()
            FROM inserted_applicant ia
            RETURNING id
        )
        SELECT jsonb_build_object(
            'user_id', ia.id,
            'name', ia.name,
            'email', ia.email,
            'phone', ia.phone,
            'data', ia.data,
            'created_at', ia.created_at
        )
        FROM inserted_applicant ia;
    `

    DeleteApplicantQuery = `
        WITH deleted_association AS (
            DELETE FROM event_applicants
            WHERE event_id = $1::uuid AND user_id = $2
            RETURNING user_id
        ),
        deleted_attendance AS (
            DELETE FROM event_attendance
            WHERE event_id = $1::uuid AND applicant_id = $2
        )
        DELETE FROM applicants
        WHERE id = $2
          AND EXISTS (SELECT 1 FROM deleted_association)
          AND NOT EXISTS (SELECT 1 FROM event_applicants WHERE user_id = $2);
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
                a.phone,
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
                'phone', fa.phone,
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
