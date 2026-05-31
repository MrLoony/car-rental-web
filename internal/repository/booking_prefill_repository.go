package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBookingPrefillNotFound = errors.New("booking prefill not found")

type BookingPrefillRepository struct {
	db *pgxpool.Pool
}

func NewBookingPrefillRepository(db *pgxpool.Pool) *BookingPrefillRepository {
	return &BookingPrefillRepository{db: db}
}

func (r *BookingPrefillRepository) CreateBookingPrefill(ctx context.Context, prefill model.BookingPrefill) error {
	const query = `
		INSERT INTO booking_prefills (
			token,
			name,
			email,
			phone,
			pickup_at,
			return_at,
			message,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		prefill.Token,
		prefill.Name,
		prefill.Email,
		prefill.Phone,
		prefill.PickupAt,
		prefill.ReturnAt,
		prefill.Message,
		prefill.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create booking prefill: %w", err)
	}

	return nil
}

func (r *BookingPrefillRepository) GetBookingPrefillByToken(ctx context.Context, token string) (model.BookingPrefill, error) {
	const query = `
		SELECT
			id,
			token,
			COALESCE(name, '') AS name,
			COALESCE(email, '') AS email,
			COALESCE(phone, '') AS phone,
			COALESCE(pickup_at, '') AS pickup_at,
			COALESCE(return_at, '') AS return_at,
			COALESCE(message, '') AS message,
			expires_at,
			created_at
		FROM booking_prefills
		WHERE token = $1
			AND expires_at > NOW()
	`

	var prefill model.BookingPrefill
	err := scanBookingPrefill(r.db.QueryRow(ctx, query, token), &prefill)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BookingPrefill{}, fmt.Errorf("get booking prefill by token: %w", ErrBookingPrefillNotFound)
		}

		return model.BookingPrefill{}, fmt.Errorf("get booking prefill by token: %w", err)
	}

	return prefill, nil
}

func (r *BookingPrefillRepository) DeleteExpiredBookingPrefills(ctx context.Context) error {
	const query = `
		DELETE FROM booking_prefills
		WHERE expires_at <= NOW()
	`

	if _, err := r.db.Exec(ctx, query); err != nil {
		return fmt.Errorf("delete expired booking prefills: %w", err)
	}

	return nil
}

type bookingPrefillScanner interface {
	Scan(dest ...any) error
}

func scanBookingPrefill(scanner bookingPrefillScanner, prefill *model.BookingPrefill) error {
	return scanner.Scan(
		&prefill.ID,
		&prefill.Token,
		&prefill.Name,
		&prefill.Email,
		&prefill.Phone,
		&prefill.PickupAt,
		&prefill.ReturnAt,
		&prefill.Message,
		&prefill.ExpiresAt,
		&prefill.CreatedAt,
	)
}
