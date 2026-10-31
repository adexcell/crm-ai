package postgres

import (
	"context"
	"fmt"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
)

func (p *Postgres) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]domain.RefreshToken, error) {
	sql := `
	SELECT id, user_id, token_hash, user_agent, ip_address, expires_at, is_revoked, created_at
	FROM refresh_tokens
	WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
	ORDER BY created_at DESC`

	rows, err := p.db.Executor(ctx).Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("querying active sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.RefreshToken
	for rows.Next() {
		var t domain.RefreshToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.TokenHash, &t.UserAgent, &t.IPAddress, &t.ExpiresAt, &t.IsRevoked, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}
		sessions = append(sessions, t)
	}
	return sessions, nil
}

func (p *Postgres) RevokeUserSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	sql := `
	UPDATE refresh_tokens
	SET is_revoked = true
	WHERE id = $1 AND user_id = $2`

	_, err := p.db.Executor(ctx).Exec(ctx, sql, sessionID, userID)
	if err != nil {
		return fmt.Errorf("executing update refresh token query: %w", err)
	}
	return nil
}
