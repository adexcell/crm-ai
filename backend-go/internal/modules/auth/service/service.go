// Package service
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/infra/crypto"
	"github.com/adexcell/crm-ai/internal/infra/jwt"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type AuthRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error
	GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]domain.UserMembership, error)

	CreateOrganization(ctx context.Context, org *domain.Organization) error
	GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error)

	CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]domain.RefreshToken, error)
	RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error

	GetMember(ctx context.Context, userID uuid.UUID) (*domain.OrganizationMember, error)
	AddMember(ctx context.Context, member *domain.OrganizationMember) error
	GetMemberByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error)
}

type AuthCashe interface {
	SetVerificationCode(ctx context.Context, email, code string, ttl time.Duration) error
	GetVerificationCode(ctx context.Context, email string) (string, error)
	DeleteVerificationCode(ctx context.Context, email string) error

	SetPasswordResetToken(ctx context.Context, tokenHash string, userID uuid.UUID, ttl time.Duration) error
	GetAndDeletePasswordResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error)
}

type Service struct {
	repo       AuthRepository
	cashe      AuthCashe
	txManager  TransactionManager
	jwtManager *jwt.Manager
	cfg        *config.Config
}

func (s *Service) issueRefreshToken(userID uuid.UUID, userAgent, clientIP string) (domain.RefreshToken, string, error) {
	rawToken, err := crypto.GenerateRandomToken()
	if err != nil {
		return domain.RefreshToken{}, "", fmt.Errorf("generating secure token: %w", err)
	}
	tokenHash := crypto.HashToken(rawToken)
	refreshExpiresAt := time.Now().Add(time.Duration(s.cfg.JWT.RefreshTokenExpireDays) * 24 * time.Hour)

	tokenEntity, err := domain.NewRefreshToken(
		userID,
		tokenHash,
		&userAgent,
		&clientIP,
		refreshExpiresAt,
	)
	if err != nil {
		return domain.RefreshToken{}, "", fmt.Errorf("creating new refresh token: %w", err)
	}
	return tokenEntity, rawToken, nil
}

func (s *Service) issueAccessToken(userID, orgID uuid.UUID, email, role string) (dto.TokenResponse, error) {
	accessDuration := s.cfg.JWT.AccessDuration()
	accessToken, err := s.jwtManager.Generate(
		userID,
		orgID,
		email,
		role,
		accessDuration,
	)
	if err != nil {
		return dto.TokenResponse{}, fmt.Errorf("generating access token: %w", err)
	}

	return dto.TokenResponse{
		AccessToken: accessToken,
		TokenType:   "bearer",
		ExpiresIn:   int64(accessDuration.Seconds()),
	}, nil
}

// resolveUserContext определяет активную организацию и роль пользователя
func (s *Service) resolveUserContext(ctx context.Context, userID uuid.UUID) (uuid.UUID, string, error) {
	member, err := s.repo.GetMember(ctx, userID)
	if err != nil && member == nil {
		return uuid.Nil, "", fmt.Errorf("finding organization member: %w", err)
	}

	// Если пользователь прикреплен к организации — возвращаем её ID и роль
	if member != nil {
		return member.OrganizationID, string(member.Role), nil
	}

	// Если организации нет (новый аккаунт или уволен) — возвращаем Nil UUID и пустую роль
	return uuid.Nil, "", nil
}

func New(
	repo AuthRepository,
	cashe AuthCashe,
	txManager TransactionManager,
	jwtManager *jwt.Manager,
	cfg *config.Config,
) *Service {
	return &Service{
		repo:       repo,
		cashe:      cashe,
		txManager:  txManager,
		jwtManager: jwtManager,
		cfg:        cfg,
	}
}
