package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MrLoony/car-rental-web/internal/model"
)

type CarService struct {
	repo carRepository
}

var (
	ErrInvalidCarImage  = errors.New("invalid car image")
	ErrCarImageNotFound = errors.New("car image not found")
)

type carRepository interface {
	ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error)
	CountCars(ctx context.Context, filter model.CarFilter) (int, error)
	ListCarsPage(ctx context.Context, filter model.CarFilter, pagination model.Pagination) ([]model.Car, error)
	GetCarBySlug(ctx context.Context, slug string) (model.Car, error)
	GetCarImagesByCarID(ctx context.Context, carID int64) ([]model.CarImage, error)
	GetCatalogImageURLsByCarIDs(ctx context.Context, carIDs []int64) (map[int64]string, error)
	GetCarImageByID(ctx context.Context, imageID int64) (model.CarImage, error)
	CreateCarImage(ctx context.Context, image model.CarImage) (int64, error)
	DeleteCarImage(ctx context.Context, imageID int64) error
	SetPrimaryCarImage(ctx context.Context, carID, imageID int64) error
	ListCarsForAdmin(ctx context.Context) ([]model.Car, error)
	CountCarsForAdmin(ctx context.Context, filter model.AdminCarFilter) (int, error)
	ListCarsForAdminPage(ctx context.Context, filter model.AdminCarFilter, pagination model.Pagination) ([]model.Car, error)
	GetCarByID(ctx context.Context, id int64) (model.Car, error)
	CreateCar(ctx context.Context, car model.Car) (int64, error)
	UpdateCar(ctx context.Context, car model.Car) error
	UpdateCarAvailability(ctx context.Context, id int64, isAvailable bool) error
	ArchiveCar(ctx context.Context, id int64) error
	UnarchiveCar(ctx context.Context, id int64) error
	CarSlugExists(ctx context.Context, slug string, excludeID int64) (bool, error)
}

func NewCarService(repo carRepository) *CarService {
	return &CarService{repo: repo}
}

func (s *CarService) ListAvailableCars(ctx context.Context) ([]model.Car, error) {
	cars, err := s.ListCars(ctx, model.CarFilter{Sort: model.SortNewest})
	if err != nil {
		return nil, fmt.Errorf("list available cars: %w", err)
	}

	return cars, nil
}

func (s *CarService) ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {
	filter.Sort = model.NormalizeCarSort(filter.Sort)

	cars, err := s.repo.ListCars(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("list cars: %w", err)
	}

	return cars, nil
}

func (s *CarService) ListCarsPage(ctx context.Context, filter model.CarFilter) ([]model.Car, model.Pagination, error) {
	filter.Sort = model.NormalizeCarSort(filter.Sort)

	totalCount, err := s.repo.CountCars(ctx, filter)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("count cars: %w", err)
	}

	pagination := model.NewPagination(filter.Page, filter.PerPage, totalCount)
	cars, err := s.repo.ListCarsPage(ctx, filter, pagination)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("list cars page: %w", err)
	}

	return cars, pagination, nil
}

func (s *CarService) GetCatalogImageURLs(ctx context.Context, cars []model.Car) (map[int64]string, error) {
	carIDs := make([]int64, 0, len(cars))
	for _, car := range cars {
		if car.ID > 0 {
			carIDs = append(carIDs, car.ID)
		}
	}
	if len(carIDs) == 0 {
		return map[int64]string{}, nil
	}

	imageURLs, err := s.repo.GetCatalogImageURLsByCarIDs(ctx, carIDs)
	if err != nil {
		return nil, fmt.Errorf("get catalog image urls: %w", err)
	}

	return imageURLs, nil
}

func (s *CarService) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	car, err := s.repo.GetCarBySlug(ctx, slug)
	if err != nil {
		return model.Car{}, fmt.Errorf("get car by slug: %w", err)
	}

	return car, nil
}

func (s *CarService) GetCarImages(ctx context.Context, carID int64) ([]model.CarImage, error) {
	images, err := s.repo.GetCarImagesByCarID(ctx, carID)
	if err != nil {
		return nil, fmt.Errorf("get car images: %w", err)
	}

	return images, nil
}

