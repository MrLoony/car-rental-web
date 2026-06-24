package service

import (
	"context"
	"errors"
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

func TestValidateCarFormUsesFriendlyNumericMessages(t *testing.T) {
	form := normalizeCarForm(validCarForm())
	form.PricePerDay = "0"
	form.Seats = "0"

	validateCarForm(&form)

	if form.Errors["price_per_day"] != "Price per day must be greater than $0." {
		t.Fatalf("price_per_day error = %q", form.Errors["price_per_day"])
	}
	if form.Errors["seats"] != "Seats must be at least 1." {
		t.Fatalf("seats error = %q", form.Errors["seats"])
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

func TestCarServiceGetCarImagesDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{
		carImages: []model.CarImage{
			{ID: 1, CarID: 42, ImageURL: "/static/uploads/cars/front.webp", IsPrimary: true},
			{ID: 2, CarID: 42, ImageURL: "/static/uploads/cars/interior.webp", SortOrder: 1},
		},
	}
	service := NewCarService(repo)

	images, err := service.GetCarImages(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetCarImages() error = %v, want nil", err)
	}

	if !repo.getCarImagesCalled {
		t.Fatal("GetCarImages() did not call repository")
	}
	if repo.getCarImagesCarID != 42 {
		t.Fatalf("GetCarImages() carID = %d, want 42", repo.getCarImagesCarID)
	}
	if len(images) != 2 {
		t.Fatalf("GetCarImages() returned %d images, want 2", len(images))
	}
	if images[0].ImageURL != "/static/uploads/cars/front.webp" {
		t.Fatalf("GetCarImages()[0].ImageURL = %q, want front image", images[0].ImageURL)
	}
}

func TestCarServiceGetCarImagesWrapsRepositoryError(t *testing.T) {
	repoErr := errors.New("database unavailable")
	repo := &fakeCarRepository{getCarImagesErr: repoErr}
	service := NewCarService(repo)

	_, err := service.GetCarImages(context.Background(), 42)
	if err == nil {
		t.Fatal("GetCarImages() error = nil, want error")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("GetCarImages() error = %v, want to wrap %v", err, repoErr)
	}
}

func TestCarServiceListCarsPageUsesPrimaryGalleryImageForCatalog(t *testing.T) {
	repo := &fakeCarRepository{
		catalogImageURLs: map[int64]string{
			42: "/static/uploads/cars/primary-gallery.webp",
		},
	}
	service := NewCarService(repo)

	imageURLs, err := service.GetCatalogImageURLs(context.Background(), []model.Car{{ID: 42}})
	if err != nil {
		t.Fatalf("GetCatalogImageURLs() error = %v, want nil", err)
	}

	if got := imageURLs[42]; got != "/static/uploads/cars/primary-gallery.webp" {
		t.Fatalf("catalog image URL = %q, want primary gallery image", got)
	}
}

func TestCarServiceListCarsPageUsesFirstGalleryImageFallback(t *testing.T) {
	repo := &fakeCarRepository{
		catalogImageURLs: map[int64]string{
			42: "/static/uploads/cars/first-gallery.webp",
		},
	}
	service := NewCarService(repo)

	imageURLs, err := service.GetCatalogImageURLs(context.Background(), []model.Car{{ID: 42}})
	if err != nil {
		t.Fatalf("GetCatalogImageURLs() error = %v, want nil", err)
	}

	if got := imageURLs[42]; got != "/static/uploads/cars/first-gallery.webp" {
		t.Fatalf("catalog image URL = %q, want first gallery image", got)
	}
}

func TestCarServiceGetCatalogImageURLsReturnsEmptyForPlaceholderFallback(t *testing.T) {
	repo := &fakeCarRepository{
		catalogImageURLs: map[int64]string{},
	}
	service := NewCarService(repo)

	imageURLs, err := service.GetCatalogImageURLs(context.Background(), []model.Car{{ID: 42}})
	if err != nil {
		t.Fatalf("GetCatalogImageURLs() error = %v, want nil", err)
	}

	if got := imageURLs[42]; got != "" {
		t.Fatalf("catalog image URL = %q, want empty", got)
	}
}

func TestCarServiceGetCatalogImageURLsHandlesNoCars(t *testing.T) {
	repo := &fakeCarRepository{}
	service := NewCarService(repo)

	imageURLs, err := service.GetCatalogImageURLs(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetCatalogImageURLs() error = %v, want nil", err)
	}

	if len(imageURLs) != 0 {
		t.Fatalf("catalog image URLs length = %d, want 0", len(imageURLs))
	}
}

func TestCarServiceAddCarImageTrimsAndDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{createCarImageID: 9}
	service := NewCarService(repo)

	id, err := service.AddCarImage(context.Background(), model.CarImage{
		CarID:    42,
		ImageURL: "  /static/uploads/cars/front.webp  ",
		AltText:  "  Front exterior  ",
	})
	if err != nil {
		t.Fatalf("AddCarImage() error = %v, want nil", err)
	}

	if id != 9 {
		t.Fatalf("AddCarImage() id = %d, want 9", id)
	}
	if !repo.createCarImageCalled {
		t.Fatal("AddCarImage() did not call repository")
	}
	if repo.createdCarImage.ImageURL != "/static/uploads/cars/front.webp" {
		t.Fatalf("created image URL = %q, want trimmed URL", repo.createdCarImage.ImageURL)
	}
	if repo.createdCarImage.AltText != "Front exterior" {
		t.Fatalf("created alt text = %q, want trimmed alt text", repo.createdCarImage.AltText)
	}
	if !repo.createdCarImage.IsPrimary {
		t.Fatal("created image primary = false, want true for first gallery image")
	}
}

func TestCarServiceAddCarImageRejectsMissingURL(t *testing.T) {
	service := NewCarService(&fakeCarRepository{})

	_, err := service.AddCarImage(context.Background(), model.CarImage{CarID: 42})
	if !errors.Is(err, ErrInvalidCarImage) {
		t.Fatalf("AddCarImage() error = %v, want ErrInvalidCarImage", err)
	}
}

func TestCarServiceAddCarImageDoesNotReplaceExistingPrimary(t *testing.T) {
	repo := &fakeCarRepository{
		createCarImageID: 10,
		carImages: []model.CarImage{
			{ID: 9, CarID: 42, ImageURL: "/static/uploads/cars/front.webp", IsPrimary: true},
		},
	}
	service := NewCarService(repo)

	_, err := service.AddCarImage(context.Background(), model.CarImage{
		CarID:    42,
		ImageURL: "/static/uploads/cars/rear.webp",
		AltText:  "Rear exterior",
	})
	if err != nil {
		t.Fatalf("AddCarImage() error = %v, want nil", err)
	}

	if repo.createdCarImage.IsPrimary {
		t.Fatal("created image primary = true, want false when gallery already has images")
	}
}

func TestCarServiceAddCarImagesMarksOnlyFirstPrimaryForEmptyGallery(t *testing.T) {
	repo := &fakeCarRepository{createCarImageID: 10}
	service := NewCarService(repo)

	_, err := service.AddCarImages(context.Background(), []model.CarImage{
		{CarID: 42, ImageURL: "/static/uploads/cars/front.webp"},
		{CarID: 42, ImageURL: "/static/uploads/cars/rear.webp"},
		{CarID: 42, ImageURL: "/static/uploads/cars/interior.webp"},
	})
	if err != nil {
		t.Fatalf("AddCarImages() error = %v, want nil", err)
	}

	if len(repo.createdCarImages) != 3 {
		t.Fatalf("created images = %d, want 3", len(repo.createdCarImages))
	}
	if !repo.createdCarImages[0].IsPrimary {
		t.Fatal("first created image primary = false, want true")
	}
	for i, image := range repo.createdCarImages[1:] {
		if image.IsPrimary {
			t.Fatalf("created image %d primary = true, want false", i+1)
		}
	}
}

func TestCarServiceAddCarImagesKeepsExistingPrimary(t *testing.T) {
	repo := &fakeCarRepository{
		createCarImageID: 10,
		carImages: []model.CarImage{
			{ID: 9, CarID: 42, ImageURL: "/static/uploads/cars/front.webp", IsPrimary: true},
		},
	}
	service := NewCarService(repo)

	_, err := service.AddCarImages(context.Background(), []model.CarImage{
		{CarID: 42, ImageURL: "/static/uploads/cars/rear.webp"},
		{CarID: 42, ImageURL: "/static/uploads/cars/interior.webp"},
	})
	if err != nil {
		t.Fatalf("AddCarImages() error = %v, want nil", err)
	}

	if len(repo.createdCarImages) != 2 {
		t.Fatalf("created images = %d, want 2", len(repo.createdCarImages))
	}
	for i, image := range repo.createdCarImages {
		if image.IsPrimary {
			t.Fatalf("created image %d primary = true, want false", i)
		}
	}
}

func TestCarServiceDeleteCarImageDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{
		carImageByID: model.CarImage{ID: 9, CarID: 42, ImageURL: "/static/uploads/cars/gallery.webp"},
	}
	service := NewCarService(repo)

	if err := service.DeleteCarImage(context.Background(), 42, 9); err != nil {
		t.Fatalf("DeleteCarImage() error = %v, want nil", err)
	}

	if repo.deleteCarImageID != 9 {
		t.Fatalf("delete image id = %d, want 9", repo.deleteCarImageID)
	}
}

