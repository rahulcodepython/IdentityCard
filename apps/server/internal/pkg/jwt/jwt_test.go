package jwt

import (
    "crypto/ed25519"
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "testing"
    "time"

    extjwt "github.com/golang-jwt/jwt/v5"
    "github.com/MicahParks/keyfunc/v3"
)

func TestEd25519Verification(t *testing.T) {
    pub, priv, err := ed25519.GenerateKey(rand.Reader)
    if err != nil {
        t.Fatalf("GenerateKey error: %v", err)
    }

    x := base64.RawURLEncoding.EncodeToString(pub)

    jwkJSON := map[string]interface{}{
        "keys": []map[string]interface{}{
            {
                "kty": "OKP",
                "crv": "Ed25519",
                "alg": "EdDSA",
                "kid": "test-key-id",
                "x":   x,
            },
        },
    }
    rawBytes, err := json.Marshal(jwkJSON)
    if err != nil {
        t.Fatalf("Marshal error: %v", err)
    }

    kf, err := keyfunc.NewJWKSetJSON(rawBytes)
    if err != nil {
        t.Fatalf("NewJWKSetJSON error: %v", err)
    }

    claims := &Claims{
        RegisteredClaims: extjwt.RegisteredClaims{
            Subject:   "a3b8c4d2-1111-2222-3333-444455556666",
            ExpiresAt: extjwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  extjwt.NewNumericDate(time.Now()),
        },
    }

    tok := extjwt.NewWithClaims(extjwt.SigningMethodEdDSA, claims)
    tok.Header["kid"] = "test-key-id"

    tokenStr, err := tok.SignedString(priv)
    if err != nil {
        t.Fatalf("SignedString error: %v", err)
    }

    verifier := &Verifier{kf: kf}
    parsedClaims, err := verifier.Parse(tokenStr)
    if err != nil {
        t.Fatalf("Parse error: %v", err)
    }

    uid, err := parsedClaims.UserID()
    if err != nil {
        t.Fatalf("UserID error: %v", err)
    }
    if uid.String() != "a3b8c4d2-1111-2222-3333-444455556666" {
        t.Fatalf("uid mismatch: %v", uid)
    }
}
