// Package config loads and validates process configuration from the
// environment. It fails fast at startup rather than surfacing a missing
// setting deep inside a request handler.
package config

import (
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/joho/godotenv"
)

type Config struct {
    Env  string // "development" | "production"
    Port string

    DatabaseURL string

    // BetterAuthJWKSURL points at the Next.js app's better-auth JWKS
    // endpoint (see apps/web/lib/auth.ts) — RequireAuth verifies every
    // bearer token against it. Fetched once and cached/refreshed in the
    // background, not re-fetched per request (internal/pkg/jwt).
    BetterAuthJWKSURL string

    WebOrigin string // the Next.js origin allowed by CORS, e.g. http://localhost:3000

    S3Endpoint  string
    S3AccessKey string
    S3SecretKey string
    S3Bucket    string
    S3UseSSL    bool

    // QRSecret signs ID-card QR tokens (internal/qrtoken) — deliberately
    // separate from JWTSecret so rotating session vs. card-verification
    // keys never have to happen together.
    QRSecret string

    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    SMTPFrom     string

    // Resend configuration
    ResendAPIKey string
    ResendFrom   string

    // Redis and Cache configurations
    RedisURL      string
    RedisHost     string
    RedisPort     string
    RedisPassword string
    RedisDB       int

    // Database connection pooling & timeouts
    DBMaxOpenConns        int
    DBMaxIdleConns        int
    DBConnMaxLifetimeMin  int
    DBConnMaxIdleTimeMin  int
    DBStatementTimeoutSec int
    RequestTimeoutSec     int
}

// Load reads configuration from the environment (loading a local .env file
// first, if present) and returns an error naming every missing required
// variable at once instead of failing one at a time.
func Load() (*Config, error) {
    _ = godotenv.Load()

    var missing []string
    require := func(key string) string {
        v := os.Getenv(key)
        if v == "" {
            missing = append(missing, key)
        }
        return v
    }

    cfg := &Config{
        Env:                   getOr("APP_ENV", "development"),
        Port:                  getOr("PORT", "8000"),
        DatabaseURL:           require("DATABASE_URL"),
        BetterAuthJWKSURL:     getOr("BETTER_AUTH_JWKS_URL", "http://localhost:3000/api/auth/jwks"),
        WebOrigin:             getOr("WEB_ORIGIN", "http://localhost:3000"),
        S3Endpoint:            getOr("S3_ENDPOINT", "localhost:9000"),
        S3AccessKey:           getOr("S3_ACCESS_KEY", "minioadmin"),
        S3SecretKey:           getOr("S3_SECRET_KEY", "minioadmin"),
        S3Bucket:              getOr("S3_BUCKET", "identitycard"),
        S3UseSSL:              getBool("S3_USE_SSL", false),
        QRSecret:              require("QR_SECRET"),
        SMTPHost:              getOr("SMTP_HOST", "localhost"),
        SMTPPort:              getInt("SMTP_PORT", 1025),
        SMTPUsername:          getOr("SMTP_USERNAME", ""),
        SMTPPassword:          getOr("SMTP_PASSWORD", ""),
        SMTPFrom:              getOr("SMTP_FROM", "no-reply@identitycard.local"),
        ResendAPIKey:          getOr("RESEND_API_KEY", ""),
        ResendFrom:            getOr("RESEND_FROM", getOr("SMTP_FROM", "no-reply@identitycard.local")),
        RedisURL:              getOr("REDIS_URL", "redis://localhost:6379"),
        RedisHost:             getOr("REDIS_HOST", "localhost"),
        RedisPort:             getOr("REDIS_PORT", "6379"),
        RedisPassword:         getOr("REDIS_PASSWORD", ""),
        RedisDB:               getInt("REDIS_DB", 0),
        DBMaxOpenConns:        getInt("DB_MAX_OPEN_CONNS", 25),
        DBMaxIdleConns:        getInt("DB_MAX_IDLE_CONNS", 10),
        DBConnMaxLifetimeMin:  getInt("DB_CONN_MAX_LIFETIME_MIN", 60),
        DBConnMaxIdleTimeMin:  getInt("DB_CONN_MAX_IDLE_TIME_MIN", 30),
        DBStatementTimeoutSec: getInt("DB_STATEMENT_TIMEOUT_SEC", 15),
        RequestTimeoutSec:     getInt("REQUEST_TIMEOUT_SEC", 30),
    }

    if len(missing) > 0 {
        return nil, fmt.Errorf("config: missing required environment variables: %s", strings.Join(missing, ", "))
    }
    return cfg, nil
}

func getOr(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}

func getInt(key string, fallback int) int {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    n, err := strconv.Atoi(v)
    if err != nil {
        return fallback
    }
    return n
}

func getBool(key string, fallback bool) bool {
    v := os.Getenv(key)
    if v == "" {
        return fallback
    }
    b, err := strconv.ParseBool(v)
    if err != nil {
        return fallback
    }
    return b
}
