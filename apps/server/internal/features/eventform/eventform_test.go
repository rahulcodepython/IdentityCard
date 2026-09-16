package eventform

import (
    "context"
    "testing"
    "time"

    "identitycard-server/internal/config"
    "identitycard-server/internal/pkg/postgres"
)

func TestEventFormLifecycle(t *testing.T) {
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

    // 1. Get current event form
    form, err := app.GetEventFormService(ctx, eventID)
    if err != nil {
        t.Fatalf("GetEventFormService failed: %v", err)
    }
    t.Logf("Found event form: id=%s, name=%s, is_locked=%v, total_applicants=%d, can_delete=%v",
        form.ID, form.Name, form.IsLocked, form.TotalApplicants, form.CanDelete)

    // 2. Since total_applicants > 0, deletion should fail
    if form.TotalApplicants > 0 {
        err = app.DeleteEventFormService(ctx, eventID)
        if err != ErrCannotDeleteWithApplicants {
            t.Fatalf("Expected ErrCannotDeleteWithApplicants, got %v", err)
        }
        t.Log("Successfully verified deletion guard with applicants")
    }

    // 3. Since form is locked, update should fail
    if form.IsLocked {
        name := "Attempted Rename"
        _, err = app.UpdateEventFormService(ctx, eventID, UpdateEventFormRequest{Name: &name})
        if err != ErrFormIsLocked {
            t.Fatalf("Expected ErrFormIsLocked, got %v", err)
        }
        t.Log("Successfully verified lock guard prevents editing")
    }

    // 4. Test validation of expiration in past
    past := time.Now().Add(-1 * time.Hour)
    _, err = app.CreateEventFormService(ctx, "00000000-0000-0000-0000-000000000000", CreateEventFormRequest{
        Source:        "scratch",
        Name:          "Test",
        MaxApplicants: 10,
        ExpiresAt:     past,
    })
    if err != ErrInvalidExpiresAt {
        t.Fatalf("Expected ErrInvalidExpiresAt, got %v", err)
    }
}
