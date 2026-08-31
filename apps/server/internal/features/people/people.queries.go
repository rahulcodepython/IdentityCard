package people

import (
	"fmt"
)

const (
	UpsertPersonQuery = `
        WITH ins AS (
            INSERT INTO people (organization_id, event_id, email, mobile, name, image_url, age, gender, joined_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            ON CONFLICT (event_id, email, mobile) DO UPDATE SET
                name = EXCLUDED.name,
                image_url = EXCLUDED.image_url,
                age = EXCLUDED.age,
                gender = EXCLUDED.gender,
                updated_at = now()
            RETURNING *, (xmax = 0) AS inserted
        )
        SELECT jsonb_build_object(
            'id', ins.id,
            'organization_id', ins.organization_id,
            'event_id', ins.event_id,
            'name', ins.name,
            'email', ins.email,
            'mobile', ins.mobile,
            'image_url', ins.image_url,
            'age', ins.age,
            'gender', ins.gender,
            'joined_at', to_char(ins.joined_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'card_sent_at', to_char(ins.card_sent_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'inserted', ins.inserted
        ) FROM ins;
    `

	GetPersonQuery = `
        SELECT jsonb_build_object(
            'id', p.id,
            'name', p.name,
            'email', p.email,
            'mobile', p.mobile,
            'image_url', p.image_url,
            'age', p.age,
            'gender', p.gender,
            'joined_at', to_char(p.joined_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'card_sent_at', to_char(p.card_sent_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'sub_event_ids', COALESCE((
                SELECT jsonb_agg(pse.sub_event_id)
                FROM people_sub_events pse
                WHERE pse.person_id = p.id
            ), '[]'::jsonb)
        )
        FROM people p
        WHERE p.id = $1 AND p.event_id = $2 AND p.organization_id = $3;
    `

	ListPeopleByIDsQuery = `
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', p.id,
            'name', p.name,
            'email', p.email,
            'mobile', p.mobile,
            'image_url', p.image_url,
            'age', p.age,
            'gender', p.gender,
            'joined_at', to_char(p.joined_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'card_sent_at', to_char(p.card_sent_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'sub_event_ids', COALESCE((
                SELECT jsonb_agg(pse.sub_event_id)
                FROM people_sub_events pse
                WHERE pse.person_id = p.id
            ), '[]'::jsonb)
        )), '[]'::jsonb)
        FROM people p
        WHERE p.event_id = $1 AND p.organization_id = $2 AND p.id = ANY($3::uuid[]);
    `

	UpdatePersonQuery = `
        WITH upd AS (
            UPDATE people
            SET name = $4, image_url = $5, age = $6, gender = $7, updated_at = now()
            WHERE id = $1 AND event_id = $2 AND organization_id = $3
            RETURNING *
        )
        SELECT jsonb_build_object(
            'id', upd.id,
            'name', upd.name,
            'email', upd.email,
            'mobile', upd.mobile,
            'image_url', upd.image_url,
            'age', upd.age,
            'gender', upd.gender,
            'joined_at', to_char(upd.joined_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'card_sent_at', to_char(upd.card_sent_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
        ) FROM upd;
    `

	DeletePersonQuery = `
        DELETE FROM people
        WHERE id = $1 AND event_id = $2 AND organization_id = $3;
    `

	ListPersonSubEventIDsQuery = `
        SELECT COALESCE(jsonb_agg(pse.sub_event_id), '[]'::jsonb)
        FROM people_sub_events pse
        WHERE pse.person_id = $1;
    `

	AddPersonToSubEventQuery = `
        INSERT INTO people_sub_events (person_id, sub_event_id)
        VALUES ($1, $2)
        ON CONFLICT (person_id, sub_event_id) DO NOTHING;
    `

	DeletePersonSubEventsQuery = `
        DELETE FROM people_sub_events
        WHERE person_id = $1;
    `

	MarkCardSentQuery = `
        UPDATE people
        SET card_sent_at = now()
        WHERE id = $1;
    `
)

func BuildListPeopleQuery(whereClause string) string {
	return fmt.Sprintf(`
        SELECT COALESCE(jsonb_agg(jsonb_build_object(
            'id', p.id,
            'name', p.name,
            'email', p.email,
            'mobile', p.mobile,
            'image_url', p.image_url,
            'age', p.age,
            'gender', p.gender,
            'joined_at', to_char(p.joined_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'card_sent_at', to_char(p.card_sent_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
            'sub_event_ids', COALESCE((
                SELECT jsonb_agg(pse.sub_event_id)
                FROM people_sub_events pse
                WHERE pse.person_id = p.id
            ), '[]'::jsonb)
        ) ORDER BY p.created_at DESC), '[]'::jsonb)
        FROM people p
        WHERE %s;
    `, whereClause)
}
