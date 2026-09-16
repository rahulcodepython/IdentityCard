package applicants

import (
    "context"
    "testing"

    "identitycard-server/internal/config"
    "identitycard-server/internal/pkg/postgres"
)

func TestListApplicantsFilter(t *testing.T) {
    ctx := context.Background()
    cfg := &config.Config{
        DatabaseURL: "postgres://identitycard:identitycard@localhost:5432/identitycard?sslmode=disable",
    }
    pool, err := postgres.Connect(ctx, cfg)
    if err != nil {
        t.Skipf("cannot connect to postgres: %v", err)
    }
    defer postgres.Close(pool)

    app := NewApp(pool)
    eventID := "7e00b724-b5d0-454f-9b63-7a7362049fed"
    filters := []ApplicantFilter{
        {
            Field: "checkbox_5",
            Op:    "contains",
            Value: "option_1",
        },
    }

    res, err := app.ListApplicantsService(ctx, eventID, "", filters, 1, 30)
    if err != nil {
        t.Fatalf("ListApplicantsService failed: %v", err)
    }

    t.Logf("Result: total=%d, len(data)=%d", res.Total, len(res.Data))
}
