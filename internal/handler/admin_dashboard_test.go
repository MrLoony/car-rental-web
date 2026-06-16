package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/service"
)

func TestAdminIndexRendersBookingStats(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(&fakeAdminDashboardBookingRepository{
		stats: model.BookingStats{
			Total:     10,
			Pending:   2,
			Confirmed: 3,
			Cancelled: 1,
			Completed: 4,
		},
		revenueStats: model.RevenueStats{
			TotalPotential: 1000,
			Pending:        200,
			Confirmed:      300,
			Completed:      400,
			Cancelled:      100,
		},
		recentBookings: []model.RecentBookingActivity{
			{
				ID:           42,
				CustomerName: "Jane Customer",
				CarName:      "Toyota Corolla",
				Status:       model.BookingStatusPending,
				PickupTime:   time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC),
				ReturnTime:   time.Date(2026, time.July, 12, 11, 0, 0, 0, time.UTC),
				CreatedAt:    time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()

	handler.AdminIndex().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	for _, want := range []string{
		"Total bookings",
		">10<",
		"Pending",
		">2<",
		"Confirmed",
		">3<",
		"Cancelled",
		">1<",
		"Completed",
		">4<",
		"Total potential revenue",
		"$1000.00",
		"Pending revenue",
		"$200.00",
		"Confirmed revenue",
		"$300.00",
		"Completed revenue",
		"$400.00",
		"Cancelled value",
		"$100.00",
		"Recent booking activity",
		"Jane Customer",
		"Toyota Corolla",
		"/admin/bookings/42",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body does not contain %q:\n%s", want, body)
		}
	}
}

func TestAdminIndexRendersZeroRevenueStatsAndEmptyRecentBookings(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	bookingRepo := &fakeAdminDashboardBookingRepository{}
	handler.bookingService = service.NewBookingService(bookingRepo, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()

	handler.AdminIndex().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	body := response.Body.String()
	if !strings.Contains(body, "$0.00") {
		t.Fatalf("body does not contain zero revenue value:\n%s", body)
	}
	if !strings.Contains(body, "No recent bookings.") {
		t.Fatalf("body does not contain recent bookings empty state:\n%s", body)
	}
	if bookingRepo.recentBookingsLimit != recentBookingsDashboardLimit {
		t.Fatalf("recentBookingsLimit = %d, want %d", bookingRepo.recentBookingsLimit, recentBookingsDashboardLimit)
	}
}

func TestAdminIndexStatsFailureReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(&fakeAdminDashboardBookingRepository{
		statsErr: errors.New("database password leaked"),
	}, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()

	handler.AdminIndex().ServeHTTP(response, request)

	assertServerErrorPage(t, response)
}

func TestAdminIndexRevenueStatsFailureReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(&fakeAdminDashboardBookingRepository{
		revenueStatsErr: errors.New("database password leaked"),
	}, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()

	handler.AdminIndex().ServeHTTP(response, request)

	assertServerErrorPage(t, response)
}

func TestAdminIndexRecentBookingsFailureReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(&fakeAdminDashboardBookingRepository{
		recentBookingsErr: errors.New("database password leaked"),
	}, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()

	handler.AdminIndex().ServeHTTP(response, request)

	assertServerErrorPage(t, response)
}

type fakeAdminDashboardBookingRepository struct {
	stats               model.BookingStats
	statsErr            error
	revenueStats        model.RevenueStats
	revenueStatsErr     error
	recentBookings      []model.RecentBookingActivity
	recentBookingsLimit int
	recentBookingsErr   error
}

func (r *fakeAdminDashboardBookingRepository) CreateBooking(ctx context.Context, booking model.Booking) (int64, error) {
	return 0, nil
}

func (r *fakeAdminDashboardBookingRepository) HasBookingConflict(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (bool, error) {
	return false, nil
}

func (r *fakeAdminDashboardBookingRepository) FindNextAvailablePickupAt(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (r *fakeAdminDashboardBookingRepository) ListBlockingBookingsForCar(ctx context.Context, carID int64, from time.Time, to time.Time) ([]model.Booking, error) {
	return nil, nil
}

func (r *fakeAdminDashboardBookingRepository) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeAdminDashboardBookingRepository) CountBookings(ctx context.Context, filter model.AdminBookingFilter) (int, error) {
	return 0, nil
}

func (r *fakeAdminDashboardBookingRepository) ListBookingsPage(ctx context.Context, filter model.AdminBookingFilter, pagination model.Pagination) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeAdminDashboardBookingRepository) ListBookingsForExport(ctx context.Context, filter model.AdminBookingFilter) ([]model.BookingExportRow, error) {
	return nil, nil
}

func (r *fakeAdminDashboardBookingRepository) GetBookingStats(ctx context.Context) (model.BookingStats, error) {
	if r.statsErr != nil {
		return model.BookingStats{}, r.statsErr
	}

	return r.stats, nil
}

func (r *fakeAdminDashboardBookingRepository) GetRevenueStats(ctx context.Context) (model.RevenueStats, error) {
	if r.revenueStatsErr != nil {
		return model.RevenueStats{}, r.revenueStatsErr
	}

	return r.revenueStats, nil
}

func (r *fakeAdminDashboardBookingRepository) GetRecentBookings(ctx context.Context, limit int) ([]model.RecentBookingActivity, error) {
	r.recentBookingsLimit = limit
	if r.recentBookingsErr != nil {
		return nil, r.recentBookingsErr
	}

	return r.recentBookings, nil
}

func (r *fakeAdminDashboardBookingRepository) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	return model.BookingAdminView{}, nil
}

func (r *fakeAdminDashboardBookingRepository) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	return nil
}
