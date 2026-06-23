package handler

import (
	"bytes"
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/go-chi/chi/v5"
)

func TestAdminCleanupPrefillsSetsSuccessFlash(t *testing.T) {
	handler := testFlashHandler()
	handler.bookingPrefillService = &fakeHandlerBookingPrefillService{}

	request := httptest.NewRequest(http.MethodPost, "/admin/cleanup/prefills", nil)
	response := httptest.NewRecorder()

	handler.AdminCleanupPrefills().ServeHTTP(response, request)

	assertRedirect(t, response, "/admin")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Expired prefill tokens cleaned successfully.")
}

func TestLogoutClearsAdminSessionAndSetsSuccessFlash(t *testing.T) {
	handler := testFlashHandler()

	loginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginResponse := httptest.NewRecorder()
	if err := handler.setAdminSession(loginResponse, loginRequest, 42); err != nil {
		t.Fatalf("setAdminSession() error = %v, want nil", err)
	}
	authCookie := latestCookie(t, loginResponse.Result().Cookies(), sessionName)

	authenticatedRequest := httptest.NewRequest(http.MethodGet, "/admin", nil)
	authenticatedRequest.AddCookie(authCookie)
	if !handler.isAdminAuthenticated(authenticatedRequest) {
		t.Fatal("isAdminAuthenticated() = false before logout, want true")
	}

	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(authCookie)
	response := httptest.NewRecorder()

	handler.Logout().ServeHTTP(response, request)

	assertRedirect(t, response, "/login")

	logoutCookie := latestCookie(t, response.Result().Cookies(), sessionName)

	loggedOutRequest := httptest.NewRequest(http.MethodGet, "/admin", nil)
	loggedOutRequest.AddCookie(logoutCookie)
	if handler.isAdminAuthenticated(loggedOutRequest) {
		t.Fatal("isAdminAuthenticated() = true after logout, want false")
	}

	adminResponse := httptest.NewRecorder()
	reachedAdmin := false
	handler.RequireAdminAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reachedAdmin = true
	})).ServeHTTP(adminResponse, loggedOutRequest)
	if reachedAdmin {
		t.Fatal("admin handler was reached after logout")
	}
	assertRedirect(t, adminResponse, "/login")

	assertResponseFlash(t, handler, response, model.FlashSuccess, "You have been logged out.")
}

func TestAdminBookingStatusUpdateSetsSuccessFlash(t *testing.T) {
	bookingRepo := &fakeHandlerBookingRepository{
		getBooking: model.BookingAdminView{
			ID:             42,
			CarID:          7,
			CarBrand:       "Toyota",
			CarModel:       "Corolla",
			CarSlug:        "toyota-corolla",
			CarYear:        2024,
			CustomerName:   "Jane Customer",
			CustomerEmail:  "jane@example.test",
			CustomerPhone:  "555-0100",
			PickupAt:       time.Now().Add(24 * time.Hour),
			ReturnAt:       time.Now().Add(48 * time.Hour),
			BillingDays:    1,
			EstimatedTotal: 90,
			Status:         model.BookingStatusConfirmed,
		},
	}
	handler := testFlashHandler()
	handler.bookingService = service.NewBookingService(bookingRepo, &fakeHandlerBookingCarRepository{}, &fakeHandlerBookingNotifier{})

	form := url.Values{"status": {model.BookingStatusConfirmed}}
	request := formPostRequestWithParam("/admin/bookings/42/status", form, "id", "42")
	response := httptest.NewRecorder()

	handler.AdminBookingStatusUpdate().ServeHTTP(response, request)

	assertRedirect(t, response, "/admin/bookings/42")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Booking status updated to confirmed.")
}

func TestAdminCarArchiveSetsSuccessFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := requestWithParam(http.MethodPost, "/admin/cars/42/archive", "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarArchive().ServeHTTP(response, request)

	if carRepo.archivedID != 42 {
		t.Fatalf("archivedID = %d, want 42", carRepo.archivedID)
	}
	assertRedirect(t, response, "/admin/cars/42")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Car archived successfully.")
}

