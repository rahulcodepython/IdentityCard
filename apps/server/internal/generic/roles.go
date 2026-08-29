// Package generic holds shared types with no business logic of their own:
// RBAC role constants and per-domain sentinel errors. Nothing here depends
// on any other internal package, so every domain can import it freely
// without risking an import cycle.
package generic

// Role is one of the roles an organization member can hold. Exactly one
// per membership — these are better-auth's organization-plugin custom
// role names (see apps/web/lib/auth-access-control.ts), which the JWT
// plugin bakes into the "role" claim at issue time (see
// apps/web/lib/auth.ts's definePayload).
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleScanner    Role = "scanner"
)
