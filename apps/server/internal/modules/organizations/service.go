package organizations

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

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/storage"
)

const maxLogoBytes = 2 << 20 // 2MB

// PNG/JPEG only — not WebP, despite it being a common upload format —
// because the card PDF renderer (internal/modules/cards) embeds this logo
// via a library that only supports these two.
var allowedLogoContentTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
}

type Service struct {
	repo    *Repository
	storage *storage.Storage
}

func NewService(repo *Repository, storage *storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) GetSettings(ctx context.Context, orgID uuid.UUID) (SettingsResponse, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return SettingsResponse{}, httpx.ErrNotFound("organization")
	}
	return SettingsResponse{
		ID: org.ID, Name: org.Name, Slug: org.Slug,
		HasLogo: org.LogoObjectKey.Valid,
	}, nil
}

// UploadLogo stores the image under a key derived from the org id, so a
// re-upload naturally overwrites the previous one (org_id + one file
// extension is deterministic — no orphaned old logos to clean up). The
// bytes are decoded, not just the declared Content-Type trusted, so a
// stored logo is guaranteed embeddable later without surprising
// cards.Service's own defensive re-check.
func (s *Service) UploadLogo(ctx context.Context, orgID uuid.UUID, contentType string, size int64, r io.Reader) error {
	ext, ok := allowedLogoContentTypes[contentType]
	if !ok {
		return httpx.NewError(http.StatusUnsupportedMediaType, "unsupported_type", "logo must be PNG or JPEG")
	}
	if size <= 0 || size > maxLogoBytes {
		return httpx.ErrValidation(map[string]string{"logo": "file must be under 2MB"})
	}

	data, err := io.ReadAll(io.LimitReader(r, maxLogoBytes+1))
	if err != nil {
		return httpx.ErrInternal()
	}
	if len(data) > maxLogoBytes {
		return httpx.ErrValidation(map[string]string{"logo": "file must be under 2MB"})
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err != nil || allowedLogoContentTypes["image/"+format] == "" {
		return httpx.ErrValidation(map[string]string{"logo": "file is not a valid PNG or JPEG image"})
	}

	key := fmt.Sprintf("org-logos/%s.%s", orgID, ext)
	if err := s.storage.Put(ctx, key, bytes.NewReader(data), int64(len(data)), contentType); err != nil {
		return httpx.ErrInternal()
	}
	if _, err := s.repo.UpdateLogo(ctx, orgID, key); err != nil {
		return httpx.ErrInternal()
	}
	return nil
}

// GetLogo is also used internally by cards.Service to embed the logo in a
// generated PDF — not just the HTTP download handler.
func (s *Service) GetLogo(ctx context.Context, orgID uuid.UUID) ([]byte, string, error) {
	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil || !org.LogoObjectKey.Valid {
		return nil, "", httpx.ErrNotFound("logo")
	}
	data, contentType, err := s.storage.Get(ctx, org.LogoObjectKey.String)
	if err != nil {
		return nil, "", httpx.ErrNotFound("logo")
	}
	return data, contentType, nil
}
