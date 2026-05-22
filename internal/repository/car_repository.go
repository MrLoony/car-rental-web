package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCarNotFound = errors.New("car not found")

type CarRepository struct {
	db *pgxpool.Pool
}

func NewCarRepository(db *pgxpool.Pool) *CarRepository {
	return &CarRepository{db: db}
}

func (r *CarRepository) ListAvailableCars(ctx context.Context) ([]model.Car, error) {
	const query = `
		SELECT
			c.id,
			c.category_id,
			cc.name AS category_name,
			c.brand,
			c.model,
			c.slug,
			c.year,
			c.price_per_day::double precision,
			c.transmission,
			c.fuel_type,
			c.seats,
			COALESCE(c.image_url, '') AS image_url,
			c.is_available,
			c.created_at,
			c.updated_at
		FROM cars c
		JOIN car_categories cc ON cc.id = c.category_id
		WHERE c.is_available = TRUE
		ORDER BY c.created_at DESC, c.id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list available cars: %w", err)
	}
	defer rows.Close()

	cars := make([]model.Car, 0)
	for rows.Next() {
		var car model.Car
		if err := scanCar(rows, &car); err != nil {
			return nil, fmt.Errorf("scan available car: %w", err)
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available cars: %w", err)
	}

	return cars, nil
}

func (r *CarRepository) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	const query = `
		SELECT
			c.id,
			c.category_id,
			cc.name AS category_name,
			c.brand,
			c.model,
			c.slug,
			c.year,
			c.price_per_day::double precision,
			c.transmission,
			c.fuel_type,
			c.seats,
			COALESCE(c.image_url, '') AS image_url,
			c.is_available,
			c.created_at,
			c.updated_at
		FROM cars c
		JOIN car_categories cc ON cc.id = c.category_id
		WHERE c.slug = $1
	`

	var car model.Car
	err := scanCar(r.db.QueryRow(ctx, query, slug), &car)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Car{}, fmt.Errorf("get car by slug %q: %w", slug, ErrCarNotFound)
		}

		return model.Car{}, fmt.Errorf("get car by slug %q: %w", slug, err)
	}

	return car, nil
}

type carScanner interface {
	Scan(dest ...any) error
}

func scanCar(scanner carScanner, car *model.Car) error {
	return scanner.Scan(
		&car.ID,
		&car.CategoryID,
		&car.CategoryName,
		&car.Brand,
		&car.Model,
		&car.Slug,
		&car.Year,
		&car.PricePerDay,
		&car.Transmission,
		&car.FuelType,
		&car.Seats,
		&car.ImageURL,
		&car.IsAvailable,
		&car.CreatedAt,
		&car.UpdatedAt,
	)
}
