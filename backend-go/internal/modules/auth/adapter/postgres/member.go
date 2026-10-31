package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) GetMember(ctx context.Context, userID uuid.UUID) (*domain.OrganizationMember, error) {
	sql := `
	SELECT id, organization_id, user_id, role, joined_at
	FROM organization_members
	WHERE user_id = $1
	ORDER BY joined_at ASC
	LIMIT 1`

	var m domain.OrganizationMember
	err := p.db.Executor(ctx).QueryRow(ctx, sql, userID).Scan(
		&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning organization member: %w", err)
	}
	return &m, nil
}

func (p *Postgres) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	sql := `
	INSERT INTO organization_members (id, organization_id, user_id, role, joined_at)
	VALUES ($1, $2, $3, $4, $5)`

	_, err := p.db.Executor(ctx).Exec(
		ctx, sql, member.ID, member.OrganizationID,
		member.UserID, member.Role, member.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("executing insert orgagnization member: %w", err)
	}

	return nil
}

func (p *Postgres) GetMemberByOrgAndUser(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	sql := `
	SELECT id, organization_id, user_id, role, joined_at
	FROM organization_members
	WHERE organization_id = $1 AND user_id = $2`

	var member domain.OrganizationMember
	err := p.db.Executor(ctx).QueryRow(ctx, sql, orgID, userID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying organization members: %w", err)
	}

	return &member, nil
}
