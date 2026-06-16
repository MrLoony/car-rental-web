package handler

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/service"
)

func TestAdminBookingsExportRequiresAdminAuth(t *testing.T) {
	handler := testFlashHandler()

	request := httptest.NewRequest(http.MethodGet, "/admin/bookings/export.csv", nil)
	response := httptest.NewRecorder()

	handler.Routes().ServeHTTP(response, request)

	assertRedirect(t, response, "/login")
}

func TestAdminBookingsExportWritesCSV(t *testing.T) {
	bookingRepo := &fakeHandlerBookingRepository{
		exportRows: []model.BookingExportRow{
			{
				ID:             42,
				Status:         model.BookingStatusPending,
				CustomerName:   "Jane Customer",
				CustomerEmail:  "jane@example.test",
				CustomerPhone:  "555-0100",
				Car:            "Toyota Corolla",
				PickupAt:       time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC),
				ReturnAt:       time.Date(2026, time.July, 12, 11, 0, 0, 0, time.UTC),
				BillingDays:    3,
				EstimatedTotal: 270,
				CreatedAt:      time.Date(2026, time.June, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	handler := testFlashHandler()
	handler.bookingService = service.NewBookingService(bookingRepo, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin/bookings/export.csv?search=%20jane%20&status=PENDING&page=3", nil)
	response := httptest.NewRecorder()

	handler.AdminBookingsExport().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/csv; charset=utf-8", got)
	}
	if got := response.Header().Get("Content-Disposition"); got != `attachment; filename="bookings.csv"` {
		t.Fatalf("Content-Disposition = %q", got)
	}
	if bookingRepo.exportFilter.Search != "jane" {
		t.Fatalf("exportFilter.Search = %q, want jane", bookingRepo.exportFilter.Search)
	}
	if bookingRepo.exportFilter.Status != model.BookingStatusPending {
		t.Fatalf("exportFilter.Status = %q, want pending", bookingRepo.exportFilter.Status)
	}

	records, err := csv.NewReader(strings.NewReader(response.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("read CSV error = %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("CSV record count = %d, want 2", len(records))
	}

	wantHeader := []string{
		"ID",
		"Status",
		"Customer Name",
		"Customer Email",
		"Customer Phone",
		"Car",
		"Pickup At",
		"Return At",
		"Billing Days",
		"Estimated Total",
		"Created At",
	}
	for i, want := range wantHeader {
		if records[0][i] != want {
			t.Fatalf("header[%d] = %q, want %q", i, records[0][i], want)
		}
	}

	row := records[1]
	if row[0] != "42" || row[1] != model.BookingStatusPending || row[2] != "Jane Customer" {
		t.Fatalf("unexpected CSV row: %#v", row)
	}
	if row[5] != "Toyota Corolla" {
		t.Fatalf("CSV car = %q, want Toyota Corolla", row[5])
	}
	if row[8] != "3" {
		t.Fatalf("CSV billing days = %q, want 3", row[8])
	}
	if row[9] != "270.00" {
		t.Fatalf("CSV estimated total = %q, want 270.00", row[9])
	}
}

func TestAdminBookingsExportFailureReturnsCustom500Page(t *testing.T) {
	chdirProjectRoot(t)
	bookingRepo := &fakeHandlerBookingRepository{exportErr: errors.New("database password leaked")}
	handler := testFlashHandler()
	handler.appName = "Test App"
	handler.bookingService = service.NewBookingService(bookingRepo, &fakeHandlerBookingCarRepository{}, nil)

	request := httptest.NewRequest(http.MethodGet, "/admin/bookings/export.csv", nil)
	response := httptest.NewRecorder()

	handler.AdminBookingsExport().ServeHTTP(response, request)

	assertServerErrorPage(t, response)
}

func TestAdminBookingExportURLPreservesFilters(t *testing.T) {
	got := adminBookingExportURL(model.AdminBookingFilter{
		Search: "Jane Customer",
		Status: model.BookingStatusConfirmed,
		Page:   3,
	})

	if got != "/admin/bookings/export.csv?search=Jane+Customer&status=confirmed" {
		t.Fatalf("adminBookingExportURL() = %q", got)
	}
}
