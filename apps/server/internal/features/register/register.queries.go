package register

const (
    // GetPublicApplyQuery loads event_forms by id, joins event metadata and form fields.
    // If form is not locked, query returns no rows (404 no such form exists).
    // If expired or full, form schema is returned as NULL.
    GetPublicApplyQuery = `
        SELECT jsonb_build_object(
            'event_form_id', ef.id,
            'status', ef.status,
            'is_expired', (ef.expires_at < now()),
            'is_full', (ef.max_applicants != -1 AND (SELECT count(*) FROM event_applicants WHERE event_id = ef.event_id) >= ef.max_applicants),
            'max_applicants', ef.max_applicants,
            'current_applicants', (SELECT count(*)::int FROM event_applicants WHERE event_id = ef.event_id),
            'expires_at', ef.expires_at,
            'event', jsonb_build_object(
                'id', e.id,
                'name', COALESCE(em.name, ''),
                'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
                'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
                'venue', em.venue,
                'logo', em.logo,
                'organizer', em.organizer
            ),
            'form', CASE 
                WHEN NOT ef.is_locked 
                  OR ef.expires_at < now() 
                  OR (ef.max_applicants != -1 AND (SELECT count(*) FROM event_applicants WHERE event_id = ef.event_id) >= ef.max_applicants)
                THEN NULL
                ELSE jsonb_build_object(
                    'id', ef.id,
                    'name', ef.name,
                    'fields', ef.fields
                )
            END
        )
        FROM event_forms ef
        JOIN events e ON e.id = ef.event_id
        LEFT JOIN event_metadata em ON em.id = e.id
        WHERE ef.id = $1
          AND ef.is_locked = true;
    `

    // SubmitApplicationCTEQuery performs form validation, capacity checks, duplicate checks,
    // and both applicant + event_applicant inserts atomically in one single database roundtrip.
    SubmitApplicationCTEQuery = `
        WITH form_lookup AS (
            SELECT
                ef.id,
                ef.event_id,
                ef.status,
                ef.is_locked,
                ef.expires_at,
                ef.max_applicants,
                (SELECT count(*)::int FROM event_applicants WHERE event_id = ef.event_id) AS current_count,
                EXISTS (SELECT 1 FROM event_applicants WHERE event_id = ef.event_id AND email = $4) AS email_exists
            FROM event_forms ef
            WHERE ef.id = $1
        ),
        validation_status AS (
            SELECT
                fl.id,
                fl.event_id,
                CASE
                    WHEN fl.id IS NULL OR NOT fl.is_locked THEN 'not_found'
                    WHEN fl.expires_at < now() THEN 'expired'
                    WHEN fl.max_applicants != -1 AND fl.current_count >= fl.max_applicants THEN 'limit_reached'
                    WHEN fl.email_exists THEN 'already_registered'
                    ELSE 'ok'
                END AS status_code
            FROM (SELECT 1) _
            LEFT JOIN form_lookup fl ON true
        ),
        inserted_applicant AS (
            INSERT INTO applicants (id, name, email, data, created_at, updated_at)
            SELECT $2, $3, $4, $5::jsonb, now(), now()
            FROM validation_status vs
            WHERE vs.status_code = 'ok'
            ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
            RETURNING id
        ),
        inserted_event_applicant AS (
            INSERT INTO event_applicants (event_id, user_id, email, created_at)
            SELECT vs.event_id, ia.id, $4, now()
            FROM validation_status vs
            JOIN inserted_applicant ia ON true
            WHERE vs.status_code = 'ok'
            RETURNING id
        )
        SELECT
            vs.status_code,
            COALESCE(vs.event_id::text, '') AS event_id
        FROM validation_status vs;
    `
)
