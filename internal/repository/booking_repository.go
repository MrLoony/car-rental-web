package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func (r *BookingRepository) ListBlockingBookingsForCar(ctx context.Context, carID int64, from time.Time, to time.Time) ([]model.Booking, error) {
	const query = `
		SELECT
			b.id,
			b.car_id,
			b.pickup_at,
			b.return_at,
			b.status
		FROM bookings b
		WHERE b.car_id = $1
			AND b.status IN ($2, $3)
			AND b.pickup_at < $4
			AND b.return_at > $5
		ORDER BY b.pickup_at ASC, b.id ASC
	`

	rows, err := r.db.Query(ctx, query, carID, model.BookingStatusPending, model.BookingStatusConfirmed, to, from)
	if err != nil {
		return nil, fmt.Errorf("list blocking bookings for car %d: %w", carID, err)
	}
	defer rows.Close()

	bookings := make([]model.Booking, 0)
	for rows.Next() {
		var booking model.Booking
		if err := rows.Scan(
			&booking.ID,
			&booking.CarID,
			&booking.PickupAt,
			&booking.ReturnAt,
			&booking.Status,
		); err != nil {
			return nil, fmt.Errorf("scan blocking booking for car %d: %w", carID, err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate blocking bookings for car %d: %w", carID, err)
	}

	return bookings, nil
}

func (r *BookingRepository) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	const query = `
		SELECT
			b.id,
			b.car_id,
			c.brand,
			c.model,
			c.slug,
			c.year,
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

func (r *BookingRepository) CountBookings(ctx context.Context, filter model.AdminBookingFilter) (int, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT COUNT(*)
		FROM bookings b
		JOIN cars c ON c.id = b.car_id
		WHERE TRUE
	`)

	args := make([]any, 0)
	appendAdminBookingFilters(&query, &args, filter)

	var count int
	if err := r.db.QueryRow(ctx, query.String(), args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count bookings: %w", err)
	}

	return count, nil
}

func (r *BookingRepository) ListBookingsPage(ctx context.Context, filter model.AdminBookingFilter, pagination model.Pagination) ([]model.BookingAdminView, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT
			b.id,
			b.car_id,
			c.brand,
			c.model,
			c.slug,
			c.year,
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
		WHERE TRUE
	`)

	args := make([]any, 0)
	appendAdminBookingFilters(&query, &args, filter)
	query.WriteString(" ORDER BY b.created_at DESC, b.id DESC")

	args = append(args, pagination.PerPage, pagination.Offset)
	limitPlaceholder := len(args) - 1
	offsetPlaceholder := len(args)
	query.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", limitPlaceholder, offsetPlaceholder))

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list bookings page: %w", err)
	}
	defer rows.Close()

	bookings := make([]model.BookingAdminView, 0)
	for rows.Next() {
		var booking model.BookingAdminView
		if err := scanBookingAdminView(rows, &booking); err != nil {
			return nil, fmt.Errorf("scan booking page: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings page: %w", err)
	}

	return bookings, nil
}

func (r *BookingRepository) ListBookingsForExport(ctx context.Context, filter model.AdminBookingFilter) ([]model.BookingExportRow, error) {
	var query strings.Builder
	query.WriteString(`
		SELECT
			b.id,
			b.status,
			b.customer_name,
			b.customer_email,
			b.customer_phone,
			CONCAT(c.brand, ' ', c.model) AS car,
			b.pickup_at,
			b.return_at,
			b.billing_days,
			b.estimated_total::double precision,
			b.created_at
		FROM bookings b
		JOIN cars c ON c.id = b.car_id
		WHERE TRUE
	`)

	args := make([]any, 0)
	appendAdminBookingFilters(&query, &args, filter)
	query.WriteString(" ORDER BY b.created_at DESC, b.id DESC")

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list bookings for export: %w", err)
	}
	defer rows.Close()

	bookings := make([]model.BookingExportRow, 0)
	for rows.Next() {
		var booking model.BookingExportRow
		if err := rows.Scan(
			&booking.ID,
			&booking.Status,
			&booking.CustomerName,
			&booking.CustomerEmail,
			&booking.CustomerPhone,
			&booking.Car,
			&booking.PickupAt,
			&booking.ReturnAt,
			&booking.BillingDays,
			&booking.EstimatedTotal,
			&booking.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan booking export row: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking export rows: %w", err)
	}

	return bookings, nil
}

func (r *BookingRepository) GetBookingStats(ctx context.Context) (model.BookingStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = $1) AS pending,
			COUNT(*) FILTER (WHERE status = $2) AS confirmed,
			COUNT(*) FILTER (WHERE status = $3) AS cancelled,
			COUNT(*) FILTER (WHERE status = $4) AS completed
		FROM bookings
	`

	var stats model.BookingStats
	if err := r.db.QueryRow(
		ctx,
		query,
		model.BookingStatusPending,
		model.BookingStatusConfirmed,
		model.BookingStatusCancelled,
		model.BookingStatusCompleted,
	).Scan(
		&stats.Total,
		&stats.Pending,
		&stats.Confirmed,
		&stats.Cancelled,
		&stats.Completed,
	); err != nil {
		return model.BookingStats{}, fmt.Errorf("get booking stats: %w", err)
	}

	return stats, nil
}

