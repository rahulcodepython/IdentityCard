package applicants

import (
    "context"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
)

func (r *App) GetAssignedFormRepository(ctx context.Context, eventID string) (*FormSummary, error) {
    return postgres.QueryJSON[FormSummary](ctx, r.DB, GetEventAssignedFormQuery, eventID)
}

func (r *App) QueryApplicantsWithFiltersRepository(ctx context.Context, whereClause string, limitIdx, offsetIdx, pageIdx, limitIdx2 int, args ...any) (*generic.PaginatedResponse[[]ApplicantItem], error) {
    sqlQuery := BuildFilteredApplicantsQuery(whereClause, limitIdx, offsetIdx, pageIdx, limitIdx2)
    result, err := postgres.QueryJSON[generic.PaginatedResponse[[]ApplicantItem]](ctx, r.DB, sqlQuery, args...)
    if err != nil {
        return nil, err
    }
    if result == nil {
        return &generic.PaginatedResponse[[]ApplicantItem]{
            Data: []ApplicantItem{},
        }, nil
    }
    if result.Data == nil {
        result.Data = []ApplicantItem{}
    }
    return result, nil
}
