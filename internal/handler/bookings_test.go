package handler

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestBookingPrefillTokenURL(t *testing.T) {
	got := bookingPrefillTokenURL("hyundai-elantra", "secure_token-123")
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	if parsed.Path != "/cars/hyundai-elantra/book" {
		t.Fatalf("path = %q, want %q", parsed.Path, "/cars/hyundai-elantra/book")
	}

	if parsed.Query().Get("prefill") != "secure_token-123" {
		t.Fatalf("prefill = %q, want %q", parsed.Query().Get("prefill"), "secure_token-123")
	}
}

func TestBookingPrefillTokenURLDoesNotIncludePersonalFields(t *testing.T) {
	got := bookingPrefillTokenURL("hyundai-elantra", "secure_token-123")

	for _, field := range []string{"name=", "email=", "phone=", "message=", "pickup_at=", "return_at="} {
		if strings.Contains(got, field) {
			t.Fatalf("bookingPrefillTokenURL() = %q contains personal field %q", got, field)
		}
	}
}

func TestBookingFormFromQueryKeepsBackwardCompatibility(t *testing.T) {
	request := httptest.NewRequest("GET", "/cars/hyundai-elantra/book?name=Stage+Sixteen&email=stage16%40example.com&phone=555-1605&pickup_at=2026-06-04T12%3A00&return_at=2026-06-04T18%3A00&message=carry+details", nil)

	form := bookingFormFromQuery(request)

	if form.CustomerName != "Stage Sixteen" {
		t.Fatalf("CustomerName = %q, want %q", form.CustomerName, "Stage Sixteen")
	}
	if form.CustomerEmail != "stage16@example.com" {
		t.Fatalf("CustomerEmail = %q, want %q", form.CustomerEmail, "stage16@example.com")
	}
	if form.CustomerPhone != "555-1605" {
		t.Fatalf("CustomerPhone = %q, want %q", form.CustomerPhone, "555-1605")
	}
	if form.PickupAt != "2026-06-04T12:00" {
		t.Fatalf("PickupAt = %q, want %q", form.PickupAt, "2026-06-04T12:00")
	}
	if form.ReturnAt != "2026-06-04T18:00" {
		t.Fatalf("ReturnAt = %q, want %q", form.ReturnAt, "2026-06-04T18:00")
	}
	if form.Message != "carry details" {
		t.Fatalf("Message = %q, want %q", form.Message, "carry details")
	}
	if form.Errors == nil {
		t.Fatal("Errors = nil, want initialized map")
	}
}
