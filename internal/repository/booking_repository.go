package repository

import (
	"context"
	"fmt"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
