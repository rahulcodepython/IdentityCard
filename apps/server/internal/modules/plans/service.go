package plans

import (
	"context"
	"encoding/json"

	"identitycard-server/internal/httpx"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]PlanResponse, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, httpx.ErrInternal()
	}

	resp := make([]PlanResponse, len(rows))
	for i, row := range rows {
		var price PriceConfig
		_ = json.Unmarshal(row.PriceConfig, &price) // seeded/admin-authored JSON; malformed just renders as zero values

		resp[i] = PlanResponse{
			ID:    row.ID,
			Code:  row.Code,
			Kind:  row.Kind,
			Tier:  row.Tier,
			Name:  row.Name,
			Price: price,
		}
	}
	return resp, nil
}
