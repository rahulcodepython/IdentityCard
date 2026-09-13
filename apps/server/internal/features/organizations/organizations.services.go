package organizations

import (
    "context"
    "errors"
    "uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
)

const maxLogoBytes = generic.MaxLogoBytes

var allowedLogoContentTypes = map[string]string{
    generic.ContentTypePNG:  "png",
    generic.ContentTypeJPEG: "jpg",
}

func (a *App) ListOrganizations(ctx context.Context, userID uuid.UUID) ([]ListOrganizationsResponse, error) {
    orgs, err := a.List(ctx, userID)
    if err != nil {
        return nil, utils.ErrInternal("Failed to list organizations.", err)
    }
    return orgs, nil
}

func (a *App) GetSettings(ctx context.Context, orgID uuid.UUID) (OrganizationSettingsResponse, error) {
    org, err := a.GetByID(ctx, orgID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrOrgsNotFound) {
            return OrganizationSettingsResponse{}, utils.ErrNotFound("Organization not found.", err)
        }
        return OrganizationSettingsResponse{}, utils.ErrInternal("Failed to get organization settings.", err)
    }
    return OrganizationSettingsResponse{
        ID:      org.ID,
        Name:    org.Name,
        Slug:    org.Slug,
        HasLogo: org.HasLogo,
    }, nil
}

func (a *App) UpdateSettings(ctx context.Context, orgID uuid.UUID, req UpdateOrganizationSettingsRequest) (OrganizationSettingsResponse, error) {
    org, err := a.UpdateName(ctx, orgID, req.Name)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrOrgsNotFound) {
            return OrganizationSettingsResponse{}, utils.ErrNotFound("Organization not found.", err)
        }
        return OrganizationSettingsResponse{}, utils.ErrInternal("Failed to update organization settings.", err)
    }
    return OrganizationSettingsResponse{
        ID:      org.ID,
        Name:    org.Name,
        Slug:    org.Slug,
        HasLogo: org.HasLogo,
    }, nil
}

func (a *App) DeleteOrganization(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) error {
    res, err := a.Delete(ctx, orgID, userID)
    if err != nil {
        return utils.ErrInternal("Failed to delete organization.", err)
    }

    if res != nil && res.UserOrgCount <= 1 {
        return utils.ErrCannotDeleteLastOrg(generic.ErrOrgsCannotDeleteLastOrg.Error(), generic.ErrOrgsCannotDeleteLastOrg)
    }

    if res == nil || !res.Deleted {
        return utils.ErrNotFound("Organization not found or access denied.", nil)
    }

    return nil
}
