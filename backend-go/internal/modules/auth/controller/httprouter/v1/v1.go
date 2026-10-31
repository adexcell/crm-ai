// Package v1 Handler
package v1

import (
	"context"

	"github.com/adexcell/crm-ai/internal/config"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
	"github.com/google/uuid"
)

type Service interface {
	EmailSendCode(ctx context.Context, input dto.EmailSendCodeRequest) error
	EmailVerifyCode(ctx context.Context, input dto.EmailVerifyCodeRequest) error
	Register(ctx context.Context, input dto.RegisterRequest, userAgent, clientIP string) (dto.AuthResponse, string, error)

	Login(ctx context.Context, input dto.LoginRequest, userAgent, clientIP string) (dto.AuthResponse, string, error)
	RefreshToken(ctx context.Context, rawRefreshToken, userAgent, clientIP string) (dto.TokenResponse, string, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	GetSessions(ctx context.Context, userID uuid.UUID, rawRefreshToken string) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error

	ForgotPassword(ctx context.Context, input dto.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, input dto.ResetPasswordRequest) error
	GetProfile(ctx context.Context, userID, currentOrgID uuid.UUID) (dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, input dto.UpdateProfileRequest) (dto.UserResponse, error)
	SwitchOrganization(ctx context.Context, userID, targetOrgID uuid.UUID) (dto.TokenResponse, error)
}

type Handler struct {
	service Service
	cfg     *config.Config
}

func New(service Service, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		cfg:     cfg,
	}
}
