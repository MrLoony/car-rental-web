package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	return r.ListCars(ctx, model.CarFilter{Sort: model.SortNewest})
}

func (r *CarRepository) ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {
	var query strings.Builder
	query.WriteString(`
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
	`)

	args := make([]any, 0)

	addFilter := func(condition string, value any) {
		args = append(args, value)
		query.WriteString(" AND ")
		query.WriteString(fmt.Sprintf(condition, len(args)))
	}

	if filter.Search != "" {
		args = append(args, "%"+filter.Search+"%")
		placeholder := len(args)
		query.WriteString(" AND ")
		query.WriteString(fmt.Sprintf("(c.brand ILIKE $%d OR c.model ILIKE $%d)", placeholder, placeholder))
	}

	if filter.CategorySlug != "" {
		addFilter("cc.slug = $%d", filter.CategorySlug)
	}

	if filter.FuelType != "" {
		addFilter("c.fuel_type = $%d", filter.FuelType)
	}

	if filter.Transmission != "" {
		addFilter("c.transmission = $%d", filter.Transmission)
	}

	switch model.NormalizeCarSort(filter.Sort) {
	case model.SortPriceAsc:
		query.WriteString(" ORDER BY c.price_per_day ASC, c.id DESC")
	case model.SortPriceDesc:
		query.WriteString(" ORDER BY c.price_per_day DESC, c.id DESC")
	default:
		query.WriteString(" ORDER BY c.created_at DESC, c.id DESC")
	}

	rows, err := r.db.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list cars: %w", err)
	}
	defer rows.Close()

	cars := make([]model.Car, 0)
	for rows.Next() {
		var car model.Car
		if err := scanCar(rows, &car); err != nil {
			return nil, fmt.Errorf("scan car: %w", err)
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cars: %w", err)
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