func TestCarServiceDeleteCarImageRejectsMismatchedCar(t *testing.T) {
	repo := &fakeCarRepository{
		carImageByID: model.CarImage{ID: 9, CarID: 100, ImageURL: "/static/uploads/cars/gallery.webp"},
	}
	service := NewCarService(repo)

	err := service.DeleteCarImage(context.Background(), 42, 9)
	if !errors.Is(err, ErrCarImageNotFound) {
		t.Fatalf("DeleteCarImage() error = %v, want ErrCarImageNotFound", err)
	}
	if repo.deleteCarImageID != 0 {
		t.Fatalf("delete image id = %d, want 0", repo.deleteCarImageID)
	}
}

func TestCarServiceSetPrimaryCarImageDelegatesToRepository(t *testing.T) {
	repo := &fakeCarRepository{}
	service := NewCarService(repo)

	if err := service.SetPrimaryCarImage(context.Background(), 42, 9); err != nil {
		t.Fatalf("SetPrimaryCarImage() error = %v, want nil", err)
	}

	if repo.primaryCarID != 42 {
		t.Fatalf("primary car id = %d, want 42", repo.primaryCarID)
	}
	if repo.primaryImageID != 9 {
		t.Fatalf("primary image id = %d, want 9", repo.primaryImageID)
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
	archiveCalled        bool
	archiveID            int64
	unarchiveCalled      bool
	unarchiveID          int64
	countCars            int
	listCarsPageCars     []model.Car
	catalogImageURLs     map[int64]string
	carImages            []model.CarImage
	getCarImagesCalled   bool
	getCarImagesCarID    int64
	getCarImagesErr      error
	createCarImageID     int64
	createCarImageCalled bool
	createdCarImage      model.CarImage
	createdCarImages     []model.CarImage
	carImageByID         model.CarImage
	deleteCarImageID     int64
	primaryCarID         int64
	primaryImageID       int64
	updatedCar           model.Car
}

func (r *fakeCarRepository) ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeCarRepository) CountCars(ctx context.Context, filter model.CarFilter) (int, error) {
	return r.countCars, nil
}

