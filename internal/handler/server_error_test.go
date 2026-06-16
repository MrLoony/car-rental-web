package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/service"
)

func TestRenderServerErrorReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"

	request := httptest.NewRequest(http.MethodGet, "/broken", nil)
	response := httptest.NewRecorder()

	handler.renderServerError(response, request, errors.New("database password leaked"))

	assertServerErrorPage(t, response)
}

func TestCarsIndexFailureReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.carService = service.NewCarService(&fakeHandlerCarRepository{
		countErr: errors.New("database password leaked"),
	})

	request := httptest.NewRequest(http.MethodGet, "/cars", nil)
	response := httptest.NewRecorder()

	handler.CarsIndex().ServeHTTP(response, request)

	assertServerErrorPage(t, response)
}

func assertServerErrorPage(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}

	body := response.Body.String()
	if !strings.Contains(body, "Something went wrong") {
		t.Fatalf("body does not contain server error heading:\n%s", body)
	}
	if !strings.Contains(body, "Please try again later") {
		t.Fatalf("body does not contain generic message:\n%s", body)
	}
	if strings.Contains(body, "database password leaked") {
		t.Fatalf("body exposes original error:\n%s", body)
	}
}
