package organizations

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/google/uuid"

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
		return OrganizationSettingsResponse{}, utils.ErrNotFound("organization")
	}
	return OrganizationSettingsResponse{
		ID:      org.ID,
		Name:    org.Name,
		Slug:    org.Slug,
		HasLogo: org.LogoObjectKey.Valid,
	}, nil
}

func (a *App) UpdateSettings(ctx context.Context, orgID uuid.UUID, req UpdateOrganizationSettingsRequest) (OrganizationSettingsResponse, error) {
	org, err := a.UpdateName(ctx, orgID, req.Name)
	if err != nil {
		return OrganizationSettingsResponse{}, utils.ErrInternal()
	}
	return OrganizationSettingsResponse{
		ID:      org.ID,
		Name:    org.Name,
		Slug:    org.Slug,
		HasLogo: org.LogoObjectKey.Valid,
	}, nil
}

func (a *App) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	if err := a.Delete(ctx, orgID); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) DeleteLogo(ctx context.Context, orgID uuid.UUID) error {
	org, err := a.GetByID(ctx, orgID)
	if err != nil {
		return utils.ErrNotFound("organization")
	}
	if org.LogoObjectKey.Valid {
		_ = a.storage.Delete(ctx, org.LogoObjectKey.String)
	}
	if _, err := a.queries.DeleteOrganizationLogo(ctx, orgID); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) UploadLogo(ctx context.Context, orgID uuid.UUID, contentType string, size int64, r io.Reader) error {
	ext, ok := allowedLogoContentTypes[contentType]
	if !ok {
		return utils.NewError(http.StatusUnsupportedMediaType, "unsupported_type", "logo must be PNG or JPEG")
	}
	if size <= 0 || size > maxLogoBytes {
		return utils.ErrValidation(map[string]string{"logo": "file must be under 2MB"})
	}

	data, err := io.ReadAll(io.LimitReader(r, maxLogoBytes+1))
	if err != nil {
		return utils.ErrInternal()
	}
	if len(data) > maxLogoBytes {
		return utils.ErrValidation(map[string]string{"logo": "file must be under 2MB"})
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || allowedLogoContentTypes["image/"+format] == "" {
		return utils.ErrValidation(map[string]string{"logo": "file is not a valid PNG or JPEG image"})
	}

	key := fmt.Sprintf("org-logos/%s.%s", orgID, ext)
	if err := a.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return utils.ErrInternal()
	}
	if _, err := a.UpdateLogo(ctx, orgID, key); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

func (a *App) GetLogo(ctx context.Context, orgID uuid.UUID) ([]byte, string, error) {
	org, err := a.GetByID(ctx, orgID)
	if err != nil || !org.LogoObjectKey.Valid {
		return nil, "", utils.ErrNotFound("logo")
	}
	data, contentType, err := a.storage.Get(ctx, org.LogoObjectKey.String)
	if err != nil {
		return nil, "", utils.ErrNotFound("logo")
	}
	return data, contentType, nil
}
