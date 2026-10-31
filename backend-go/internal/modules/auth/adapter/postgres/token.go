package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	sql := `
	INSERT INTO refresh_tokens (
		id, user_id, token_hash,
		user_agent, ip_address, expires_at,
		is_revoked, created_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := p.db.Executor(ctx).Exec(
		ctx, sql, token.ID, token.UserID, token.TokenHash,
		token.UserAgent, token.IPAddress, token.ExpiresAt,
		token.IsRevoked, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("executing insert refresh token query: %w", err)
	}

	return nil
}

func (p *Postgres) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	sql := `
	UPDATE refresh_tokens
	SET is_revoked = true
	WHERE id = $1`

	_, err := p.db.Executor(ctx).Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("executing update refresh token query: %w", err)
	}

	return nil
}

func (p *Postgres) GetRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	sql := `
	SELECT id, user_id, token_hash,
		   user_agent, ip_address, expires_at,
		   is_revoked, created_at
	FROM refresh_tokens
	WHERE token_hash = $1`

	var t domain.RefreshToken
	err := p.db.Executor(ctx).QueryRow(ctx, sql, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash,
		&t.UserAgent, &t.IPAddress, &t.ExpiresAt,
		&t.IsRevoked, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning refresh token row: %w", err)
	}
	return &t, nil
}

func (p *Postgres) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	sql := `
	UPDATE refresh_tokens
	SET is_revoked = true
	WHERE user_id = $1 AND is_revoked = false`

	_, err := p.db.Executor(ctx).Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("executing update token query: %w", err)
	}

	return nil
}
