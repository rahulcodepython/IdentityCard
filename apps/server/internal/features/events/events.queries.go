package events

const (
	CreateEventQuery = `
        WITH ins AS (
            INSERT INTO events (organization_id, event_type, start_date, end_date)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	GetEventQuery = `
        SELECT jsonb_build_object(
            'id', e.id,
            'name', COALESCE(em.name, ''),
            'event_type', e.event_type,
            'status', e.status,
            'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
            'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
            'venue', em.venue,
            'organizer_name', em.organizer_name,
            'has_image', (em.image_object_key IS NOT NULL AND em.image_object_key <> ''),
            'has_organizer_signature', (em.organizer_signature_object_key IS NOT NULL AND em.organizer_signature_object_key <> ''),
            'published_at', to_char(e.published_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'days', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                    'date', to_char(d.date, 'YYYY-MM-DD'),
                    'entry_time', to_char(d.entry_time, 'HH24:MI'),
                    'exit_time', to_char(d.exit_time, 'HH24:MI')
                ) ORDER BY d.date)
                FROM event_days d WHERE d.event_id = e.id
            ), '[]'::jsonb)
        )
        FROM events e
        LEFT JOIN event_metadata em ON em.event_id = e.id
        WHERE e.id = $1 AND e.organization_id = $2;
    `

	GetRawEventQuery = `
        SELECT row_to_json(e) FROM events e WHERE e.id = $1 AND e.organization_id = $2;
    `

	ListEventsQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', e.id,
            'name', COALESCE(em.name, ''),
            'event_type', e.event_type,
            'status', e.status,
            'start_date', to_char(e.start_date, 'YYYY-MM-DD'),
            'end_date', to_char(e.end_date, 'YYYY-MM-DD'),
            'venue', em.venue,
            'organizer_name', em.organizer_name,
            'has_image', (em.image_object_key IS NOT NULL AND em.image_object_key <> ''),
            'has_organizer_signature', (em.organizer_signature_object_key IS NOT NULL AND em.organizer_signature_object_key <> ''),
            'published_at', to_char(e.published_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
        ) ORDER BY e.start_date DESC, e.created_at DESC), '[]'::jsonb)
        FROM events e
        LEFT JOIN event_metadata em ON em.event_id = e.id
        WHERE e.organization_id = $1;
    `

	UpdateEventDatesQuery = `
        WITH upd AS (
            UPDATE events
            SET start_date = $3, end_date = $4, updated_at = now()
            WHERE id = $1 AND organization_id = $2 AND status = 'draft'
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	PublishEventQuery = `
        WITH upd AS (
            UPDATE events
            SET status = 'published', published_at = now(), updated_at = now()
            WHERE id = $1 AND organization_id = $2 AND status = 'draft'
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	DeleteDraftEventQuery = `
        DELETE FROM events
        WHERE id = $1 AND organization_id = $2 AND status = 'draft';
    `

	DeleteEventQuery = `
        DELETE FROM events
        WHERE id = $1;
    `

	ListFlashEventsOlderThanQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(e)), '[]'::jsonb)
        FROM (
            SELECT e.* FROM events e
            JOIN event_days d ON d.event_id = e.id
            WHERE e.event_type = 'flash' AND d.date < $1::date
        ) e;
    `

	MarkUnjoinedPeopleJoinedAtQuery = `
        UPDATE people
        SET joined_at = $2, updated_at = now()
        WHERE event_id = $1 AND joined_at IS NULL;
    `

	CreateEventDayQuery = `
        WITH ins AS (
            INSERT INTO event_days (event_id, date, entry_time, exit_time)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	ListEventDaysQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'date', to_char(d.date, 'YYYY-MM-DD'),
            'entry_time', to_char(d.entry_time, 'HH24:MI'),
            'exit_time', to_char(d.exit_time, 'HH24:MI')
        ) ORDER BY d.date), '[]'::jsonb)
        FROM event_days d
        WHERE d.event_id = $1;
    `

	DeleteEventDaysForEventQuery = `
        DELETE FROM event_days
        WHERE event_id = $1;
    `

	CreateEventMetadataQuery = `
        WITH ins AS (
            INSERT INTO event_metadata (event_id, name, venue, organizer_name)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	GetEventMetadataQuery = `
        SELECT row_to_json(em)
        FROM event_metadata em
        WHERE em.event_id = $1;
    `

	UpdateEventMetadataQuery = `
        WITH upd AS (
            UPDATE event_metadata
            SET name = $2, venue = $3, organizer_name = $4, updated_at = now()
            WHERE event_id = $1
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	UpdateEventImageQuery = `
        WITH upd AS (
            UPDATE event_metadata
            SET image_object_key = $2, updated_at = now()
            WHERE event_id = $1
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	UpdateEventOrganizerSignatureQuery = `
        WITH upd AS (
            UPDATE event_metadata
            SET organizer_signature_object_key = $2, updated_at = now()
            WHERE event_id = $1
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `
)
