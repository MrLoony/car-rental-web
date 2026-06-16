package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/MrLoony/car-rental-web/internal/service"
)

func TestRenderNotFoundReturnsCustom404Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"

	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	response := httptest.NewRecorder()

	handler.renderNotFound(response, request)

	assertNotFoundPage(t, response)
}

func TestRoutesFallbackReturnsCustom404Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"

	request := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	assertNotFoundPage(t, response)
}

func TestMissingVehicleReturnsCustom404Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.carService = service.NewCarService(&fakeHandlerCarRepository{
		getBySlugErr: repository.ErrCarNotFound,
	})

	request := requestWithParam(http.MethodGet, "/cars/missing-car", "slug", "missing-car")
	response := httptest.NewRecorder()

	handler.CarsShow().ServeHTTP(response, request)

	assertNotFoundPage(t, response)
}

func TestMissingBookingReturnsCustom404Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(
		&fakeHandlerBookingRepository{getErr: repository.ErrBookingNotFound},
		&fakeHandlerBookingCarRepository{},
		&fakeHandlerBookingNotifier{},
	)

	request := requestWithParam(http.MethodGet, "/admin/bookings/42", "id", "42")
	response := httptest.NewRecorder()

	handler.AdminBookingsShow().ServeHTTP(response, request)

	assertNotFoundPage(t, response)
}

func assertNotFoundPage(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	body := response.Body.String()
	if !strings.Contains(body, "Page Not Found") {
		t.Fatalf("body does not contain 404 heading:\n%s", body)
	}
	if !strings.Contains(body, "Back to Home") {
		t.Fatalf("body does not contain home link:\n%s", body)
	}
	if !strings.Contains(body, "Browse Cars") {
		t.Fatalf("body does not contain cars link:\n%s", body)
	}
}
