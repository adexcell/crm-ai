package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateOrganization(ctx context.Context, org *domain.Organization) error {
	sql := `
	INSERT INTO organizations (id, name, slug, is_active, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := p.db.Executor(ctx).Exec(
		ctx, sql, org.ID, org.Name, org.Slug,
		org.IsActive, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("executing insert organization query: %w", err)
	}

	return nil
}

func (p *Postgres) GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	sql := `
	SELECT id, name, slug, is_active, created_at
	FROM organizations
	WHERE slug = $1`

	var org domain.Organization
	err := p.db.Executor(ctx).QueryRow(ctx, sql, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.IsActive,
		&org.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("querying organization: %w", err)
	}

	return &org, nil
}
