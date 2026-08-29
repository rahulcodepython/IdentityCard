// Single source of truth for organization role names — mirrors
// apps/server/internal/generic/roles.go. Keep these two files in lockstep:
// Go's RequireRole() compares JWT claims against exactly these strings.
export const ROLE_SUPER_ADMIN = "super_admin"
export const ROLE_ADMIN = "admin"
export const ROLE_SCANNER = "scanner"

export type OrgRole = typeof ROLE_SUPER_ADMIN | typeof ROLE_ADMIN | typeof ROLE_SCANNER
