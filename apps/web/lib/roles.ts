// System/Platform User Roles
export const USER_ROLE_ADMIN = "admin"
export const USER_ROLE_USER = "user"

export type UserRole = typeof USER_ROLE_ADMIN | typeof USER_ROLE_USER

// Organization Member Roles
export const ROLE_ADMIN = "admin"
export const ROLE_MEMBER = "member"

export type OrgRole = typeof ROLE_ADMIN | typeof ROLE_MEMBER