func TestAdminCarUnarchiveSetsSuccessFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := requestWithParam(http.MethodPost, "/admin/cars/42/unarchive", "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarUnarchive().ServeHTTP(response, request)

	if carRepo.unarchivedID != 42 {
		t.Fatalf("unarchivedID = %d, want 42", carRepo.unarchivedID)
	}
	assertRedirect(t, response, "/admin/cars/42")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Car restored successfully.")
}

func TestAdminCarAvailabilityUpdateSetsEnabledFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	form := url.Values{"is_available": {"on"}}
	request := formPostRequestWithParam("/admin/cars/42/availability", form, "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarAvailabilityUpdate().ServeHTTP(response, request)

	if !carRepo.availabilityValue {
		t.Fatal("availabilityValue = false, want true")
	}
	assertRedirect(t, response, "/admin/cars/42")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Car is now available.")
}

func TestAdminCarAvailabilityUpdateSetsDisabledFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := formPostRequestWithParam("/admin/cars/42/availability", url.Values{}, "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarAvailabilityUpdate().ServeHTTP(response, request)

	if carRepo.availabilityValue {
		t.Fatal("availabilityValue = true, want false")
	}
	assertRedirect(t, response, "/admin/cars/42")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Car is now unavailable.")
}

func TestAdminCarGalleryCreateSetsSuccessFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	form := url.Values{
		"gallery_image_url": {"/static/uploads/cars/gallery.webp"},
		"gallery_alt_text":  {"Front exterior"},
	}
	request := formPostRequestWithParam("/admin/cars/42/gallery", form, "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarGalleryCreate().ServeHTTP(response, request)

	if carRepo.createdCarImage.CarID != 42 {
		t.Fatalf("created car id = %d, want 42", carRepo.createdCarImage.CarID)
	}
	if carRepo.createdCarImage.ImageURL != "/static/uploads/cars/gallery.webp" {
		t.Fatalf("created image URL = %q, want gallery URL", carRepo.createdCarImage.ImageURL)
	}
	if !carRepo.createdCarImage.IsPrimary {
		t.Fatal("created image primary = false, want true for first gallery image")
	}
	assertRedirect(t, response, "/admin/cars/42/edit")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Gallery image added.")
}

func TestAdminCarGalleryCreateRejectsMissingImage(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := formPostRequestWithParam("/admin/cars/42/gallery", url.Values{}, "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarGalleryCreate().ServeHTTP(response, request)

	if carRepo.createdCarImage.ImageURL != "" {
		t.Fatalf("created image URL = %q, want empty", carRepo.createdCarImage.ImageURL)
	}
	assertRedirect(t, response, "/admin/cars/42/edit")
	assertResponseFlash(t, handler, response, model.FlashError, "Image URL or image file is required.")
}

func TestAdminCarGalleryCreateUsesUploadedFile(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{
		getByIDCar: model.Car{ID: 42, Slug: "toyota-corolla"},
		carImages: []model.CarImage{
			{ID: 1, CarID: 42, ImageURL: "/static/uploads/cars/existing.webp", IsPrimary: true},
		},
	}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := multipartGalleryUploadRequestWithParam(t, "/admin/cars/42/gallery", "id", "42")
	response := httptest.NewRecorder()

	handler.AdminCarGalleryCreate().ServeHTTP(response, request)
	t.Cleanup(func() {
		if carRepo.createdCarImage.ImageURL != "" {
			_ = os.Remove("web/static" + strings.TrimPrefix(carRepo.createdCarImage.ImageURL, "/static"))
		}
	})

	if !strings.HasPrefix(carRepo.createdCarImage.ImageURL, "/static/uploads/cars/toyota-corolla-") {
		t.Fatalf("created image URL = %q, want uploaded car image URL", carRepo.createdCarImage.ImageURL)
	}
	if carRepo.createdCarImage.IsPrimary {
		t.Fatal("created image primary = true, want false when gallery already has images")
	}
	assertRedirect(t, response, "/admin/cars/42/edit")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Gallery image added.")
}

func TestAdminCarGallerySetPrimarySetsSuccessFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := requestWithParams(http.MethodPost, "/admin/cars/42/gallery/9/primary", map[string]string{
		"id":      "42",
		"imageID": "9",
	})
	response := httptest.NewRecorder()

	handler.AdminCarGallerySetPrimary().ServeHTTP(response, request)

	if carRepo.primaryCarID != 42 {
		t.Fatalf("primary car id = %d, want 42", carRepo.primaryCarID)
	}
	if carRepo.primaryImageID != 9 {
		t.Fatalf("primary image id = %d, want 9", carRepo.primaryImageID)
	}
	assertRedirect(t, response, "/admin/cars/42/edit")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Primary image updated.")
}

