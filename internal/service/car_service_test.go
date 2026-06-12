package service

import (
	"context"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestValidateCarFormRejectsInvalidSlug(t *testing.T) {
	form := normalizeCarForm(validCarForm())
	form.Slug = "Toyota Corolla"

	validateCarForm(&form)

	if form.Errors["slug"] == "" {
		t.Fatal("validateCarForm() did not reject invalid slug")
	}
}

func TestValidateCarFormRejectsInvalidNumericFields(t *testing.T) {
	form := normalizeCarForm(validCarForm())
	form.CategoryID = "invalid"
	form.Year = "1989"
	form.PricePerDay = "0"
	form.Seats = "-1"

	validateCarForm(&form)

	if form.Errors["category_id"] == "" {
		t.Fatal("validateCarForm() did not reject invalid category id")
	}

	if form.Errors["year"] == "" {
		t.Fatal("validateCarForm() did not reject year before 1990")
	}

	if form.Errors["price_per_day"] == "" {
		t.Fatal("validateCarForm() did not reject non-positive price")
	}

	if form.Errors["seats"] == "" {
		t.Fatal("validateCarForm() did not reject non-positive seats")
	}
}

func TestValidateCarFormAcceptsValidInput(t *testing.T) {
	form := normalizeCarForm(validCarForm())

	car := validateCarForm(&form)

	if form.HasErrors() {
		t.Fatalf("validateCarForm() errors = %v, want none", form.Errors)
	}

	if car.CategoryID != 1 {
		t.Fatalf("validateCarForm() CategoryID = %d, want 1", car.CategoryID)
	}

	if car.PricePerDay != 75.50 {
		t.Fatalf("validateCarForm() PricePerDay = %f, want 75.50", car.PricePerDay)
	}
}

func TestValidateCarFormImageURL(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
		wantErr  bool
	}{
		{
			name:     "empty image url",
			imageURL: "",
			wantErr:  false,
		},
		{
			name:     "https image url",
			imageURL: "https://example.com/car.jpg",
			wantErr:  false,
		},
		{
			name:     "http image url",
			imageURL: "http://example.com/car.jpg",
			wantErr:  false,
		},
		{
			name:     "static image url",
			imageURL: "/static/uploads/cars/car.jpg",
			wantErr:  false,
		},
		{
			name:     "invalid image url",
			imageURL: "example.com/car.jpg",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := normalizeCarForm(validCarForm())
			form.ImageURL = tt.imageURL

			validateCarForm(&form)

			gotErr := form.Errors["image_url"] != ""
			if gotErr != tt.wantErr {
				t.Fatalf("validateCarForm() image_url error = %v, want %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestCarServiceArchiveCarDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{}
	service := NewCarService(repo)

	if err := service.ArchiveCar(context.Background(), 42); err != nil {
		t.Fatalf("ArchiveCar() error = %v, want nil", err)
	}

	if !repo.archiveCalled {
		t.Fatal("ArchiveCar() did not call repository")
	}

	if repo.archiveID != 42 {
		t.Fatalf("ArchiveCar() id = %d, want 42", repo.archiveID)
	}
}

func TestCarServiceUnarchiveCarDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{}
	service := NewCarService(repo)

	if err := service.UnarchiveCar(context.Background(), 42); err != nil {
		t.Fatalf("UnarchiveCar() error = %v, want nil", err)
	}

	if !repo.unarchiveCalled {
		t.Fatal("UnarchiveCar() did not call repository")
	}

	if repo.unarchiveID != 42 {
		t.Fatalf("UnarchiveCar() id = %d, want 42", repo.unarchiveID)
	}
}

func validCarForm() model.CarForm {
	return model.CarForm{
		CategoryID:   "1",
		Brand:        "Toyota",
		Model:        "Corolla",
		Slug:         "toyota-corolla",
		Year:         "2024",
		PricePerDay:  "75.50",
		Transmission: "Automatic",
		FuelType:     "Gasoline",
		Seats:        "5",
		IsAvailable:  true,
		Errors:       make(map[string]string),
	}
}

type fakeCarRepository struct {
	archiveCalled   bool
	archiveID       int64
	unarchiveCalled bool
	unarchiveID     int64
}

func (r *fakeCarRepository) ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeCarRepository) CountCars(ctx context.Context, filter model.CarFilter) (int, error) {
	return 0, nil
}

func (r *fakeCarRepository) ListCarsPage(ctx context.Context, filter model.CarFilter, pagination model.Pagination) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeCarRepository) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	return model.Car{}, nil
}

func (r *fakeCarRepository) ListCarsForAdmin(ctx context.Context) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeCarRepository) CountCarsForAdmin(ctx context.Context, filter model.AdminCarFilter) (int, error) {
	return 0, nil
}

func (r *fakeCarRepository) ListCarsForAdminPage(ctx context.Context, filter model.AdminCarFilter, pagination model.Pagination) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeCarRepository) GetCarByID(ctx context.Context, id int64) (model.Car, error) {
	return model.Car{}, nil
}

func (r *fakeCarRepository) CreateCar(ctx context.Context, car model.Car) (int64, error) {
	return 0, nil
}

func (r *fakeCarRepository) UpdateCar(ctx context.Context, car model.Car) error {
	return nil
}

func (r *fakeCarRepository) UpdateCarAvailability(ctx context.Context, id int64, isAvailable bool) error {
	return nil
}

func (r *fakeCarRepository) ArchiveCar(ctx context.Context, id int64) error {
	r.archiveCalled = true
	r.archiveID = id
	return nil
}

func (r *fakeCarRepository) UnarchiveCar(ctx context.Context, id int64) error {
	r.unarchiveCalled = true
	r.unarchiveID = id
	return nil
}

func (r *fakeCarRepository) CarSlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	return false, nil
}
