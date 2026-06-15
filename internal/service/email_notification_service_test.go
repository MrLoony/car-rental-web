package service

import (
	"strings"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestBuildBookingCreatedAdminMessage(t *testing.T) {
	service := newTestEmailNotificationService(t)

	message, err := service.BuildBookingCreatedAdminMessage(testEmailBooking(), testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingCreatedAdminMessage() error = %v, want nil", err)
	}

	if message.To != "admin@example.test" {
		t.Fatalf("To = %q, want %q", message.To, "admin@example.test")
	}
	assertNonEmptyEmailMessage(t, message)
	assertContains(t, message.Subject, "New booking request #42")
	assertContains(t, message.Subject, "Toyota Corolla")
	assertContains(t, message.TextBody, "New booking request #42")
	assertContains(t, message.TextBody, "Toyota Corolla")
	assertContains(t, message.HTMLBody, "New booking request #42")
	assertContains(t, message.HTMLBody, "Toyota Corolla")
}

func TestBuildBookingStatusCustomerMessage(t *testing.T) {
	service := newTestEmailNotificationService(t)

	message, err := service.BuildBookingStatusCustomerMessage(testEmailBooking(), testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingStatusCustomerMessage() error = %v, want nil", err)
	}

	if message.To != "customer@example.test" {
		t.Fatalf("To = %q, want %q", message.To, "customer@example.test")
	}
	assertNonEmptyEmailMessage(t, message)
	assertContains(t, message.Subject, "Your booking request #42 is confirmed")
	assertContains(t, message.TextBody, "Your booking request #42 is confirmed")
	assertContains(t, message.HTMLBody, "Booking request #42")
	assertContains(t, message.HTMLBody, "confirmed")
}

func TestBuildBookingCreatedAdminMessageHandlesEmptyCustomerMessage(t *testing.T) {
	service := newTestEmailNotificationService(t)
	booking := testEmailBooking()
	booking.Message = ""

	message, err := service.BuildBookingCreatedAdminMessage(booking, testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingCreatedAdminMessage() error = %v, want nil", err)
	}

	assertContains(t, message.TextBody, "Customer message: none")
	assertContains(t, message.HTMLBody, "No customer message was provided.")
}

func TestBuildBookingStatusCustomerMessageHandlesEmptyCustomerMessage(t *testing.T) {
	service := newTestEmailNotificationService(t)
	booking := testEmailBooking()
	booking.Message = ""

	message, err := service.BuildBookingStatusCustomerMessage(booking, testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingStatusCustomerMessage() error = %v, want nil", err)
	}

	assertNotContains(t, message.TextBody, "Your message:")
	assertNotContains(t, message.HTMLBody, "<h2")
}

func TestEmailNotificationHTMLRenderingEscapesUnsafeCustomerInput(t *testing.T) {
	service := newTestEmailNotificationService(t)
	booking := testEmailBooking()
	booking.CustomerName = `<script>alert("x")</script>`
	booking.Message = `<img src=x onerror="alert(1)">`

	message, err := service.BuildBookingCreatedAdminMessage(booking, testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingCreatedAdminMessage() error = %v, want nil", err)
	}

	assertNotContains(t, message.HTMLBody, `<script>alert("x")</script>`)
	assertNotContains(t, message.HTMLBody, `<img src=x onerror="alert(1)">`)
	assertContains(t, message.HTMLBody, "&lt;script&gt;")
	assertContains(t, message.HTMLBody, "&lt;img")
}

func TestEmailNotificationFormatsDateTimeAndMoney(t *testing.T) {
	service := newTestEmailNotificationService(t)

	message, err := service.BuildBookingStatusCustomerMessage(testEmailBooking(), testEmailCar())
	if err != nil {
		t.Fatalf("BuildBookingStatusCustomerMessage() error = %v, want nil", err)
	}

	assertContains(t, message.TextBody, "10 Jul 2026 09:30")
	assertContains(t, message.TextBody, "12 Jul 2026 11:00")
	assertContains(t, message.TextBody, "$270.00")
}

func newTestEmailNotificationService(t *testing.T) *EmailNotificationService {
	t.Helper()

	service, err := NewEmailNotificationServiceWithAdminEmail("admin@example.test")
	if err != nil {
		t.Fatalf("NewEmailNotificationServiceWithAdminEmail() error = %v, want nil", err)
	}

	return service
}

func assertNonEmptyEmailMessage(t *testing.T, message EmailMessage) {
	t.Helper()

	if strings.TrimSpace(message.Subject) == "" {
		t.Fatal("Subject is empty")
	}
	if strings.TrimSpace(message.TextBody) == "" {
		t.Fatal("TextBody is empty")
	}
	if strings.TrimSpace(message.HTMLBody) == "" {
		t.Fatal("HTMLBody is empty")
	}
}

func testEmailBooking() model.Booking {
	return model.Booking{
		ID:             42,
		CustomerName:   "Jane Customer",
		CustomerEmail:  "customer@example.test",
		CustomerPhone:  "555-0100",
		PickupAt:       time.Date(2026, 7, 10, 9, 30, 0, 0, time.UTC),
		ReturnAt:       time.Date(2026, 7, 12, 11, 0, 0, 0, time.UTC),
		BillingDays:    3,
		EstimatedTotal: 270,
		Message:        "Please prepare a child seat.",
		Status:         model.BookingStatusConfirmed,
	}
}

func testEmailCar() model.Car {
	return model.Car{
		Brand: "Toyota",
		Model: "Corolla",
		Year:  2024,
	}
}
