package handler

import (
	"net/url"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestBookingPrefillURLEmptyForm(t *testing.T) {
	got := bookingPrefillURL("hyundai-elantra", model.BookingForm{})
	want := "/cars/hyundai-elantra/book"

	if got != want {
		t.Fatalf("bookingPrefillURL() = %q, want %q", got, want)
	}
}

func TestBookingPrefillURLEncodesFormValues(t *testing.T) {
	form := model.BookingForm{
		CustomerName:  "Stage Fifteen Seven",
		CustomerEmail: "stage15-7@example.com",
		CustomerPhone: "+1 555 1570",
		PickupAt:      "2026-06-04T12:00",
		ReturnAt:      "2026-06-04T18:00",
		Message:       "Carry details & notes",
	}

	got := bookingPrefillURL("hyundai-elantra", form)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	if parsed.Path != "/cars/hyundai-elantra/book" {
		t.Fatalf("path = %q, want %q", parsed.Path, "/cars/hyundai-elantra/book")
	}

	query := parsed.Query()
	assertQueryValue(t, query, "name", form.CustomerName)
	assertQueryValue(t, query, "email", form.CustomerEmail)
	assertQueryValue(t, query, "phone", form.CustomerPhone)
	assertQueryValue(t, query, "pickup_at", form.PickupAt)
	assertQueryValue(t, query, "return_at", form.ReturnAt)
	assertQueryValue(t, query, "message", form.Message)
}

func TestBookingPrefillURLOmitsEmptyValues(t *testing.T) {
	form := model.BookingForm{
		CustomerEmail: "stage15-7@example.com",
		PickupAt:      "2026-06-04T12:00",
	}

	got := bookingPrefillURL("hyundai-elantra", form)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	query := parsed.Query()
	assertQueryValue(t, query, "email", form.CustomerEmail)
	assertQueryValue(t, query, "pickup_at", form.PickupAt)

	for _, key := range []string{"name", "phone", "return_at", "message"} {
		if query.Has(key) {
			t.Fatalf("query contains %q, want omitted", key)
		}
	}
}

func assertQueryValue(t *testing.T, query url.Values, key string, want string) {
	t.Helper()

	got := query.Get(key)
	if got != want {
		t.Fatalf("query[%q] = %q, want %q", key, got, want)
	}
}
