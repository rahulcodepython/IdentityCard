package services

import (
	"errors"
	"time"

	"identitycard-server/internal/pkg/jwt"
)

// oauthStateTTL is short because the state only needs to survive one
// browser round trip to Google and back.
const oauthStateTTL = 10 * time.Minute

// oauthStateClaims carries the intent (login vs. register) through the
// browser round trip to Google and back — the callback is a plain
// redirect, not a fetch, so it can't travel any other way. Signed with
// cfg.JWTSecret (reusing the session-signing key is fine here — this is an
// internal round trip, not a public contract like qrtoken). Register
// intent no longer needs an org name or plan up front: a Google signup
// collects its organization name afterward, on /onboarding.
type oauthStateClaims struct {
	jwt.RegisteredClaims
	Intent string `json:"intent"`
}

var errInvalidOAuthState = errors.New("auth: invalid or expired oauth state")

func signOAuthState(secret, intent string) (string, error) {
	claims := oauthStateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(oauthStateTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Intent: intent,
	}
	return jwt.SignCustom(secret, claims)
}

func parseOAuthState(secret, token string) (oauthStateClaims, error) {
	claims := &oauthStateClaims{}
	if err := jwt.ParseCustom(secret, token, claims); err != nil {
		return oauthStateClaims{}, errInvalidOAuthState
	}
	return *claims, nil
}
