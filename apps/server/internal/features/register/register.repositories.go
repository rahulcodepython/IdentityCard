package register

import (
    "context"

    "identitycard-server/internal/pkg/postgres"
)

func (r *App) GetPublicApplyConfigRepository(ctx context.Context, eventFormID string) (*PublicApplyConfigResponse, error) {
    return postgres.QueryJSON[PublicApplyConfigResponse](ctx, r.DB, GetPublicApplyQuery, eventFormID)
}

func (r *App) SubmitApplicationRepository(ctx context.Context, eventFormID, userID, name, email string, dataJSON []byte) (string, error) {
    var statusCode string
    var eventID string
    err := r.DB.QueryRow(ctx, SubmitApplicationCTEQuery, eventFormID, userID, name, email, string(dataJSON)).Scan(&statusCode, &eventID)
    if err != nil {
        return "", err
    }
    return statusCode, nil
}
