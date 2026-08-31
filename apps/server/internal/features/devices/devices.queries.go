package devices

const (
	CreateDeviceQuery = `
        WITH ins AS (
            INSERT INTO devices (organization_id, name, otp_code, otp_expires_at)
            VALUES ($1, $2, $3, $4)
            RETURNING *
        )
        SELECT row_to_json(ins) FROM ins;
    `

	ListDevicesQuery = `
        SELECT COALESCE(jsonb_agg(row_to_json(d) ORDER BY d.created_at DESC), '[]'::jsonb)
        FROM devices d
        WHERE d.organization_id = $1;
    `

	GetDeviceQuery = `
        SELECT row_to_json(d)
        FROM devices d
        WHERE d.id = $1 AND d.organization_id = $2;
    `

	GetPendingDeviceByOTPQuery = `
        SELECT row_to_json(d)
        FROM devices d
        WHERE d.otp_code = $1 AND d.status = 'pending' AND d.otp_expires_at > now();
    `

	MarkDeviceVerifiedQuery = `
        WITH upd AS (
            UPDATE devices
            SET status = 'verified', key_hash = $2, verified_at = now(),
                otp_code = NULL, otp_expires_at = NULL, updated_at = now()
            WHERE id = $1
            RETURNING *
        )
        SELECT row_to_json(upd) FROM upd;
    `

	GetDeviceByKeyHashQuery = `
        SELECT row_to_json(d)
        FROM devices d
        WHERE d.key_hash = $1 AND d.status = 'verified';
    `

	RevokeDeviceQuery = `
        UPDATE devices
        SET status = 'revoked', key_hash = NULL, updated_at = now()
        WHERE id = $1 AND organization_id = $2 AND status != 'revoked';
    `
)
