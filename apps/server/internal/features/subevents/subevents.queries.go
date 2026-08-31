package subevents

const (
	CreateSubEventQuery = `
        WITH ins AS (
            INSERT INTO sub_events (organization_id, event_id, name, date, entry_time, exit_time)
            VALUES ($1, $2, $3, $4, $5, $6)
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', ins.id,
            'name', ins.name,
            'date', to_char(ins.date, 'YYYY-MM-DD'),
            'entry_time', to_char(ins.entry_time, 'HH24:MI'),
            'exit_time', to_char(ins.exit_time, 'HH24:MI')
        )
        FROM ins;
    `

	GetSubEventQuery = `
        SELECT jsonb_build_object(
            'id', se.id,
            'name', se.name,
            'date', to_char(se.date, 'YYYY-MM-DD'),
            'entry_time', to_char(se.entry_time, 'HH24:MI'),
            'exit_time', to_char(se.exit_time, 'HH24:MI')
        )
        FROM sub_events se
        WHERE se.id = $1 AND se.event_id = $2 AND se.organization_id = $3;
    `

	ListSubEventsQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', se.id,
            'name', se.name,
            'date', to_char(se.date, 'YYYY-MM-DD'),
            'entry_time', to_char(se.entry_time, 'HH24:MI'),
            'exit_time', to_char(se.exit_time, 'HH24:MI')
        ) ORDER BY se.date), '[]'::jsonb)
        FROM sub_events se
        WHERE se.event_id = $1 AND se.organization_id = $2;
    `

	UpdateSubEventQuery = `
        WITH upd AS (
            UPDATE sub_events
            SET name = $4, date = $5, entry_time = $6, exit_time = $7, updated_at = now()
            WHERE id = $1 AND event_id = $2 AND organization_id = $3
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', u.id,
            'name', u.name,
            'date', to_char(u.date, 'YYYY-MM-DD'),
            'entry_time', to_char(u.entry_time, 'HH24:MI'),
            'exit_time', to_char(u.exit_time, 'HH24:MI')
        )
        FROM upd u;
    `

	DeleteSubEventQuery = `
        DELETE FROM sub_events
        WHERE id = $1 AND event_id = $2 AND organization_id = $3;
    `

	ValidateSubEventIDsQuery = `
        SELECT COALESCE(jsonb_agg(se.id), '[]'::jsonb)
        FROM sub_events se
        WHERE se.id = ANY($3::uuid[]) AND se.event_id = $1 AND se.organization_id = $2;
    `
)
