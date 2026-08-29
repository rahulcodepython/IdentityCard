import { createAccessControl } from "better-auth/plugins/access"
import { defaultStatements } from "better-auth/plugins/organization/access"

import { ROLE_ADMIN, ROLE_SCANNER, ROLE_SUPER_ADMIN } from "@/lib/roles"

// Shared between the server config (lib/auth.ts) and the client config
// (lib/auth-client.ts) — better-auth needs the identical ac/roles object
// shape on both sides for its type inference to line up.
//
// Replaces the organization plugin's default owner/admin/member roles
// with this app's three roles (super_admin/admin/scanner — see
// lib/roles.ts). scanner intentionally gets zero organization-management
// permissions: it's a device-pairing role, not a human admin seat.
export const statement = defaultStatements

export const ac = createAccessControl(statement)

export const superAdminRole = ac.newRole({
    organization: ["update", "delete"],
    member: ["create", "update", "delete"],
    invitation: ["create", "cancel"],
    team: ["create", "update", "delete"],
    ac: ["create", "read", "update", "delete"],
})

export const adminRole = ac.newRole({
    organization: ["update"],
    member: ["create", "update", "delete"],
    invitation: ["create", "cancel"],
    team: ["create", "update", "delete"],
    ac: ["read"],
})

export const scannerRole = ac.newRole({
    organization: [],
    member: [],
    invitation: [],
    team: [],
    ac: [],
})

export const roles = {
    [ROLE_SUPER_ADMIN]: superAdminRole,
    [ROLE_ADMIN]: adminRole,
    [ROLE_SCANNER]: scannerRole,
}
