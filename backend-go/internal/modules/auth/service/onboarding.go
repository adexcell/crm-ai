package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/adexcell/crm-ai/internal/infra/crypto"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/adexcell/crm-ai/internal/modules/auth/dto"
)

const (
	OTPCodeLength = 6
	OTPTTL        = 5 * time.Minute
)

func generateSlug(s string) string {
	translitMap := map[rune]string{
		'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
		'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
		'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
		'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "sch",
		'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	}

	var b strings.Builder
	b.Grow(len(s))

	prevDash := false
	for _, r := range strings.ToLower(s) {
		// Кириллица -> транслит
		if tr, ok := translitMap[r]; ok {
			b.WriteString(tr)
			prevDash = false
			continue
		}

		// Латиница и цифры
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}

		// Любые другие символы -> дефис (но без повторов)
		if !prevDash && b.Len() > 0 {
			b.WriteRune('-')
			prevDash = true
		}
	}

	slug := strings.Trim(b.String(), "-")
	return slug
}

// randomSuffix генерирует короткую случайную строку (например, "a1b2")
func randomSuffix() (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) Register(ctx context.Context, input dto.RegisterRequest, userAgent, clientIP string) (dto.AuthResponse, string, error) {
	existingUser, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("finding user by email: %w", err)
	}
	if existingUser != nil {
		return dto.AuthResponse{}, "", domain.ErrUserAlreadyExists
	}

	pwdHash, err := crypto.HashPassword(input.Password)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("hashing password: %w", err)
	}

	user, err := domain.NewUser(
		input.Email,
		pwdHash,
		input.FirstName,
		input.LastName,
	)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("creatig user domain: %w", err)
	}

	slug := generateSlug(input.OrganizationName)

	existingOrg, err := s.repo.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("finding organization by slug: %w", err)
	}
	if existingOrg != nil {
		suffix, err := randomSuffix()
		if err != nil {
			return dto.AuthResponse{}, "", fmt.Errorf("generating random suffix for slug: %w", err)
		}
		slug = fmt.Sprintf("%s-%s", slug, suffix)
	}

	org, err := domain.NewOrganization(input.OrganizationName, slug)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("creating org domain: %w", err)
	}

	member, err := domain.NewOrganizationMember(org.ID, user.ID, domain.RoleOwner)
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("creating org member domain: %w", err)
	}

	tokenEntity, rawRefreshToken, err := s.issueRefreshToken(user.ID, userAgent, clientIP)
	if err != nil {
		return dto.AuthResponse{}, "", err
	}

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateUser(txCtx, &user); err != nil {
			return fmt.Errorf("storing new user: %w", err)
		}

		if err := s.repo.CreateOrganization(txCtx, &org); err != nil {
			return fmt.Errorf("storing org: %w", err)
		}

		if err := s.repo.AddMember(txCtx, &member); err != nil {
			return fmt.Errorf("adding membership: %w", err)
		}

		if err := s.repo.CreateRefreshToken(txCtx, &tokenEntity); err != nil {
			return fmt.Errorf("storing refresh token: %w", err)
		}

		return nil
	})
	if err != nil {
		return dto.AuthResponse{}, "", err
	}

	tokenResponse, err := s.issueAccessToken(user.ID, org.ID, user.Email, string(member.Role))
	if err != nil {
		return dto.AuthResponse{}, "", fmt.Errorf("issuing access token: %w", err)
	}

	return dto.AuthResponse{
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
		},
		Tokens: tokenResponse,
	}, rawRefreshToken, nil
}

func (s *Service) EmailSendCode(ctx context.Context, input dto.EmailSendCodeRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return fmt.Errorf("finding user by email: %w", err)
	}
	if user != nil {
		return domain.ErrUserAlreadyExists
	}

	code, err := crypto.GenerateOTP(OTPCodeLength)
	if err != nil {
		return fmt.Errorf("generating OTP code: %w", err)
	}

	if err := s.cashe.SetVerificationCode(ctx, input.Email, code, OTPTTL); err != nil {
		return fmt.Errorf("cashing verification code: %w", err)
	}

	// 4. Отправка письма (пока логируем в консоль для удобства тестирования)
	// В будущем здесь будет публикация события в RabbitMQ
	fmt.Printf("\n📧 [EMAIL MOCK] Verification Code for %s: >>> %s <<<\n\n", input.Email, code)
	return nil
}

func (s *Service) EmailVerifyCode(ctx context.Context, input dto.EmailVerifyCodeRequest) error {
	savedCode, err := s.cashe.GetVerificationCode(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrCodeNotFound) {
			return domain.ErrCodeNotFound
		}
		return fmt.Errorf("fiding verification code from cache: %w", err)
	}

	if savedCode != input.Code {
		return domain.ErrInvalidCode
	}

	if err := s.cashe.DeleteVerificationCode(ctx, input.Email); err != nil {
		slog.Warn("deleting verification code from cache", "err", err)
	}

	return nil
}