func TestAdminCarGalleryDeleteSetsSuccessFlash(t *testing.T) {
	carRepo := &fakeHandlerCarRepository{}
	handler := testFlashHandler()
	handler.carService = service.NewCarService(carRepo)

	request := requestWithParams(http.MethodPost, "/admin/cars/42/gallery/9/delete", map[string]string{
		"id":      "42",
		"imageID": "9",
	})
	response := httptest.NewRecorder()

	handler.AdminCarGalleryDelete().ServeHTTP(response, request)

	if carRepo.deletedCarImageID != 9 {
		t.Fatalf("deleted image id = %d, want 9", carRepo.deletedCarImageID)
	}
	assertRedirect(t, response, "/admin/cars/42/edit")
	assertResponseFlash(t, handler, response, model.FlashSuccess, "Gallery image deleted.")
}

func assertRedirect(t *testing.T, response *httptest.ResponseRecorder, location string) {
	t.Helper()

	if response.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusSeeOther)
	}

	if got := response.Header().Get("Location"); got != location {
		t.Fatalf("Location = %q, want %q", got, location)
	}
}

func assertResponseFlash(t *testing.T, handler *Handler, response *httptest.ResponseRecorder, flashType model.FlashType, message string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(latestCookie(t, response.Result().Cookies(), sessionName))
	popResponse := httptest.NewRecorder()

	flash, err := handler.popFlash(popResponse, request)
	if err != nil {
		t.Fatalf("popFlash() error = %v, want nil", err)
	}
	if flash == nil {
		t.Fatal("popFlash() flash = nil, want flash")
	}
	if flash.Type != flashType {
		t.Fatalf("flash.Type = %q, want %q", flash.Type, flashType)
	}
	if flash.Message != message {
		t.Fatalf("flash.Message = %q, want %q", flash.Message, message)
	}
}

func requestWithParam(method, target, key, value string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	return addRouteParam(request, key, value)
}

func requestWithParams(method, target string, params map[string]string) *http.Request {
	request := httptest.NewRequest(method, target, nil)
	return addRouteParams(request, params)
}

func formPostRequestWithParam(target string, form url.Values, key, value string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return addRouteParam(request, key, value)
}

