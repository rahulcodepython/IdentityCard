package forms

import (
    "context"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
)

// ListRepository queries total count and paginated forms in a single network round-trip.
func (r *App) ListRepository(ctx context.Context, search string, page, limit, offset int) (*generic.PaginatedResponse[[]Form], error) {
    result, err := postgres.QueryJSON[generic.PaginatedResponse[[]Form]](ctx, r.DB, ListFormsQuery, search, limit, offset, page)
    if err != nil {
        return nil, err
    }
    if result == nil {
        return &generic.PaginatedResponse[[]Form]{
            Data:  []Form{},
            Total: 0,
            Page:  page,
            Limit: limit,
        }, nil
    }
    if result.Data == nil {
        result.Data = []Form{}
    }
    return result, nil
}

// GetRepository finds a form by UUID in a single network round-trip.
func (r *App) GetRepository(ctx context.Context, id string) (*Form, error) {
    return postgres.QueryJSON[Form](ctx, r.DB, GetFormQuery, id)
}

// CreateRepository inserts a new form with its initial JSONB fields.
func (r *App) CreateRepository(ctx context.Context, name string, fieldsJSON []byte) (*Form, error) {
    return postgres.QueryJSON[Form](ctx, r.DB, CreateFormQuery, name, string(fieldsJSON))
}

// UpdateRepository updates form metadata (name).
func (r *App) UpdateRepository(ctx context.Context, id string, req UpdateFormRequest) (*Form, error) {
    return postgres.QueryJSON[Form](ctx, r.DB, UpdateFormQuery, id, req.Name)
}

// UpdateFieldsRepository updates the fields JSONB array atomically.
func (r *App) UpdateFieldsRepository(ctx context.Context, id string, fieldsJSON []byte) (*Form, error) {
    return postgres.QueryJSON[Form](ctx, r.DB, UpdateFormFieldsQuery, id, string(fieldsJSON))
}

// DeleteRepository removes a form by UUID.
func (r *App) DeleteRepository(ctx context.Context, id string) error {
    ct, err := r.DB.Exec(ctx, DeleteFormQuery, id)
    if err != nil {
        return err
    }
    if ct.RowsAffected() == 0 {
        return postgres.ErrNotFound
    }
    return nil
}