func (s *CarService) AddCarImage(ctx context.Context, image model.CarImage) (int64, error) {
	image.ImageURL = strings.TrimSpace(image.ImageURL)
	image.AltText = strings.TrimSpace(image.AltText)
	if image.CarID < 1 || image.ImageURL == "" {
		return 0, ErrInvalidCarImage
	}

	existingImages, err := s.repo.GetCarImagesByCarID(ctx, image.CarID)
	if err != nil {
		return 0, fmt.Errorf("count car images before add: %w", err)
	}
	image.IsPrimary = len(existingImages) == 0

	id, err := s.repo.CreateCarImage(ctx, image)
	if err != nil {
		return 0, fmt.Errorf("add car image: %w", err)
	}

	return id, nil
}

func (s *CarService) DeleteCarImage(ctx context.Context, carID, imageID int64) error {
	if carID < 1 || imageID < 1 {
		return ErrInvalidCarImage
	}

	image, err := s.repo.GetCarImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("get car image: %w", err)
	}
	if image.CarID != carID {
		return fmt.Errorf("delete car image: %w", ErrCarImageNotFound)
	}

	if err := s.repo.DeleteCarImage(ctx, imageID); err != nil {
		return fmt.Errorf("delete car image: %w", err)
	}

	return nil
}

func (s *CarService) SetPrimaryCarImage(ctx context.Context, carID, imageID int64) error {
	if carID < 1 || imageID < 1 {
		return ErrInvalidCarImage
	}

	if err := s.repo.SetPrimaryCarImage(ctx, carID, imageID); err != nil {
		return fmt.Errorf("set primary car image: %w", err)
	}

	return nil
}

func (s *CarService) ListCarsForAdmin(ctx context.Context) ([]model.Car, error) {
	cars, err := s.repo.ListCarsForAdmin(ctx)
	if err != nil {
		return nil, fmt.Errorf("list cars for admin: %w", err)
	}

	return cars, nil
}

func (s *CarService) ListCarsForAdminPage(ctx context.Context, filter model.AdminCarFilter) ([]model.Car, model.Pagination, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Availability = model.NormalizeAdminCarAvailability(filter.Availability)

	totalCount, err := s.repo.CountCarsForAdmin(ctx, filter)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("count cars for admin: %w", err)
	}

	pagination := model.NewPagination(filter.Page, filter.PerPage, totalCount)
	cars, err := s.repo.ListCarsForAdminPage(ctx, filter, pagination)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("list cars for admin page: %w", err)
	}

	return cars, pagination, nil
}

func (s *CarService) GetCarByID(ctx context.Context, id int64) (model.Car, error) {
	car, err := s.repo.GetCarByID(ctx, id)
	if err != nil {
		return model.Car{}, fmt.Errorf("get car by id: %w", err)
	}

	return car, nil
}

func (s *CarService) CreateCar(ctx context.Context, form model.CarForm) (int64, model.CarForm, error) {
	form = normalizeCarForm(form)
	car := validateCarForm(&form)
	if form.HasErrors() {
		return 0, form, nil
	}

	exists, err := s.repo.CarSlugExists(ctx, form.Slug, 0)
	if err != nil {
		return 0, form, fmt.Errorf("check car slug: %w", err)
	}
	if exists {
		form.Errors["slug"] = "Slug is already used."
		return 0, form, nil
	}

	id, err := s.repo.CreateCar(ctx, car)
	if err != nil {
		return 0, form, fmt.Errorf("create car: %w", err)
	}

	return id, form, nil
}

func (s *CarService) UpdateCar(ctx context.Context, id int64, form model.CarForm) (model.CarForm, error) {
	form = normalizeCarForm(form)
	car := validateCarForm(&form)
	car.ID = id
	if form.HasErrors() {
		return form, nil
	}

	exists, err := s.repo.CarSlugExists(ctx, form.Slug, id)
	if err != nil {
		return form, fmt.Errorf("check car slug: %w", err)
	}
	if exists {
		form.Errors["slug"] = "Slug is already used."
		return form, nil
	}

	if err := s.repo.UpdateCar(ctx, car); err != nil {
		return form, fmt.Errorf("update car: %w", err)
	}

	return form, nil
}

