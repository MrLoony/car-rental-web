package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
)

func TestGenerateBookingPrefillToken(t *testing.T) {
	token, err := generateBookingPrefillToken()
	if err != nil {
		t.Fatalf("generateBookingPrefillToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("generateBookingPrefillToken() returned empty token")
	}

	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !pattern.MatchString(token) {
		t.Fatalf("generateBookingPrefillToken() = %q, want URL-safe token", token)
	}
}

func TestGenerateBookingPrefillTokenDiffers(t *testing.T) {
	first, err := generateBookingPrefillToken()
	if err != nil {
		t.Fatalf("generateBookingPrefillToken() first error = %v", err)
	}

	second, err := generateBookingPrefillToken()
	if err != nil {
		t.Fatalf("generateBookingPrefillToken() second error = %v", err)
	}

	if first == second {
		t.Fatal("generateBookingPrefillToken() returned duplicate tokens")
	}
}

func TestBookingFormToPrefillKeepsExpectedFields(t *testing.T) {
	expiresAt := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	form := model.BookingForm{
		CustomerName:  "Stage Sixteen",
		CustomerEmail: "stage16@example.com",
		CustomerPhone: "555-1604",
		PickupAt:      "2026-06-04T12:00",
		ReturnAt:      "2026-06-04T18:00",
		Message:       "server side state",
	}

	prefill := bookingFormToPrefill(form, "token", expiresAt)

	if prefill.Token != "token" {
		t.Fatalf("Token = %q, want %q", prefill.Token, "token")
	}
	if prefill.Name != form.CustomerName {
		t.Fatalf("Name = %q, want %q", prefill.Name, form.CustomerName)
	}
	if prefill.Email != form.CustomerEmail {
		t.Fatalf("Email = %q, want %q", prefill.Email, form.CustomerEmail)
	}
	if prefill.Phone != form.CustomerPhone {
		t.Fatalf("Phone = %q, want %q", prefill.Phone, form.CustomerPhone)
	}
	if prefill.PickupAt != form.PickupAt {
		t.Fatalf("PickupAt = %q, want %q", prefill.PickupAt, form.PickupAt)
	}
	if prefill.ReturnAt != form.ReturnAt {
		t.Fatalf("ReturnAt = %q, want %q", prefill.ReturnAt, form.ReturnAt)
	}
	if prefill.Message != form.Message {
		t.Fatalf("Message = %q, want %q", prefill.Message, form.Message)
	}
	if !prefill.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", prefill.ExpiresAt, expiresAt)
	}
}

func TestBookingPrefillToFormInitializesErrors(t *testing.T) {
	prefill := model.BookingPrefill{
		Name:     "Stage Sixteen",
		Email:    "stage16@example.com",
		Phone:    "555-1604",
		PickupAt: "2026-06-04T12:00",
		ReturnAt: "2026-06-04T18:00",
		Message:  "server side state",
	}

	form := bookingPrefillToForm(prefill)

	if form.CustomerName != prefill.Name {
		t.Fatalf("CustomerName = %q, want %q", form.CustomerName, prefill.Name)
	}
	if form.CustomerEmail != prefill.Email {
		t.Fatalf("CustomerEmail = %q, want %q", form.CustomerEmail, prefill.Email)
	}
	if form.CustomerPhone != prefill.Phone {
		t.Fatalf("CustomerPhone = %q, want %q", form.CustomerPhone, prefill.Phone)
	}
	if form.PickupAt != prefill.PickupAt {
		t.Fatalf("PickupAt = %q, want %q", form.PickupAt, prefill.PickupAt)
	}
	if form.ReturnAt != prefill.ReturnAt {
		t.Fatalf("ReturnAt = %q, want %q", form.ReturnAt, prefill.ReturnAt)
	}
	if form.Message != prefill.Message {
		t.Fatalf("Message = %q, want %q", form.Message, prefill.Message)
	}
	if form.Errors == nil {
		t.Fatal("Errors = nil, want initialized map")
	}
	if form.HasErrors() {
		t.Fatal("HasErrors() = true, want false")
	}
}

func TestGetFormByTokenBlankTokenReturnsNotFound(t *testing.T) {
	service := BookingPrefillService{}

	form, err := service.GetFormByToken(context.Background(), "   ")
	if err == nil {
		t.Fatal("GetFormByToken() error = nil, want not found")
	}
	if !errors.Is(err, repository.ErrBookingPrefillNotFound) {
		t.Fatalf("GetFormByToken() error = %v, want ErrBookingPrefillNotFound", err)
	}
	if form.Errors == nil {
		t.Fatal("Errors = nil, want initialized map")
	}
}
