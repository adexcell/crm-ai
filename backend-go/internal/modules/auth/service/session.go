package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/crypto"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
	"github.com/google/uuid"
)

// Фейковый предрассчитанный хеш `Argon2id` для симуляции проверки несуществующего юзера
const dummyHash = "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$RGF0YVNvbWVGYWtlSGFzaEZvclRpbWluZ1Byb3RlY3Rpb24"

func (s *Service) Login(ctx context.Context, input dto.LoginRequest, userAgent, clientIP string) (dto.AuthResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("finding user by email: %w", err)
	}

	hashToVerify := dummyHash
	if user != nil {
		hashToVerify = user.PasswordHash
	}

	match, err := crypto.VerifyPassword(input.Password, hashToVerify)
	if err != nil {
		slog.WarnContext(ctx, "verifying password", "err", err)
	}

	if user == nil || !match {
		return dto.AuthResponse{}, "", domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return dto.AuthResponse{}, "", domain.ErrUserInactive
	}

	orgID, role, err := s.resolveUserContext(ctx, user.ID)
	if err != nil {
		return dto.AuthResponse{}, "", err
	}

	tokenEntity, rawRefreshToken, err := s.issueRefreshToken(user.ID, userAgent, clientIP)
	if err != nil {
		return dto.AuthResponse{}, "", err
	}
	if err := s.repo.CreateRefreshToken(ctx, &tokenEntity); err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("creating refresh token: %w", err)
	}

	tokenResponse, err := s.issueAccessToken(user.ID, orgID, user.Email, role)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("issuing access token: %w", err)
	}

	// TODO: добавить проверку на подозрительный вход и отправку email с предупреждением

	return dto.AuthResponse{
		User: dto.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			IsActive:    user.IsActive,
			IsSuperuser: user.IsSuperuser,
			CreatedAt:   user.CreatedAt,
		},
		Tokens: tokenResponse,
	}, rawRefreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, rawRefreshToken, userAgent, clientIP string) (dto.TokenResponse, string, error) {
	tokenHash := crypto.HashToken(rawRefreshToken)

	refresh, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return dto.TokenResponse{}, "", fmt.Errorf("finding refresh token by hash: %w", err)
	}
	if refresh == nil {
		return dto.TokenResponse{}, "", domain.ErrInvalidToken
	}

	if refresh.IsRevoked {
		err := s.repo.RevokeAllUserTokens(ctx, refresh.UserID)
		if err != nil {
			slog.Warn("revoking all user tokens", "err", err)
		}
		return dto.TokenResponse{}, "", domain.ErrTokenReplayDetected
	}

	if time.Now().After(refresh.ExpiresAt) {
		return dto.TokenResponse{}, "", domain.ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, refresh.UserID)
	if err != nil {
		return dto.TokenResponse{}, "", fmt.Errorf("finding user by id: %w", err)
	}
	if !user.IsActive {
		return dto.TokenResponse{}, "", domain.ErrUserInactive
	}

	orgID, role, err := s.resolveUserContext(ctx, user.ID)
	if err != nil {
		return dto.TokenResponse{}, "", err
	}

	tokenEntity, rawRefreshToken, err := s.issueRefreshToken(user.ID, userAgent, clientIP)
	if err != nil {
		return dto.TokenResponse{}, "", err
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.RevokeRefreshToken(txCtx, refresh.ID); err != nil {
			return fmt.Errorf("revoking refresh token: %w", err)
		}
		if err := s.repo.CreateRefreshToken(txCtx, &tokenEntity); err != nil {
			return fmt.Errorf("storing refresh token: %w", err)
		}
		return nil
	})
	if err != nil {
		return dto.TokenResponse{}, "", err
	}

	tokenResponse, err := s.issueAccessToken(user.ID, orgID, user.Email, role)
	if err != nil {
		return dto.TokenResponse{}, "", fmt.Errorf("issuing access token: %w", err)
	}

	return tokenResponse, rawRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}

	tokenHash := crypto.HashToken(rawRefreshToken)
	token, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil || token == nil {
		return nil
	}

	return s.repo.RevokeRefreshToken(ctx, token.ID)
}

func (s *Service) GetSessions(ctx context.Context, userID uuid.UUID, rawRefreshToken string) ([]dto.SessionResponse, error) {
	var currentTokenHash string
	if rawRefreshToken != "" {
		currentTokenHash = crypto.HashToken(rawRefreshToken)
	}

	sessions, err := s.repo.GetActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("finding active sessions: %w", err)
	}

	result := make([]dto.SessionResponse, 0, len(sessions))
	for _, sess := range sessions {
		userAgent := ""
		if sess.UserAgent != nil {
			userAgent = *sess.UserAgent
		}
		ipAddress := ""
		if sess.IPAddress != nil {
			ipAddress = *sess.IPAddress
		}

		result = append(result, dto.SessionResponse{
			ID:        sess.ID,
			UserAgent: userAgent,
			IPAddress: ipAddress,
			CreatedAt: sess.CreatedAt,
			ExpiresAt: sess.ExpiresAt,
			IsCurrent: sess.TokenHash == currentTokenHash,
		})
	}

	return result, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	return s.repo.RevokeUserSession(ctx, userID, sessionID)
}

func (s *Service) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeAllUserTokens(ctx, userID)
}
