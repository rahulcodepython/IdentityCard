package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // decoders for image.DecodeConfig's format sniff below
	_ "image/png"
	"io"
	"net/http"

	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/utils"
)

const maxLogoBytes = 2 << 20 // 2MB

// PNG/JPEG only — not WebP, despite it being a common upload format —
// because the card PDF renderer embeds this logo via a library that only
// supports these two.
var allowedLogoContentTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
}

type OrganizationsService struct {
	repo    *repositories.OrganizationsRepository
	storage *storage.Storage
}

func NewOrganizationsService(repo *repositories.OrganizationsRepository, storage *storage.Storage) *OrganizationsService {
	return &OrganizationsService{repo: repo, storage: storage}
}

func (s *OrganizationsService) GetSettings(ctx context.Context, orgID uuid.UUID) (entities.OrganizationSettingsResponse, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return entities.OrganizationSettingsResponse{}, utils.ErrNotFound("organization")
	}
	return entities.OrganizationSettingsResponse{
		ID: org.ID, Name: org.Name, Slug: org.Slug,
		HasLogo: org.LogoObjectKey.Valid,
	}, nil
}

// UploadLogo stores the image under a key derived from the org id, so a
// re-upload naturally overwrites the previous one (org_id + one file
// extension is deterministic — no orphaned old logos to clean up). The
// bytes are decoded, not just the declared Content-Type trusted, so a
// stored logo is guaranteed embeddable later without surprising
// CardsService's own defensive re-check.
func (s *OrganizationsService) UploadLogo(ctx context.Context, orgID uuid.UUID, contentType string, size int64, r io.Reader) error {
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
	if err := s.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return utils.ErrInternal()
	}
	if _, err := s.repo.UpdateLogo(ctx, orgID, key); err != nil {
		return utils.ErrInternal()
	}
	return nil
}

// GetLogo is also used internally by CardsService to embed the logo in a
// generated PDF — not just the HTTP download handler.
func (s *OrganizationsService) GetLogo(ctx context.Context, orgID uuid.UUID) ([]byte, string, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil || !org.LogoObjectKey.Valid {
		return nil, "", utils.ErrNotFound("logo")
	}
	data, contentType, err := s.storage.Get(ctx, org.LogoObjectKey.String)
	if err != nil {
		return nil, "", utils.ErrNotFound("logo")
	}
	return data, contentType, nil
}
