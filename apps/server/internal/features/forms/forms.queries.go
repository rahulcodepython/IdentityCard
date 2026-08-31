package forms

const (
	CreateEventFormQuery = `
        WITH ins AS (
            INSERT INTO event_forms (organization_id, event_id, sub_event_id, token, capacity)
            VALUES ($1, $2, $3, $4, $5)
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', ins.id,
            'token', ins.token,
            'sub_event_id', ins.sub_event_id,
            'capacity', ins.capacity,
            'submissions_count', ins.submissions_count,
            'is_active', ins.is_active
        ) FROM ins;
    `

	ListEventFormsQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', f.id,
            'token', f.token,
            'sub_event_id', f.sub_event_id,
            'capacity', f.capacity,
            'submissions_count', f.submissions_count,
            'is_active', f.is_active
        ) ORDER BY f.created_at DESC), '[]'::jsonb)
        FROM event_forms f
        WHERE f.event_id = $1 AND f.organization_id = $2;
    `

	GetEventFormQuery = `
        SELECT jsonb_build_object(
            'id', f.id,
            'token', f.token,
            'sub_event_id', f.sub_event_id,
            'capacity', f.capacity,
            'submissions_count', f.submissions_count,
            'is_active', f.is_active
        )
        FROM event_forms f
        WHERE f.id = $1 AND f.event_id = $2 AND f.organization_id = $3;
    `

	UpdateEventFormQuery = `
        WITH upd AS (
            UPDATE event_forms
            SET capacity = $4, is_active = $5, updated_at = now()
            WHERE id = $1 AND event_id = $2 AND organization_id = $3
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', upd.id,
            'token', upd.token,
            'sub_event_id', upd.sub_event_id,
            'capacity', upd.capacity,
            'submissions_count', upd.submissions_count,
            'is_active', upd.is_active
        ) FROM upd;
    `

	DeleteEventFormQuery = `
        DELETE FROM event_forms
        WHERE id = $1 AND event_id = $2 AND organization_id = $3;
    `

	GetPublicFormByTokenQuery = `
        SELECT jsonb_build_object(
            'id', ef.id,
            'token', ef.token,
            'capacity', ef.capacity,
            'submissions_count', ef.submissions_count,
            'is_active', ef.is_active,
            'event_id', ef.event_id,
            'sub_event_id', ef.sub_event_id,
            'organization_id', ef.organization_id,
            'event_name', em.name,
            'sub_event_name', se.name
        )
        FROM event_forms ef
        JOIN events e ON e.id = ef.event_id
        LEFT JOIN event_metadata em ON em.event_id = e.id
        LEFT JOIN sub_events se ON se.id = ef.sub_event_id
        WHERE ef.token = $1;
    `

	IncrementEventFormSubmissionsQuery = `
        WITH upd AS (
            UPDATE event_forms
            SET submissions_count = submissions_count + 1, updated_at = now()
            WHERE id = $1 AND is_active = true AND (capacity IS NULL OR submissions_count < capacity)
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', upd.id,
            'token', upd.token,
            'sub_event_id', upd.sub_event_id,
            'capacity', upd.capacity,
            'submissions_count', upd.submissions_count,
            'is_active', upd.is_active
        ) FROM upd;
    `

	DecrementEventFormSubmissionsQuery = `
        UPDATE event_forms
        SET submissions_count = GREATEST(submissions_count - 1, 0), updated_at = now()
        WHERE id = $1;
    `
)
