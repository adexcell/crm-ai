// Package redis
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/adexcell/crm-ai/internal/infra/redis"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
)

const (
	OTPprefix       = "auth:otp:%s"
	PassResetPrefix = "auth:pass:%s"
)

type Redis struct {
	client *redis.Client
}

func New(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) SetVerificationCode(ctx context.Context, email, code string, ttl time.Duration) error {
	key := fmt.Sprintf(OTPprefix, email)
	return r.client.Set(ctx, key, code, ttl).Err()
}

func (r *Redis) GetVerificationCode(ctx context.Context, email string) (string, error) {
	key := fmt.Sprintf(OTPprefix, email)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", domain.ErrCodeNotFound
		}
		return "", fmt.Errorf("getting verification code: %w", err)
	}

	return val, nil
}

func (r *Redis) DeleteVerificationCode(ctx context.Context, email string) error {
	key := fmt.Sprintf(OTPprefix, email)
	return r.client.Del(ctx, key).Err()
}

func (r *Redis) SetPasswordResetToken(ctx context.Context, tokenHash string, userID uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf(PassResetPrefix, tokenHash)
	return r.client.Set(ctx, key, userID.String(), ttl).Err()
}

func (r *Redis) GetAndDeletePasswordResetToken(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	key := fmt.Sprintf(PassResetPrefix, tokenHash)
	val, err := r.client.GetDel(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return uuid.Nil, domain.ErrInvalidToken
		}
		return uuid.Nil, fmt.Errorf("fetching and deleting password reset token: %w", err)
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		if len(val) == 16 {
			return uuid.FromBytes([]byte(val))
		}
		return uuid.Nil, fmt.Errorf("parsing uuid from redis token payload: %w", err)
	}

	return userID, nil
}
