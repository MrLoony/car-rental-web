package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrBookingNotFound = errors.New("booking not found")

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CreateBooking(ctx context.Context, booking model.Booking) (int64, error) {
	const query = `
		INSERT INTO bookings (
			car_id,
			customer_name,
			customer_email,
			customer_phone,
			pickup_at,
			return_at,
			billing_days,
			estimated_total,
			message,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		ctx,
		query,
		booking.CarID,
		booking.CustomerName,
		booking.CustomerEmail,
		booking.CustomerPhone,
		booking.PickupAt,
		booking.ReturnAt,
		booking.BillingDays,
		booking.EstimatedTotal,
		booking.Message,
		booking.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create booking: %w", err)
	}

	return id, nil
}

func (r *BookingRepository) HasBookingConflict(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM bookings b
			WHERE b.car_id = $1
				AND b.status IN ($2, $3)
				AND $4 < b.return_at + ($6 * interval '1 hour')
				AND $5 > b.pickup_at
		)
	`

	var hasConflict bool
	err := r.db.QueryRow(
		ctx,
		query,
		carID,
		model.BookingStatusPending,
		model.BookingStatusConfirmed,
		pickupAt,
		returnAt,
		bufferHours,
	).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("check booking conflict: %w", err)
	}

	return hasConflict, nil
}

func (r *BookingRepository) FindNextAvailablePickupAt(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (time.Time, bool, error) {
	const query = `
		SELECT MAX(b.return_at + ($6 * interval '1 hour'))
		FROM bookings b
		WHERE b.car_id = $1
			AND b.status IN ($2, $3)
			AND $4 < b.return_at + ($6 * interval '1 hour')
			AND $5 > b.pickup_at
	`

	var suggestedAt pgtype.Timestamptz
	err := r.db.QueryRow(
		ctx,
		query,
		carID,
		model.BookingStatusPending,
		model.BookingStatusConfirmed,
		pickupAt,
		returnAt,
		bufferHours,
	).Scan(&suggestedAt)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("find next available pickup time: %w", err)
	}

	if !suggestedAt.Valid {
		return time.Time{}, false, nil
	}

	return suggestedAt.Time, true, nil
}

func (r *BookingRepository) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	const query = `
		SELECT
			b.id,
			b.car_id,
			c.brand,
			c.model,
			c.slug,
			b.customer_name,
			b.customer_email,
			b.customer_phone,
			b.pickup_at,
			b.return_at,
			b.billing_days,
			b.estimated_total::double precision,
			COALESCE(b.message, '') AS message,
			b.status,
			b.created_at,
			b.updated_at
		FROM bookings b
		JOIN cars c ON c.id = b.car_id
		ORDER BY b.created_at DESC, b.id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]model.BookingAdminView, 0)
	for rows.Next() {
		var booking model.BookingAdminView
		if err := scanBookingAdminView(rows, &booking); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}

	return bookings, nil
}

func (r *BookingRepository) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	const query = `
		SELECT
			b.id,
			b.car_id,
			c.brand,
			c.model,
			c.slug,
			b.customer_name,
			b.customer_email,
			b.customer_phone,
			b.pickup_at,
			b.return_at,
			b.billing_days,
			b.estimated_total::double precision,
			COALESCE(b.message, '') AS message,
			b.status,
			b.created_at,
			b.updated_at
		FROM bookings b
		JOIN cars c ON c.id = b.car_id
		WHERE b.id = $1
	`

	var booking model.BookingAdminView
	err := scanBookingAdminView(r.db.QueryRow(ctx, query, id), &booking)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.BookingAdminView{}, fmt.Errorf("get booking by id %d: %w", id, ErrBookingNotFound)
		}

		return model.BookingAdminView{}, fmt.Errorf("get booking by id %d: %w", id, err)
	}

	return booking, nil
}

func (r *BookingRepository) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	const query = `
		UPDATE bookings
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("update booking status id %d: %w", id, ErrBookingNotFound)
	}

	return nil
}

type bookingAdminScanner interface {
	Scan(dest ...any) error
}

func scanBookingAdminView(scanner bookingAdminScanner, booking *model.BookingAdminView) error {
	return scanner.Scan(
		&booking.ID,
		&booking.CarID,
		&booking.CarBrand,
		&booking.CarModel,
		&booking.CarSlug,
		&booking.CustomerName,
		&booking.CustomerEmail,
		&booking.CustomerPhone,
		&booking.PickupAt,
		&booking.ReturnAt,
		&booking.BillingDays,
		&booking.EstimatedTotal,
		&booking.Message,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
}
