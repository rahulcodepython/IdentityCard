package services

import (
	"context"

	"github.com/google/uuid"

	"identitycard-server/internal/entities"
	"identitycard-server/internal/repositories"
)

type MembersService struct {
	repo *repositories.MembersRepository
}

func NewMembersService(repo *repositories.MembersRepository) *MembersService {
	return &MembersService{repo: repo}
}

func (s *MembersService) ListOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]entities.OrganizationMemberResponse, error) {
	rows, err := s.repo.ListOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var members []entities.OrganizationMemberResponse
	for _, row := range rows {
		members = append(members, entities.OrganizationMemberResponse{
			ID:             row.ID,
			UserID:         row.UserID,
			Name:           row.Name,
			Email:          row.Email,
			Roles:          row.Roles,
			CreatedAt:      row.CreatedAt.Time,
			Status:         "Active", // MOCKED
			AssignedEvents: []string{}, // MOCKED
		})
	}
	return members, nil
}

func (s *MembersService) DeleteMember(ctx context.Context, orgID, memberID uuid.UUID) error {
	return s.repo.Delete(ctx, orgID, memberID)
}