func (s *CarService) UpdateCarAvailability(ctx context.Context, id int64, isAvailable bool) error {
	if err := s.repo.UpdateCarAvailability(ctx, id, isAvailable); err != nil {
		return fmt.Errorf("update car availability: %w", err)
	}

	return nil
}

func (s *CarService) ArchiveCar(ctx context.Context, id int64) error {
	if err := s.repo.ArchiveCar(ctx, id); err != nil {
		return fmt.Errorf("archive car: %w", err)
	}

	return nil
}

func (s *CarService) UnarchiveCar(ctx context.Context, id int64) error {
	if err := s.repo.UnarchiveCar(ctx, id); err != nil {
		return fmt.Errorf("unarchive car: %w", err)
	}

	return nil
}

func normalizeCarForm(form model.CarForm) model.CarForm {
	if form.Errors == nil {
		form.Errors = make(map[string]string)
	}

	form.CategoryID = strings.TrimSpace(form.CategoryID)
	form.Brand = strings.TrimSpace(form.Brand)
	form.Model = strings.TrimSpace(form.Model)
	form.Slug = strings.TrimSpace(form.Slug)
	form.Year = strings.TrimSpace(form.Year)
	form.PricePerDay = strings.TrimSpace(form.PricePerDay)
	form.Transmission = strings.TrimSpace(form.Transmission)
	form.FuelType = strings.TrimSpace(form.FuelType)
	form.Seats = strings.TrimSpace(form.Seats)

	return form
}

func validateCarForm(form *model.CarForm) model.Car {
	categoryID := parseRequiredInt64(form.CategoryID, "category_id", "Category is required.", form.Errors)
	if categoryID <= 0 && form.Errors["category_id"] == "" {
		form.Errors["category_id"] = "Select a valid category."
	}

	if form.Brand == "" {
		form.Errors["brand"] = "Brand is required."
	}

	if form.Model == "" {
		form.Errors["model"] = "Model is required."
	}

	if form.Slug == "" {
		form.Errors["slug"] = "Slug is required."
	} else if !isValidCarSlug(form.Slug) {
		form.Errors["slug"] = "Slug must contain only lowercase letters, numbers, and hyphens."
	}

	year := parseRequiredInt(form.Year, "year", "Year is required.", form.Errors)
	if year != 0 && year < 1990 {
		form.Errors["year"] = "Year must be 1990 or newer."
	}

	pricePerDay := parseRequiredFloat(form.PricePerDay, "price_per_day", "Price per day is required.", form.Errors)
	if pricePerDay <= 0 && form.Errors["price_per_day"] == "" {
		form.Errors["price_per_day"] = "Price per day must be greater than $0."
	}

	if form.Transmission == "" {
		form.Errors["transmission"] = "Transmission is required."
	}

	if form.FuelType == "" {
		form.Errors["fuel_type"] = "Fuel type is required."
	}

	seats := parseRequiredInt(form.Seats, "seats", "Seats are required.", form.Errors)
	if seats <= 0 && form.Errors["seats"] == "" {
		form.Errors["seats"] = "Seats must be at least 1."
	}

	return model.Car{
		CategoryID:   categoryID,
		Brand:        form.Brand,
		Model:        form.Model,
		Slug:         form.Slug,
		Year:         year,
		PricePerDay:  pricePerDay,
		Transmission: form.Transmission,
		FuelType:     form.FuelType,
		Seats:        seats,
		IsAvailable:  form.IsAvailable,
	}
}

func parseRequiredInt64(value, field, requiredMessage string, errors map[string]string) int64 {
	if value == "" {
		errors[field] = requiredMessage
		return 0
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		errors[field] = "Enter a valid number."
		return 0
	}

	return parsed
}

func parseRequiredInt(value, field, requiredMessage string, errors map[string]string) int {
	if value == "" {
		errors[field] = requiredMessage
		return 0
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		errors[field] = "Enter a valid number."
		return 0
	}

	return parsed
}

func parseRequiredFloat(value, field, requiredMessage string, errors map[string]string) float64 {
	if value == "" {
		errors[field] = requiredMessage
		return 0
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		errors[field] = "Enter a valid number."
		return 0
	}

	return parsed
}

func isValidCarSlug(slug string) bool {
	for _, char := range slug {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '-' {
			continue
		}

		return false
	}

	return true
}