func multipartGalleryUploadRequestWithParam(t *testing.T, target, key, value string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("gallery_alt_text", "Uploaded gallery image"); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}
	part, err := writer.CreateFormFile("gallery_image_file", "gallery.jpg")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(jpegBytes()); err != nil {
		t.Fatalf("part.Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, target, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return addRouteParam(request, key, value)
}

func addRouteParam(request *http.Request, key, value string) *http.Request {
	return addRouteParams(request, map[string]string{key: value})
}

func addRouteParams(request *http.Request, params map[string]string) *http.Request {
	routeContext := chi.NewRouteContext()
	for key, value := range params {
		routeContext.URLParams.Add(key, value)
	}
	ctx := context.WithValue(request.Context(), chi.RouteCtxKey, routeContext)
	return request.WithContext(ctx)
}

type fakeHandlerBookingPrefillService struct {
	err error
}

func (s *fakeHandlerBookingPrefillService) CreateFromBookingForm(ctx context.Context, form model.BookingForm) (string, error) {
	return "", errors.New("not implemented")
}

func (s *fakeHandlerBookingPrefillService) GetFormByToken(ctx context.Context, token string) (model.BookingForm, error) {
	return model.NewBookingForm(), errors.New("not implemented")
}

func (s *fakeHandlerBookingPrefillService) CleanupExpiredPrefills(ctx context.Context) error {
	return s.err
}

type fakeHandlerBookingRepository struct {
	getBooking   model.BookingAdminView
	getErr       error
	updateErr    error
	exportRows   []model.BookingExportRow
	exportErr    error
	exportFilter model.AdminBookingFilter
}

func (r *fakeHandlerBookingRepository) CreateBooking(ctx context.Context, booking model.Booking) (int64, error) {
	return 0, nil
}

func (r *fakeHandlerBookingRepository) HasBookingConflict(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (bool, error) {
	return false, nil
}

func (r *fakeHandlerBookingRepository) FindNextAvailablePickupAt(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (r *fakeHandlerBookingRepository) ListBlockingBookingsForCar(ctx context.Context, carID int64, from time.Time, to time.Time) ([]model.Booking, error) {
	return nil, nil
}

func (r *fakeHandlerBookingRepository) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeHandlerBookingRepository) CountBookings(ctx context.Context, filter model.AdminBookingFilter) (int, error) {
	return 0, nil
}

func (r *fakeHandlerBookingRepository) ListBookingsPage(ctx context.Context, filter model.AdminBookingFilter, pagination model.Pagination) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeHandlerBookingRepository) ListBookingsForExport(ctx context.Context, filter model.AdminBookingFilter) ([]model.BookingExportRow, error) {
	r.exportFilter = filter
	if r.exportErr != nil {
		return nil, r.exportErr
	}

	return r.exportRows, nil
}

func (r *fakeHandlerBookingRepository) GetBookingStats(ctx context.Context, dashboardRange model.DashboardRange) (model.BookingStats, error) {
	return model.BookingStats{}, nil
}

func (r *fakeHandlerBookingRepository) GetRevenueStats(ctx context.Context, dashboardRange model.DashboardRange) (model.RevenueStats, error) {
	return model.RevenueStats{}, nil
}

func (r *fakeHandlerBookingRepository) GetRecentBookings(ctx context.Context, limit int, dashboardRange model.DashboardRange) ([]model.RecentBookingActivity, error) {
	return nil, nil
}

func (r *fakeHandlerBookingRepository) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	if r.getErr != nil {
		return model.BookingAdminView{}, r.getErr
	}

	return r.getBooking, nil
}

func (r *fakeHandlerBookingRepository) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	if r.updateErr != nil {
		return r.updateErr
	}

	r.getBooking.ID = id
	r.getBooking.Status = status
	return nil
}

type fakeHandlerBookingCarRepository struct{}

func (r *fakeHandlerBookingCarRepository) ListAvailableAlternativeCars(
	ctx context.Context,
	currentCarID int64,
	categoryID int64,
	minPrice float64,
	maxPrice float64,
	pickupAt time.Time,
	returnAt time.Time,
	bufferHours int,
	limit int,
) ([]model.Car, error) {
	return nil, nil
}

type fakeHandlerBookingNotifier struct{}

func (n *fakeHandlerBookingNotifier) NotifyAdminBookingCreated(ctx context.Context, booking model.Booking, car model.Car) error {
	return nil
}

func (n *fakeHandlerBookingNotifier) NotifyCustomerBookingStatusChanged(ctx context.Context, booking model.Booking, car model.Car) error {
	return nil
}

type fakeHandlerCarRepository struct {
	archivedID        int64
	unarchivedID      int64
	availabilityID    int64
	availabilityValue bool
	getByIDCar        model.Car
	getBySlugCar      model.Car
	carImages         []model.CarImage
	getCarImagesID    int64
	createdCarImage   model.CarImage
	deletedCarImageID int64
	primaryCarID      int64
	primaryImageID    int64
	countErr          error
	getBySlugErr      error
	getCarImagesErr   error
	getByIDErr        error
	updateErr         error
	availabilityErr   error
	archiveErr        error
	unarchiveErr      error
}

