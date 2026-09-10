import { createAccessControl } from "better-auth/plugins/access"
import { defaultStatements } from "better-auth/plugins/organization/access"

import { ROLE_ADMIN, ROLE_MEMBER } from "@/lib/constants"

export const statement = defaultStatements

export const ac = createAccessControl(statement)

export const adminRole = ac.newRole({
    organization: ["update", "delete"],
    member: ["create", "update", "delete"],
    invitation: ["create", "cancel"],
    team: ["create", "update", "delete"],
    ac: ["create", "read", "update", "delete"],
})

export const memberRole = ac.newRole({
    organization: [],
    member: ["create", "update", "delete"],
    invitation: ["create", "cancel"],
    team: ["create", "update", "delete"],
    ac: ["read"],
})

export const roles = {
    [ROLE_ADMIN]: adminRole,
    [ROLE_MEMBER]: memberRole,
}
