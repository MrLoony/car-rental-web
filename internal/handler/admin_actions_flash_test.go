package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func formPostRequestWithParam(target string, form url.Values, key, value string) *http.Request {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return addRouteParam(request, key, value)
}

func addRouteParam(request *http.Request, key, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
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

func (r *fakeHandlerBookingRepository) GetBookingStats(ctx context.Context) (model.BookingStats, error) {
	return model.BookingStats{}, nil
}

func (r *fakeHandlerBookingRepository) GetRevenueStats(ctx context.Context) (model.RevenueStats, error) {
	return model.RevenueStats{}, nil
}

func (r *fakeHandlerBookingRepository) GetRecentBookings(ctx context.Context, limit int) ([]model.RecentBookingActivity, error) {
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
	countErr          error
	getBySlugErr      error
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

	return model.Car{}, nil
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
