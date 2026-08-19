// Package generic holds shared types with no business logic of their own:
// RBAC role constants and per-domain sentinel errors. Nothing here depends
// on any other internal package, so every domain can import it freely
// without risking an import cycle.
package generic

// Role is one of the roles an organization member can hold. A member may
// hold more than one at once (e.g. an admin who also carries the scanner
// role), so roles are always represented as a slice.
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleScanner    Role = "scanner"
)
