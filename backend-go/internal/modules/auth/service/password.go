package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/crypto"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

const ResetPasswordTTL = 15 * time.Minute

func (s *Service) ForgotPassword(ctx context.Context, input dto.ForgotPasswordRequest) error {
	rawToken, err := crypto.GenerateRandomToken()
	if err != nil {
		return fmt.Errorf("generating secure token: %w", err)
	}

	tokenHash := crypto.HashToken(rawToken)

	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("finding user by email: %w", err)
	}
	if user == nil {
		return nil
	}

	err = s.cashe.SetPasswordResetToken(ctx, tokenHash, user.ID, ResetPasswordTTL)
	if err != nil {
		return fmt.Errorf("setting password reset token: %w", err)
	}

	slog.InfoContext(ctx, "mock send email", "email", input.Email, "reset_token", rawToken)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, input dto.ResetPasswordRequest) error {
	tokenHash := crypto.HashToken(input.Token)
	userID, err := s.cashe.GetAndDeletePasswordResetToken(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("fetching and deleting password reset token: %w", err)
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("finding user by id: %w", err)
	}
	if user == nil {
		return domain.ErrNotFound
	}

	passwordHash, err := crypto.HashPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.UpdatePassword(txCtx, userID, passwordHash); err != nil {
			return fmt.Errorf("updating password: %w", err)
		}
		if err := s.repo.RevokeAllUserTokens(txCtx, userID); err != nil {
			return fmt.Errorf("revoking all user tokens: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("executing transaction: %w", err)
	}

	return nil
}
