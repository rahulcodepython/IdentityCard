package eventdates

const (
    // ListEventDatesByMonthQuery retrieves all scheduled dates for an event in a specified month.
    ListEventDatesByMonthQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', ed.id,
            'event_id', ed.event_id,
            'date', to_char(ed.date, 'YYYY-MM-DD'),
            'start_time', to_char(ed.start_time, 'HH24:MI'),
            'end_time', to_char(ed.end_time, 'HH24:MI'),
            'created_at', ed.created_at,
            'updated_at', ed.updated_at
        ) ORDER BY ed.date ASC), '[]'::jsonb)
        FROM event_dates ed
        WHERE ed.event_id = $1
            AND ($2 = '' OR to_char(ed.date, 'YYYY-MM') = $2);
    `

    // BulkUpsertEventDatesQuery atomically checks event existence and validates that all dates
    // fall within the event start_date and end_date range before upserting.
    BulkUpsertEventDatesQuery = `
        WITH target_event AS (
            SELECT id, start_date, end_date
            FROM events
            WHERE id = $1
        ),
        incoming_dates AS (
            SELECT
                (d->>'date')::date AS date,
                (d->>'start_time')::time AS start_time,
                (d->>'end_time')::time AS end_time
            FROM jsonb_array_elements($2::jsonb) AS d
        ),
        validation AS (
            SELECT
                CASE
                    WHEN NOT EXISTS (SELECT 1 FROM target_event) THEN 'not_found'
                    WHEN EXISTS (
                        SELECT 1 FROM incoming_dates ind
                        CROSS JOIN target_event te
                        WHERE ind.date < te.start_date OR ind.date > te.end_date
                    ) THEN 'date_out_of_range'
                    ELSE 'valid'
                END AS status
        ),
        upsert_dates AS (
            INSERT INTO event_dates (id, event_id, date, start_time, end_time, created_at, updated_at)
            SELECT
                gen_random_uuid(),
                $1,
                ind.date,
                ind.start_time,
                ind.end_time,
                now(),
                now()
            FROM incoming_dates ind
            WHERE (SELECT status FROM validation) = 'valid'
            ON CONFLICT (event_id, date) DO UPDATE
            SET
                start_time = EXCLUDED.start_time,
                end_time = EXCLUDED.end_time,
                updated_at = now()
            RETURNING id, event_id, date, start_time, end_time, created_at, updated_at
        )
        SELECT jsonb_build_object(
            'status', (SELECT status FROM validation),
            'data', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                    'id', u.id,
                    'event_id', u.event_id,
                    'date', to_char(u.date, 'YYYY-MM-DD'),
                    'start_time', to_char(u.start_time, 'HH24:MI'),
                    'end_time', to_char(u.end_time, 'HH24:MI'),
                    'created_at', u.created_at,
                    'updated_at', u.updated_at
                ) ORDER BY u.date ASC) FROM upsert_dates u
            ), '[]'::jsonb)
        );
    `

    // BulkDeleteEventDatesQuery removes multiple dates for an event in one statement.
    BulkDeleteEventDatesQuery = `
        DELETE FROM event_dates
        WHERE event_id = $1
            AND date = ANY($2::date[]);
    `

    // ValidateEventDatesQuery validates that an event exists and incoming dates are within [start_date, end_date].
    ValidateEventDatesQuery = `
        WITH target_event AS (
            SELECT id, start_date, end_date
            FROM events
            WHERE id = $1
        ),
        incoming_dates AS (
            SELECT
                (d->>'date')::date AS date
            FROM jsonb_array_elements($2::jsonb) AS d
        )
        SELECT
            CASE
                WHEN NOT EXISTS (SELECT 1 FROM target_event) THEN 'not_found'
                WHEN EXISTS (
                    SELECT 1 FROM incoming_dates ind
                    CROSS JOIN target_event te
                    WHERE ind.date < te.start_date OR ind.date > te.end_date
                ) THEN 'date_out_of_range'
                ELSE 'valid'
            END AS status;
    `

    // DeleteAllEventDatesQuery removes all date records for a specific event.
    DeleteAllEventDatesQuery = `
        DELETE FROM event_dates
        WHERE event_id = $1;
    `

    // InsertEventDatesQuery inserts a batch of event dates and returns the newly inserted rows as JSON.
    InsertEventDatesQuery = `
        WITH incoming_dates AS (
            SELECT
                (d->>'date')::date AS date,
                (d->>'start_time')::time AS start_time,
                (d->>'end_time')::time AS end_time
            FROM jsonb_array_elements($2::jsonb) AS d
        ),
        inserted AS (
            INSERT INTO event_dates (id, event_id, date, start_time, end_time, created_at, updated_at)
            SELECT
                gen_random_uuid(),
                $1,
                ind.date,
                ind.start_time,
                ind.end_time,
                now(),
                now()
            FROM incoming_dates ind
            RETURNING id, event_id, date, start_time, end_time, created_at, updated_at
        )
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', u.id,
            'event_id', u.event_id,
            'date', to_char(u.date, 'YYYY-MM-DD'),
            'start_time', to_char(u.start_time, 'HH24:MI'),
            'end_time', to_char(u.end_time, 'HH24:MI'),
            'created_at', u.created_at,
            'updated_at', u.updated_at
        ) ORDER BY u.date ASC), '[]'::jsonb)
        FROM inserted u;
    `
)
