package eventform

const (
    // GetEventFormByEventIDQuery retrieves the event-specific form details.
    GetEventFormByEventIDQuery = `
        SELECT jsonb_build_object(
            'id', ef.id,
            'event_id', ef.event_id,
            'name', ef.name,
            'fields', ef.fields,
            'is_locked', ef.is_locked,
            'locked_at', ef.locked_at,
            'max_applicants', ef.max_applicants,
            'expires_at', ef.expires_at,
            'total_applicants', (SELECT count(*)::int FROM event_applicants WHERE event_id = ef.event_id),
            'can_delete', ((SELECT count(*) FROM event_applicants WHERE event_id = ef.event_id) = 0),
            'created_at', ef.created_at,
            'updated_at', ef.updated_at
        )
        FROM event_forms ef
        WHERE ef.event_id = $1::uuid;
    `

    // CreateEventFormFromTemplateQuery clones fields from an existing template.
    CreateEventFormFromTemplateQuery = `
        INSERT INTO event_forms (
            event_id,
            name,
            fields,
            max_applicants,
            expires_at,
            is_locked
        )
        SELECT
            $1::uuid,
            COALESCE(NULLIF($3, ''), ft.name),
            ft.fields,
            $4,
            $5,
            false
        FROM form_templates ft
        WHERE ft.id = $2::uuid
        RETURNING id;
    `

    // CreateEventFormFromScratchQuery creates a brand new form instance.
    CreateEventFormFromScratchQuery = `
        INSERT INTO event_forms (
            event_id,
            name,
            fields,
            max_applicants,
            expires_at,
            is_locked
        )
        VALUES (
            $1::uuid,
            $2,
            $3::jsonb,
            $4,
            $5,
            false
        )
        RETURNING id;
    `

    // UpdateEventFormQuery updates event form fields and limits while unlocked.
    UpdateEventFormQuery = `
        UPDATE event_forms
        SET
            name = COALESCE($2, name),
            fields = COALESCE($3::jsonb, fields),
            max_applicants = COALESCE($4, max_applicants),
            expires_at = COALESCE($5, expires_at),
            updated_at = now()
        WHERE event_id = $1::uuid
          AND is_locked = false
        RETURNING id;
    `

    // LockEventFormQuery permanently locks the event form and marks status live.
    LockEventFormQuery = `
        UPDATE event_forms
        SET
            is_locked = true,
            locked_at = now(),
            status = 'live',
            updated_at = now()
        WHERE event_id = $1::uuid
          AND is_locked = false
        RETURNING id;
    `

    // DeleteEventFormQuery deletes the event form only if no applicants exist.
    DeleteEventFormQuery = `
        DELETE FROM event_forms
        WHERE event_id = $1::uuid
          AND NOT EXISTS (
              SELECT 1 FROM event_applicants WHERE event_id = $1::uuid
          )
        RETURNING id;
    `

    // CountApplicantsForEventQuery counts applicants registered for the event.
    CountApplicantsForEventQuery = `
        SELECT count(*)::int FROM event_applicants WHERE event_id = $1::uuid;
    `

    // CheckEventFormLockStatusQuery checks if an event form exists and is locked.
    CheckEventFormLockStatusQuery = `
        SELECT is_locked FROM event_forms WHERE event_id = $1::uuid;
    `
)
