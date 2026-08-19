// Package jwt wraps golang-jwt/jwt/v5 — the only place in the codebase
// that imports it directly. Callers get typed Issue/Parse functions and
// this package's own claim types, never a raw *jwt.Token.
package jwt

import (
	"errors"
	"slices"
	"time"

	extjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"identitycard-server/internal/generic"
)

var ErrInvalidToken = errors.New("jwt: invalid or expired token")

// RegisteredClaims re-exports the standard JWT claim set so callers can
// embed it in their own claims structs without importing golang-jwt
// directly — see SignCustom/ParseCustom.
type RegisteredClaims = extjwt.RegisteredClaims

func NewNumericDate(t time.Time) *extjwt.NumericDate { return extjwt.NewNumericDate(t) }

// SignCustom signs an arbitrary claims struct (embedding RegisteredClaims)
// — for internal, non-session tokens like the OAuth state param, where the
// shape doesn't fit AccessClaims/RefreshClaims.
func SignCustom(secret string, claims extjwt.Claims) (string, error) {
	return extjwt.NewWithClaims(extjwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseCustom parses a token signed by SignCustom into dst, a pointer to a
// claims struct embedding RegisteredClaims.
func ParseCustom(secret, tokenStr string, dst extjwt.Claims) error {
	token, err := extjwt.ParseWithClaims(tokenStr, dst, func(t *extjwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}
	return nil
}

// AccessClaims is embedded in the short-lived access token. Roles are
// baked in at issue time so RBAC checks never need a DB round trip; a role
// change takes effect the next time the access token is refreshed.
type AccessClaims struct {
	extjwt.RegisteredClaims
	UserID         uuid.UUID      `json:"user_id"`
	OrganizationID uuid.UUID      `json:"organization_id"`
	Roles          []generic.Role `json:"roles"`
}

func (c AccessClaims) HasRole(role generic.Role) bool {
	return slices.Contains(c.Roles, role)
}

// RefreshClaims identifies only the user and a random token ID (jti). The
// jti is what gets stored/rotated in Redis so a refresh token can be
// revoked (logout, reuse detection) without invalidating every session.
type RefreshClaims struct {
	extjwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
}

func IssueAccessToken(secret string, ttl time.Duration, userID, orgID uuid.UUID, roles []generic.Role) (string, error) {
	claims := AccessClaims{
		RegisteredClaims: extjwt.RegisteredClaims{
			ExpiresAt: extjwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  extjwt.NewNumericDate(time.Now()),
		},
		UserID:         userID,
		OrganizationID: orgID,
		Roles:          roles,
	}
	return extjwt.NewWithClaims(extjwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseAccessToken(secret, tokenStr string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := extjwt.ParseWithClaims(tokenStr, claims, func(t *extjwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// IssueRefreshToken returns the signed token together with its jti, which
// the caller stores in Redis (see services.AuthService) as the source of
// truth for whether the token is still valid.
func IssueRefreshToken(secret string, ttl time.Duration, userID uuid.UUID) (token, jti string, err error) {
	jti = uuid.NewString()
	claims := RefreshClaims{
		RegisteredClaims: extjwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: extjwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  extjwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}
	token, err = extjwt.NewWithClaims(extjwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, jti, err
}

func ParseRefreshToken(secret, tokenStr string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := extjwt.ParseWithClaims(tokenStr, claims, func(t *extjwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
