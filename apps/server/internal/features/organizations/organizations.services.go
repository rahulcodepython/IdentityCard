package organizations

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "image"
    _ "image/jpeg"
    _ "image/png"
    "io"
    "net/http"

    "github.com/google/uuid"

    "identitycard-server/internal/generic"
    "identitycard-server/internal/pkg/postgres"
    "identitycard-server/internal/utils"
)

const maxLogoBytes = 2 << 20 // 2MB

var allowedLogoContentTypes = map[string]string{
    "image/png":  "png",
    "image/jpeg": "jpg",
}

func (a *App) GetSettings(ctx context.Context, orgID uuid.UUID) (OrganizationSettingsResponse, error) {
    org, err := a.GetByID(ctx, orgID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrOrgsNotFound) {
            return OrganizationSettingsResponse{}, utils.ErrNotFound("Organization not found.", err)
        }
        return OrganizationSettingsResponse{}, utils.ErrInternal("Failed to fetch organization settings.", err)
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

func (a *App) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
    if err := a.Delete(ctx, orgID); err != nil {
        return utils.ErrInternal("Failed to delete organization.", err)
    }
    return nil
}

func (a *App) DeleteLogo(ctx context.Context, orgID uuid.UUID) error {
    org, err := a.GetByID(ctx, orgID)
    if err != nil {
        if errors.Is(err, postgres.ErrNotFound) || errors.Is(err, generic.ErrOrgsNotFound) {
            return utils.ErrNotFound("Organization not found.", err)
        }
        return utils.ErrInternal("Failed to fetch organization.", err)
    }
    if org.LogoObjectKey != nil && *org.LogoObjectKey != "" {
        _ = a.storage.Delete(ctx, *org.LogoObjectKey)
    }
    if err := a.DeleteLogoRepo(ctx, orgID); err != nil {
        return utils.ErrInternal("Failed to delete organization logo.", err)
    }
    return nil
}

func (a *App) UploadLogo(ctx context.Context, orgID uuid.UUID, contentType string, size int64, r io.Reader) error {
    ext, ok := allowedLogoContentTypes[contentType]
    if !ok {
        return utils.NewError(http.StatusUnsupportedMediaType, "Logo must be PNG or JPEG.", nil)
    }
    if size <= 0 || size > maxLogoBytes {
        return utils.ErrValidation(map[string]string{"logo": "File must be under 2MB."})
    }

    data, err := io.ReadAll(io.LimitReader(r, maxLogoBytes+1))
    if err != nil {
        return utils.ErrInternal("Failed to read uploaded logo.", err)
    }
    if len(data) > maxLogoBytes {
        return utils.ErrValidation(map[string]string{"logo": "File must be under 2MB."})
    }
    if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || allowedLogoContentTypes["image/"+format] == "" {
        return utils.ErrValidation(map[string]string{"logo": "File is not a valid PNG or JPEG image."})
    }

    key := fmt.Sprintf("org-logos/%s.%s", orgID, ext)
    if err := a.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
        return utils.ErrInternal("Failed to store organization logo.", err)
    }
    if _, err := a.UpdateLogo(ctx, orgID, key); err != nil {
        return utils.ErrInternal("Failed to update organization logo.", err)
    }
    return nil
}

func (a *App) GetLogo(ctx context.Context, orgID uuid.UUID) ([]byte, string, error) {
    org, err := a.GetByID(ctx, orgID)
    if err != nil || !org.HasLogo || org.LogoObjectKey == nil || *org.LogoObjectKey == "" {
        return nil, "", utils.ErrNotFound("Logo not found.", err)
    }
    data, contentType, err := a.storage.Get(ctx, *org.LogoObjectKey)
    if err != nil {
        return nil, "", utils.ErrNotFound("Logo not found in storage.", err)
    }
    return data, contentType, nil
}
