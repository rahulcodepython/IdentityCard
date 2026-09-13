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
    "fmt"
    "time"
    "uuid"

    "github.com/MicahParks/jwkset"
    "github.com/MicahParks/keyfunc/v3"
    extjwt "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("jwt: invalid or expired token")

// Claims is the JWT payload shape produced by better-auth's jwt plugin.
type Claims struct {
    extjwt.RegisteredClaims
    ID             string `json:"id,omitempty"`
    Email          string `json:"email,omitempty"`
    Name           string `json:"name,omitempty"`
    OrganizationID string `json:"organizationId,omitempty"`
    Role           string `json:"role,omitempty"`
    UserRole       string `json:"userRole,omitempty"`
}

// UserID parses the standard "sub" claim (falling back to "id") — better-auth's jwt plugin
// defaults getSubject to the session user's id.
func (c Claims) UserID() (uuid.UUID, error) {
    sub := c.Subject
    if sub == "" {
        sub = c.ID
    }
    if sub == "" {
        return uuid.Nil(), errors.New("jwt: missing subject or id claim")
    }
    return uuid.Parse(sub)
}

// OrgID parses the optional "organizationId" claim.
func (c Claims) OrgID() (uuid.UUID, error) {
    if c.OrganizationID == "" {
        return uuid.Nil(), errors.New("jwt: missing organizationId claim")
    }
    return uuid.Parse(c.OrganizationID)
}

type Verifier struct {
    kf keyfunc.Keyfunc
}

// NewVerifier fetches the JWKS at jwksURL and hands back a verifier
// backed by a keyset kept fresh in the background.
// Uses NoErrorReturnFirstHTTPReq: true so the Go server does not crash on startup
// if the web app (JWKS endpoint) is still starting up.
func NewVerifier(ctx context.Context, jwksURL string) (*Verifier, error) {
    storage, err := jwkset.NewStorageFromHTTP(jwksURL, jwkset.HTTPClientStorageOptions{
        Ctx:                       ctx,
        NoErrorReturnFirstHTTPReq: true,
        RefreshInterval:           time.Hour,
    })
    if err != nil {
        // Fallback to NewDefaultCtx if custom HTTP storage creation fails
        kf, kfErr := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
        if kfErr != nil {
            return nil, kfErr
        }
        return &Verifier{kf: kf}, nil
    }

    kf, err := keyfunc.New(keyfunc.Options{
        Ctx:     ctx,
        Storage: storage,
    })
    if err != nil {
        return nil, err
    }
    return &Verifier{kf: kf}, nil
}

func (v *Verifier) Parse(tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := extjwt.ParseWithClaims(tokenStr, claims, v.kf.Keyfunc)
    if err != nil || !token.Valid {
        if err != nil {
            return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
        }
        return nil, ErrInvalidToken
    }
    return claims, nil
}
