// Package qrtoken signs and (for a later phase's scanner-verify endpoint)
// will parse the compact token embedded in an ID card's QR code. Signed
// with its own secret (config.QRSecret) rather than the session JWT
// secret — a card's validity window is measured in months, a login
// session in minutes, and they should never have to rotate together.
package qrtoken

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalid = errors.New("qrtoken: invalid or expired token")

type Claims struct {
	jwt.RegisteredClaims
	PersonID       uuid.UUID `json:"pid"`
	EventID        uuid.UUID `json:"eid"`
	OrganizationID uuid.UUID `json:"oid"`
}

// Issue is deterministic in every field but signature timing, so
// regenerating a person's card before the event's schedule changes
// produces a token identical in meaning (same subject, same expiry).
func Issue(secret string, personID, eventID, orgID uuid.UUID, expiresAt time.Time) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		PersonID: personID, EventID: eventID, OrganizationID: orgID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// Parse is unused until the Phase 5 scanner-verify endpoint exists, but
// the token format is fixed now — see Issue.
func Parse(secret, token string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalid
	}
	return claims, nil
}
