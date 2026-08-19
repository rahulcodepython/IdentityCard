package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// generateOTP returns a zero-padded 6-digit numeric code — same shape as
// the device-pairing OTP in internal/modules/devices.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
