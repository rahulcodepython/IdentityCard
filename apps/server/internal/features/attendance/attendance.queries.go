package attendance

const (
	GetAttendanceRecordQuery = `
        SELECT row_to_json(a)
        FROM attendance_records a
        WHERE a.person_id = $1 AND a.date = $2;
    `

	CreateAttendanceEntryQuery = `
        WITH ins AS (
            INSERT INTO attendance_records (organization_id, event_id, person_id, date, entry_at, entry_status, entry_device_id)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	RecordAttendanceExitQuery = `
        WITH upd AS (
            UPDATE attendance_records
            SET exit_at = $3, exit_status = $4, exit_device_id = $5, updated_at = now()
            WHERE person_id = $1 AND date = $2
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	ListAttendanceForEventQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', ar.id,
            'organization_id', ar.organization_id,
            'event_id', ar.event_id,
            'person_id', ar.person_id,
            'date', to_char(ar.date, 'YYYY-MM-DD'),
            'entry_at', to_char(ar.entry_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'entry_status', ar.entry_status,
            'entry_device_id', ar.entry_device_id,
            'exit_at', to_char(ar.exit_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'exit_status', ar.exit_status,
            'exit_device_id', ar.exit_device_id,
            'person_name', p.name
        ) ORDER BY ar.date, p.name), '[]'::jsonb)
        FROM attendance_records ar
        JOIN people p ON p.id = ar.person_id
        WHERE ar.event_id = $1 AND ar.organization_id = $2;
    `
)
