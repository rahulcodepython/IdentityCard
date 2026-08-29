// Package jwt verifies access tokens issued by better-auth's jwt plugin
// (apps/web/lib/auth.ts) against its JWKS endpoint — golang-jwt/jwt/v5
// does the actual signature/expiry verification, github.com/MicahParks/
// keyfunc/v3 supplies the jwt.Keyfunc backed by a JWKS keyset that's
// fetched once and kept fresh in the background (periodic refresh, plus
// an automatic refetch if an unrecognized kid shows up) — see NewVerifier.
// Go never issues tokens anymore; better-auth owns that entirely.
package jwt

import (
	"context"
	"errors"

	"github.com/MicahParks/keyfunc/v3"
	extjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"identitycard-server/internal/generic"
)

var ErrInvalidToken = errors.New("jwt: invalid or expired token")

// Claims is the JWT payload shape produced by definePayload in
// apps/web/lib/auth.ts — keep the two in lockstep. OrganizationID/Role
// come back as their Go zero values (uuid.Nil / "") when the token was
// issued for a user with no active organization, since encoding/json
// leaves a non-pointer field alone on a JSON null.
type Claims struct {
	extjwt.RegisteredClaims
	OrganizationID uuid.UUID    `json:"organizationId"`
	Role           generic.Role `json:"role"`
}

// UserID parses the standard "sub" claim — better-auth's jwt plugin
// defaults getSubject to the session user's id.
func (c Claims) UserID() (uuid.UUID, error) {
	return uuid.Parse(c.Subject)
}

func (c Claims) HasRole(role generic.Role) bool {
	return c.Role == role
}

type Verifier struct {
	kf keyfunc.Keyfunc
}

// NewVerifier fetches the JWKS at jwksURL once and hands back a verifier
// backed by a keyset kept fresh for the lifetime of ctx — callers should
// pass a context tied to process shutdown so the background refresh goroutine
// stops cleanly.
func NewVerifier(ctx context.Context, jwksURL string) (*Verifier, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, err
	}
	return &Verifier{kf: kf}, nil
}

func (v *Verifier) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := extjwt.ParseWithClaims(tokenStr, claims, v.kf.Keyfunc)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
