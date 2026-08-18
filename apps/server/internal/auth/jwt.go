package auth

import (
	"errors"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Role is one of the roles an organization member can hold. A member may
// hold more than one at once (e.g. an admin who also carries the scanner
// role), so roles are always represented as a slice.
type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleScanner    Role = "scanner"
)

var ErrInvalidToken = errors.New("auth: invalid or expired token")

// AccessClaims is embedded in the short-lived access token. Roles are
// baked in at issue time so RBAC checks never need a DB round trip; a role
// change takes effect the next time the access token is refreshed.
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID         uuid.UUID `json:"user_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Roles          []Role    `json:"roles"`
}

func (c AccessClaims) HasRole(role Role) bool {
	return slices.Contains(c.Roles, role)
}

// RefreshClaims identifies only the user and a random token ID (jti). The
// jti is what gets stored/rotated in Redis so a refresh token can be
// revoked (logout, reuse detection) without invalidating every session.
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func IssueAccessToken(secret string, ttl time.Duration, userID, orgID uuid.UUID, roles []Role) (string, error) {
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:         userID,
		OrganizationID: orgID,
		Roles:          roles,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseAccessToken(secret, tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// IssueRefreshToken returns the signed token together with its jti, which
// the caller stores in Redis (see internal/modules/auth) as the source of
// truth for whether the token is still valid.
func IssueRefreshToken(secret string, ttl time.Duration, userID uuid.UUID) (token, jti string, err error) {
	jti = uuid.NewString()
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, jti, err
}

func ParseRefreshToken(secret, tokenStr string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