func (r *fakeHandlerCarRepository) ListCars(ctx context.Context, filter model.CarFilter) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeHandlerCarRepository) CountCars(ctx context.Context, filter model.CarFilter) (int, error) {
	if r.countErr != nil {
		return 0, r.countErr
	}

	return 0, nil
}

func (r *fakeHandlerCarRepository) ListCarsPage(ctx context.Context, filter model.CarFilter, pagination model.Pagination) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeHandlerCarRepository) GetCarBySlug(ctx context.Context, slug string) (model.Car, error) {
	if r.getBySlugErr != nil {
		return model.Car{}, r.getBySlugErr
	}

	return r.getBySlugCar, nil
}

func (r *fakeHandlerCarRepository) GetCarImagesByCarID(ctx context.Context, carID int64) ([]model.CarImage, error) {
	if r.getCarImagesErr != nil {
		return nil, r.getCarImagesErr
	}

	r.getCarImagesID = carID
	return r.carImages, nil
}

func (r *fakeHandlerCarRepository) GetCatalogImageURLsByCarIDs(ctx context.Context, carIDs []int64) (map[int64]string, error) {
	return map[int64]string{}, nil
}

func (r *fakeHandlerCarRepository) CreateCarImage(ctx context.Context, image model.CarImage) (int64, error) {
	r.createdCarImage = image
	return 1, nil
}

func (r *fakeHandlerCarRepository) GetCarImageByID(ctx context.Context, imageID int64) (model.CarImage, error) {
	return model.CarImage{ID: imageID, CarID: 42, ImageURL: "/static/uploads/cars/gallery.webp"}, nil
}

func (r *fakeHandlerCarRepository) DeleteCarImage(ctx context.Context, imageID int64) error {
	r.deletedCarImageID = imageID
	return nil
}

func (r *fakeHandlerCarRepository) SetPrimaryCarImage(ctx context.Context, carID, imageID int64) error {
	r.primaryCarID = carID
	r.primaryImageID = imageID
	return nil
}

func (r *fakeHandlerCarRepository) ListCarsForAdmin(ctx context.Context) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeHandlerCarRepository) CountCarsForAdmin(ctx context.Context, filter model.AdminCarFilter) (int, error) {
	return 0, nil
}

func (r *fakeHandlerCarRepository) ListCarsForAdminPage(ctx context.Context, filter model.AdminCarFilter, pagination model.Pagination) ([]model.Car, error) {
	return nil, nil
}

func (r *fakeHandlerCarRepository) GetCarByID(ctx context.Context, id int64) (model.Car, error) {
	if r.getByIDErr != nil {
		return model.Car{}, r.getByIDErr
	}

	if r.getByIDCar.ID != 0 {
		return r.getByIDCar, nil
	}

	return model.Car{ID: id}, nil
}

func (r *fakeHandlerCarRepository) CreateCar(ctx context.Context, car model.Car) (int64, error) {
	return 0, nil
}

func (r *fakeHandlerCarRepository) UpdateCar(ctx context.Context, car model.Car) error {
	if r.updateErr != nil {
		return r.updateErr
	}

	return nil
}

func (r *fakeHandlerCarRepository) UpdateCarAvailability(ctx context.Context, id int64, isAvailable bool) error {
	if r.availabilityErr != nil {
		return r.availabilityErr
	}

	r.availabilityID = id
	r.availabilityValue = isAvailable
	return nil
}

func (r *fakeHandlerCarRepository) ArchiveCar(ctx context.Context, id int64) error {
	if r.archiveErr != nil {
		return r.archiveErr
	}

	r.archivedID = id
	return nil
}

func (r *fakeHandlerCarRepository) UnarchiveCar(ctx context.Context, id int64) error {
	if r.unarchiveErr != nil {
		return r.unarchiveErr
	}

	r.unarchivedID = id
	return nil
}

func (r *fakeHandlerCarRepository) CarSlugExists(ctx context.Context, slug string, excludeID int64) (bool, error) {
	return false, nil
}
