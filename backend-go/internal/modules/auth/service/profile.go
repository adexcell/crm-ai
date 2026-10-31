package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
	"github.com/google/uuid"
)

func (s *Service) GetProfile(ctx context.Context, userID, currentOrgID uuid.UUID) (dto.ProfileResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.ProfileResponse{}, fmt.Errorf("finding user by id: %w", err)
	}
	if user == nil {
		return dto.ProfileResponse{}, domain.ErrNotFound
	}
	if !user.IsActive {
		return dto.ProfileResponse{}, domain.ErrUserInactive
	}

	memberships, err := s.repo.GetUserMemberships(ctx, userID)
	if err != nil {
		return dto.ProfileResponse{}, fmt.Errorf("finding user memberships: %w", err)
	}

	membershipsDTO := make([]dto.UserMembershipResponse, 0, len(memberships))
	var currentOrg *dto.UserMembershipResponse

	for _, m := range memberships {
		mDTO := dto.UserMembershipResponse{
			OrganizationID:   m.OrganizationID,
			OrganizationName: m.OrganizationName,
			Slug:             m.Slug,
			Role:             m.Role,
		}
		membershipsDTO = append(membershipsDTO, mDTO)
		if currentOrgID != uuid.Nil && m.OrganizationID == currentOrgID {
			currentOrg = &membershipsDTO[len(membershipsDTO)-1]
		}
	}

	return dto.ProfileResponse{
		User: dto.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			IsActive:    user.IsActive,
			IsSuperuser: user.IsSuperuser,
			CreatedAt:   user.CreatedAt,
		},
		CurrentOrganization: currentOrg,
		Memberships:         membershipsDTO,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, input dto.UpdateProfileRequest) (dto.UserResponse, error) {
	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)
	if firstName == "" || lastName == "" {
		return dto.UserResponse{}, domain.ErrInvalidInput
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.UserResponse{}, fmt.Errorf("finding user by id: %w", err)
	}
	if user == nil {
		return dto.UserResponse{}, domain.ErrNotFound
	}
	if !user.IsActive {
		return dto.UserResponse{}, domain.ErrUserInactive
	}

	if err := s.repo.UpdateUserProfile(ctx, userID, firstName, lastName); err != nil {
		return dto.UserResponse{}, fmt.Errorf("updating user profile: %w", err)
	}

	return dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   firstName,
		LastName:    lastName,
		IsActive:    user.IsActive,
		IsSuperuser: user.IsSuperuser,
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (s *Service) SwitchOrganization(ctx context.Context, userID, targetOrgID uuid.UUID) (dto.TokenResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.TokenResponse{}, fmt.Errorf("finding user by id: %w", err)
	}
	if user == nil {
		return dto.TokenResponse{}, domain.ErrNotFound
	}
	if !user.IsActive {
		return dto.TokenResponse{}, domain.ErrUserInactive
	}

	member, err := s.repo.GetMemberByOrgAndUser(ctx, targetOrgID, userID)
	if err != nil {
		return dto.TokenResponse{}, fmt.Errorf("finding membership by org and user: %w", err)
	}
	if member == nil {
		return dto.TokenResponse{}, domain.ErrForbidden
	}

	tokenResp, err := s.issueAccessToken(user.ID, targetOrgID, user.Email, string(member.Role))
	if err != nil {
		return dto.TokenResponse{}, fmt.Errorf("issuing access token: %w", err)
	}

	return tokenResp, nil
}
