package eventsharing

const (
    // GetEventSharingQuery retrieves event_forms record, assigned form info, and total applicant count.
    GetEventSharingQuery = `
        SELECT jsonb_build_object(
            'event_form', (
                SELECT jsonb_build_object(
                    'id', ef.id,
                    'event_id', ef.event_id,
                    'form_id', ef.form_id,
                    'max_applicants', ef.max_applicants,
                    'expires_at', ef.expires_at,
                    'status', ef.status,
                    'created_at', ef.created_at,
                    'updated_at', ef.updated_at
                )
                FROM event_forms ef
                WHERE ef.event_id = $1
            ),
            'assigned_form', (
                SELECT jsonb_build_object(
                    'id', f.id,
                    'name', f.name,
                    'fields_count', COALESCE(jsonb_array_length(f.fields), 0),
                    'is_published', f.is_published
                )
                FROM event_forms ef
                JOIN forms f ON f.id = ef.form_id
                WHERE ef.event_id = $1
            ),
            'total_applicants', (
                SELECT count(*)::int
                FROM event_applicants ea
                WHERE ea.event_id = $1
            ),
            'can_change_form', (
                SELECT CASE WHEN count(*) = 0 THEN true ELSE false END
                FROM event_applicants ea
                WHERE ea.event_id = $1
            )
        );
    `

    // UpsertSharingCTEQuery validates form publication, existing submissions constraints,
    // upserts event_forms, and returns the response all in a single atomic CTE query.
    UpsertSharingCTEQuery = `
        WITH form_check AS (
            SELECT
                f.id,
                f.is_published
            FROM forms f
            WHERE f.id = $2
        ),
        existing_ef AS (
            SELECT
                ef.id,
                ef.form_id,
                (SELECT count(*)::int FROM event_applicants ea WHERE ea.event_id = $1) AS applicant_count
            FROM event_forms ef
            WHERE ef.event_id = $1
        ),
        validation AS (
            SELECT
                CASE
                    WHEN fc.id IS NULL THEN 'form_not_found'
                    WHEN NOT fc.is_published THEN 'form_not_published'
                    WHEN eef.applicant_count > 0 AND eef.form_id != $2 THEN 'cannot_change_form'
                    ELSE 'ok'
                END AS status_code
            FROM (SELECT 1) _
            LEFT JOIN form_check fc ON true
            LEFT JOIN existing_ef eef ON true
        ),
        upserted_ef AS (
            INSERT INTO event_forms (event_id, form_id, max_applicants, expires_at, status, updated_at)
            SELECT $1, $2, $3, $4, $5, now()
            FROM validation v
            WHERE v.status_code = 'ok'
            ON CONFLICT (event_id) DO UPDATE
            SET
                form_id = EXCLUDED.form_id,
                max_applicants = EXCLUDED.max_applicants,
                expires_at = EXCLUDED.expires_at,
                status = EXCLUDED.status,
                updated_at = now()
            RETURNING *
        )
        SELECT jsonb_build_object(
            'status_code', v.status_code,
            'response', CASE
                WHEN v.status_code = 'ok' THEN (
                    SELECT jsonb_build_object(
                        'event_form', jsonb_build_object(
                            'id', ef.id,
                            'event_id', ef.event_id,
                            'form_id', ef.form_id,
                            'max_applicants', ef.max_applicants,
                            'expires_at', ef.expires_at,
                            'status', ef.status,
                            'created_at', ef.created_at,
                            'updated_at', ef.updated_at
                        ),
                        'assigned_form', jsonb_build_object(
                            'id', f.id,
                            'name', f.name,
                            'fields_count', COALESCE(jsonb_array_length(f.fields), 0),
                            'is_published', f.is_published
                        ),
                        'total_applicants', (
                            SELECT count(*)::int
                            FROM event_applicants ea
                            WHERE ea.event_id = $1
                        ),
                        'can_change_form', (
                            SELECT CASE WHEN count(*) = 0 THEN true ELSE false END
                            FROM event_applicants ea
                            WHERE ea.event_id = $1
                        )
                    )
                    FROM upserted_ef ef
                    JOIN forms f ON f.id = ef.form_id
                )
                ELSE NULL
            END
        )
        FROM validation v;
    `
)