func (r *fakeCarRepository) ListCarsPage(ctx context.Context, filter model.CarFilter, pagination model.Pagination) ([]model.Car, error) {
	return r.listCarsPageCars, nil
}

func (r *fakeCarRepository) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	return model.Car{}, nil
}

func (r *fakeCarRepository) GetCarImagesByCarID(ctx context.Context, carID int64) ([]model.CarImage, error) {
	r.getCarImagesCalled = true
	r.getCarImagesCarID = carID
	if r.getCarImagesErr != nil {
		return nil, r.getCarImagesErr
	}

	return r.carImages, nil
}

func (r *fakeCarRepository) GetCatalogImageURLsByCarIDs(ctx context.Context, carIDs []int64) (map[int64]string, error) {
	if r.catalogImageURLs == nil {
		return map[int64]string{}, nil
	}

	return r.catalogImageURLs, nil
}

func (r *fakeCarRepository) CreateCarImage(ctx context.Context, image model.CarImage) (int64, error) {
	r.createCarImageCalled = true
	r.createdCarImage = image
	r.createdCarImages = append(r.createdCarImages, image)
	return r.createCarImageID, nil
}

func (r *fakeCarRepository) GetCarImageByID(ctx context.Context, imageID int64) (model.CarImage, error) {
	if r.carImageByID.ID == 0 {
		return model.CarImage{ID: imageID}, nil
	}

	return r.carImageByID, nil
}

func (r *fakeCarRepository) DeleteCarImage(ctx context.Context, imageID int64) error {
	r.deleteCarImageID = imageID
	return nil
}

func (r *fakeCarRepository) SetPrimaryCarImage(ctx context.Context, carID, imageID int64) error {
	r.primaryCarID = carID
	r.primaryImageID = imageID
	return nil
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
	r.updatedCar = car
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
