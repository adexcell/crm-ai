package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/adexcell/crm-ai/internal/modules/auth/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Postgres) CreateUser(ctx context.Context, user *domain.User) error {
	sql := `INSERT INTO users (
	id, email, password_hash,
	first_name, last_name, is_active,
	is_superuser, created_at, updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := p.db.Executor(ctx).Exec(
		ctx,
		sql,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.IsSuperuser,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("executing insert user query: %w", err)
	}
	return nil
}

func (p *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	sql := `
	SELECT id, email, password_hash,
		   first_name, last_name, is_active,
		   is_superuser, created_at, updated_at
	FROM users
	WHERE email = $1`

	var u domain.User
	err := p.db.Executor(ctx).QueryRow(ctx, sql, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName,
		&u.IsActive, &u.IsSuperuser, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning user query: %w", err)
	}
	return &u, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	sql := `
	SELECT id, email, password_hash,
		   first_name, last_name, is_active,
		   is_superuser, created_at, updated_at
	FROM users
	WHERE id = $1`

	var u domain.User
	err := p.db.Executor(ctx).QueryRow(ctx, sql, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName,
		&u.IsActive, &u.IsSuperuser, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scanning user query: %w", err)
	}
	return &u, nil
}

func (p *Postgres) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	sql := `
	UPDATE users
	SET password_hash = $1, updated_at = NOW()
	WHERE id = $2`

	_, err := p.db.Executor(ctx).Exec(ctx, sql, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("executing update password query: %w", err)
	}

	return nil
}

func (p *Postgres) UpdateUserProfile(ctx context.Context, userID uuid.UUID, firstName, lastName string) error {
	sql := `
	UPDATE users
	SET first_name = $1, last_name = $2, updated_at = NOW()
	WHERE id = $3`

	_, err := p.db.Executor(ctx).Exec(ctx, sql, firstName, lastName, userID)
	if err != nil {
		return fmt.Errorf("executing update profile query: %w", err)
	}

	return nil
}
