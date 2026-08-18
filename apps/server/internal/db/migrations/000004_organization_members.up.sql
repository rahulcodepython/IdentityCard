CREATE TABLE organization_members (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    roles           TEXT[] NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (organization_id, user_id),
    CONSTRAINT organization_members_roles_valid CHECK (
        roles <@ ARRAY['super_admin', 'admin', 'scanner']::TEXT[]
        AND array_length(roles, 1) > 0
    )
);

CREATE INDEX organization_members_user_id_idx ON organization_members (user_id);

-- Exactly one super_admin per organization.
CREATE UNIQUE INDEX organization_members_one_super_admin_idx
    ON organization_members (organization_id)
    WHERE 'super_admin' = ANY (roles);
