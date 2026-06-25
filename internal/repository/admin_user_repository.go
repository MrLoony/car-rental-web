package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAdminUserNotFound = errors.New("admin user not found")

type AdminUserRepository struct {
	db *pgxpool.Pool
}

func NewAdminUserRepository(db *pgxpool.Pool) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) GetAdminUserByEmail(ctx context.Context, email string) (model.AdminUser, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			created_at,
			updated_at
		FROM admin_users
		WHERE email = $1
	`

	var user model.AdminUser
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.AdminUser{}, fmt.Errorf("get admin user by email %q: %w", email, ErrAdminUserNotFound)
		}

		return model.AdminUser{}, fmt.Errorf("get admin user by email %q: %w", email, err)
	}

	return user, nil
}

func (r *AdminUserRepository) UpsertAdminUserPasswordHash(ctx context.Context, email string, passwordHash string) error {
	const query = `
		INSERT INTO admin_users (email, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (email)
		DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			updated_at = NOW()
	`

	if _, err := r.db.Exec(ctx, query, email, passwordHash); err != nil {
		return fmt.Errorf("upsert admin user %q: %w", email, err)
	}

	return nil
}

func (r *AdminUserRepository) DeleteAdminUserByEmail(ctx context.Context, email string) error {
	const query = `DELETE FROM admin_users WHERE email = $1`

	if _, err := r.db.Exec(ctx, query, email); err != nil {
		return fmt.Errorf("delete admin user %q: %w", email, err)
	}

	return nil
}
