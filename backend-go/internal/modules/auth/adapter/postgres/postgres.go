// Package postgres
package postgres

import (
	"context"
	"fmt"

	"github.com/adexcell/crm-ai/internal/infra/postgres"
	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
)

type Postgres struct {
	db *postgres.Pool
}

func New(db *postgres.Pool) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) GetUserMemberships(ctx context.Context, userID uuid.UUID) ([]domain.UserMembership, error) {
	sql := `
	SELECT om.organization_id, o.name, o.slug, om.role
	FROM organization_members om
	JOIN organizations o ON o.id = om.organization_id
	WHERE om.user_id = $1 AND o.is_active = true
	ORDER BY om.joined_at ASC`

	rows, err := p.db.Executor(ctx).Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var userMemberships []domain.UserMembership
	for rows.Next() {
		var t domain.UserMembership
		if err := rows.Scan(
			&t.OrganizationID, &t.OrganizationName, &t.Slug, &t.Role,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		userMemberships = append(userMemberships, t)
	}
	return userMemberships, nil
}