func (r *BookingRepository) GetRevenueStats(ctx context.Context) (model.RevenueStats, error) {
	const query = `
		SELECT
			COALESCE(SUM(estimated_total), 0)::double precision AS total_potential,
			COALESCE(SUM(estimated_total) FILTER (WHERE status = $1), 0)::double precision AS pending,
			COALESCE(SUM(estimated_total) FILTER (WHERE status = $2), 0)::double precision AS confirmed,
			COALESCE(SUM(estimated_total) FILTER (WHERE status = $3), 0)::double precision AS completed,
			COALESCE(SUM(estimated_total) FILTER (WHERE status = $4), 0)::double precision AS cancelled
		FROM bookings
	`

	var stats model.RevenueStats
	if err := r.db.QueryRow(
		ctx,
		query,
		model.BookingStatusPending,
		model.BookingStatusConfirmed,
		model.BookingStatusCompleted,
		model.BookingStatusCancelled,
	).Scan(
		&stats.TotalPotential,
		&stats.Pending,
		&stats.Confirmed,
		&stats.Completed,
		&stats.Cancelled,
	); err != nil {
		return model.RevenueStats{}, fmt.Errorf("get revenue stats: %w", err)
	}

	return stats, nil
}

func (r *BookingRepository) GetRecentBookings(ctx context.Context, limit int) ([]model.RecentBookingActivity, error) {
	if limit <= 0 {
		limit = 10
	}

	const query = `
		SELECT
			b.id,
			b.customer_name,
			CONCAT(c.brand, ' ', c.model) AS car_name,
			b.status,
			b.pickup_at,
			b.return_at,
			b.created_at
		FROM bookings b
		JOIN cars c ON c.id = b.car_id
		ORDER BY b.created_at DESC, b.id DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent bookings: %w", err)
	}
	defer rows.Close()

	activities := make([]model.RecentBookingActivity, 0)
	for rows.Next() {
		var activity model.RecentBookingActivity
		if err := rows.Scan(
			&activity.ID,
			&activity.CustomerName,
			&activity.CarName,
			&activity.Status,
			&activity.PickupTime,
			&activity.ReturnTime,
			&activity.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent booking: %w", err)
		}

		activities = append(activities, activity)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent bookings: %w", err)
	}

	return activities, nil
}

func (r *BookingRepository) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	const query = `
		SELECT
			b.id,
			b.car_id,
			c.brand,
			c.model,
			c.slug,
			c.year,
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
		&booking.CarYear,
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

func appendAdminBookingFilters(query *strings.Builder, args *[]any, filter model.AdminBookingFilter) {
	if filter.Search != "" {
		*args = append(*args, "%"+filter.Search+"%")
		placeholder := len(*args)
		query.WriteString(" AND ")
		query.WriteString(fmt.Sprintf(`(
			b.customer_name ILIKE $%d OR
			b.customer_email ILIKE $%d OR
			b.customer_phone ILIKE $%d OR
			c.brand ILIKE $%d OR
			c.model ILIKE $%d OR
			c.slug ILIKE $%d
		)`, placeholder, placeholder, placeholder, placeholder, placeholder, placeholder))
	}

	status := model.NormalizeAdminBookingStatus(filter.Status)
	if status != model.AdminBookingStatusAll {
		*args = append(*args, status)
		query.WriteString(" AND ")
		query.WriteString(fmt.Sprintf("b.status = $%d", len(*args)))
	}
}
