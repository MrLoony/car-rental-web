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

func (r *CarRepository) ListCarsForAdmin(ctx context.Context) ([]model.Car, error) {
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
		ORDER BY c.created_at DESC, c.id DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list cars for admin: %w", err)
	}
	defer rows.Close()

	cars := make([]model.Car, 0)
	for rows.Next() {
		var car model.Car
		if err := scanCar(rows, &car); err != nil {
			return nil, fmt.Errorf("scan admin car: %w", err)
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin cars: %w", err)
	}

	return cars, nil
}

func (r *CarRepository) GetCarByID(ctx context.Context, id int64) (model.Car, error) {
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
		WHERE c.id = $1
	`

	var car model.Car
	err := scanCar(r.db.QueryRow(ctx, query, id), &car)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Car{}, fmt.Errorf("get car by id %d: %w", id, ErrCarNotFound)
		}

		return model.Car{}, fmt.Errorf("get car by id %d: %w", id, err)
	}

	return car, nil
}

func (r *CarRepository) CreateCar(ctx context.Context, car model.Car) (int64, error) {
	const query = `
		INSERT INTO cars (
			category_id,
			brand,
			model,
			slug,
			year,
			price_per_day,
			transmission,
			fuel_type,
			seats,
			image_url,
			is_available
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		ctx,
		query,
		car.CategoryID,
		car.Brand,
		car.Model,
		car.Slug,
		car.Year,
		car.PricePerDay,
		car.Transmission,
		car.FuelType,
		car.Seats,
		car.ImageURL,
		car.IsAvailable,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create car: %w", err)
	}

	return id, nil
}

func (r *CarRepository) UpdateCar(ctx context.Context, car model.Car) error {
	const query = `
		UPDATE cars
		SET
			category_id = $1,
			brand = $2,
			model = $3,
			slug = $4,
			year = $5,
			price_per_day = $6,
			transmission = $7,
			fuel_type = $8,
			seats = $9,
			image_url = NULLIF($10, ''),
			is_available = $11,
			updated_at = NOW()
		WHERE id = $12
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		car.CategoryID,
		car.Brand,
		car.Model,
		car.Slug,
		car.Year,
		car.PricePerDay,
		car.Transmission,
		car.FuelType,
		car.Seats,
		car.ImageURL,
		car.IsAvailable,
		car.ID,
	)
	if err != nil {
		return fmt.Errorf("update car %d: %w", car.ID, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update car %d: %w", car.ID, ErrCarNotFound)
	}

	return nil
}

func (r *CarRepository) UpdateCarAvailability(ctx context.Context, id int64, isAvailable bool) error {
	const query = `
		UPDATE cars
		SET is_available = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	tag, err := r.db.Exec(ctx, query, isAvailable, id)
	if err != nil {
		return fmt.Errorf("update car availability %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update car availability %d: %w", id, ErrCarNotFound)
	}

	return nil
}

func (r *CarRepository) CarSlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM cars
			WHERE slug = $1
				AND ($2::bigint = 0 OR id <> $2)
		)
	`

	var exists bool
	if err := r.db.QueryRow(ctx, query, slug, excludeID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check car slug %q: %w", slug, err)
	}

	return exists, nil
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
