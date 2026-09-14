package events

const (
	// ListEventsQuery retrieves paginated events conforming directly to generic.PaginatedResponse in a single CTE round-trip.
	ListEventsQuery = `
        WITH filtered_events AS (
            SELECT
                e.id,
                em.name,
                to_char(e.start_date, 'YYYY-MM-DD') AS start_date,
                to_char(e.end_date, 'YYYY-MM-DD') AS end_date,
                e.status,
                e.published_at,
                em.venue,
                em.logo,
                em.organizer,
                e.created_at,
                e.updated_at
            FROM events e
            JOIN event_metadata em ON em.id = e.id
            WHERE ($1 = '' OR em.name ILIKE '%' || $1 || '%' OR COALESCE(em.venue, '') ILIKE '%' || $1 || '%')
        ),
        total_count AS (
            SELECT count(*) AS count FROM filtered_events
        ),
        paginated_data AS (
            SELECT jsonb_build_object(
                'id', fe.id,
                'name', fe.name,
                'start_date', fe.start_date,
                'end_date', fe.end_date,
                'status', fe.status,
                'published_at', fe.published_at,
                'venue', fe.venue,
                'logo', fe.logo,
                'organizer', fe.organizer,
                'created_at', fe.created_at,
                'updated_at', fe.updated_at
            ) AS item
            FROM filtered_events fe
            ORDER BY fe.created_at DESC
            LIMIT $2 OFFSET $3
        )
        SELECT jsonb_build_object(
            'data', COALESCE((SELECT jsonb_agg(item) FROM paginated_data), '[]'::jsonb),
            'total', COALESCE((SELECT count FROM total_count), 0),
            'page', $4::int,
            'limit', $2::int
        );
    `

	// GetEventQuery fetches a single event by ID along with its metadata in a single JSON document.
	GetEventQuery = `
        SELECT jsonb_build_object(
            'id', e.id,
            'name', em.name,
            'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
            'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
            'status', e.status,
            'published_at', e.published_at,
            'venue', em.venue,
            'logo', em.logo,
            'organizer', em.organizer,
            'created_at', e.created_at,
            'updated_at', e.updated_at
        )
        FROM events e
        JOIN event_metadata em ON em.id = e.id
        WHERE e.id = $1;
    `

	// CreateEventQuery atomically inserts into events and event_metadata in a single CTE statement.
	CreateEventQuery = `
        WITH ins_event AS (
            INSERT INTO events (
                id,
                start_date,
                end_date,
                status,
                published_at,
                created_at,
                updated_at
            ) VALUES (
                gen_random_uuid(),
                $1::date,
                $2::date,
                $3,
                CASE WHEN $3 = 'published' THEN now() ELSE NULL END,
                now(),
                now()
            )
            RETURNING id, start_date, end_date, status, published_at, created_at, updated_at
        ),
        ins_meta AS (
            INSERT INTO event_metadata (
                id,
                name
            )
            SELECT
                id,
                $4
            FROM ins_event
            RETURNING id, name, venue, logo, organizer
        )
        SELECT jsonb_build_object(
            'id', e.id,
            'name', m.name,
            'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
            'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
            'status', e.status,
            'published_at', e.published_at,
            'venue', m.venue,
            'logo', m.logo,
            'organizer', m.organizer,
            'created_at', e.created_at,
            'updated_at', e.updated_at
        )
        FROM ins_event e
        JOIN ins_meta m ON m.id = e.id;
    `

	// UpdateEventQuery atomically updates events and event_metadata in a single CTE statement.
	UpdateEventQuery = `
        WITH upd_event AS (
            UPDATE events
            SET
                start_date = COALESCE($2::date, start_date),
                end_date = COALESCE($3::date, end_date),
                status = COALESCE($4, status),
                published_at = CASE
                    WHEN $4 = 'published' AND published_at IS NULL THEN now()
                    WHEN $4 = 'draft' THEN NULL
                    ELSE published_at
                END,
                updated_at = now()
            WHERE id = $1
            RETURNING id, start_date, end_date, status, published_at, created_at, updated_at
        ),
        upd_meta AS (
            UPDATE event_metadata
            SET
                name = COALESCE($5, name),
                venue = COALESCE($6, venue),
                logo = COALESCE($7, logo),
                organizer = COALESCE($8, organizer)
            WHERE id = $1
            RETURNING id, name, venue, logo, organizer
        )
        SELECT jsonb_build_object(
            'id', e.id,
            'name', m.name,
            'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
            'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
            'status', e.status,
            'published_at', e.published_at,
            'venue', m.venue,
            'logo', m.logo,
            'organizer', m.organizer,
            'created_at', e.created_at,
            'updated_at', e.updated_at
        )
        FROM upd_event e
        JOIN upd_meta m ON m.id = e.id;
    `

	// DeleteEventQuery deletes the event and cascades to event_metadata in a single CTE statement.
	DeleteEventQuery = `
        WITH del AS (
            DELETE FROM events
            WHERE id = $1
            RETURNING id
        )
        SELECT id FROM del;
    `
)
