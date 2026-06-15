package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestCalculateBillingDays(t *testing.T) {
	pickupAt := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		returnAt time.Time
		want     int
	}{
		{
			name:     "exactly 24 hours",
			returnAt: pickupAt.Add(24 * time.Hour),
			want:     1,
		},
		{
			name:     "25 hours",
			returnAt: pickupAt.Add(25 * time.Hour),
			want:     2,
		},
		{
			name:     "48 hours",
			returnAt: pickupAt.Add(48 * time.Hour),
			want:     2,
		},
		{
			name:     "49 hours",
			returnAt: pickupAt.Add(49 * time.Hour),
			want:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBillingDays(pickupAt, tt.returnAt)
			if got != tt.want {
				t.Fatalf("calculateBillingDays() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCreateBookingInitializesFormErrors(t *testing.T) {
	service := BookingService{}

	id, form, err := service.CreateBooking(context.Background(), model.Car{}, model.BookingForm{})
	if err != nil {
		t.Fatalf("CreateBooking() error = %v", err)
	}

	if id != 0 {
		t.Fatalf("CreateBooking() id = %d, want 0", id)
	}

	if !form.HasErrors() {
		t.Fatal("CreateBooking() form has no validation errors")
	}

	if form.Errors["customer_name"] == "" {
		t.Fatal("CreateBooking() did not validate customer name")
	}
}

func TestCreateBookingSendsAdminNotificationAfterSuccessfulCreate(t *testing.T) {
	bookingRepo := &fakeBookingRepository{createID: 101}
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	id, form, err := service.CreateBooking(context.Background(), testBookingCar(), validFutureBookingForm())
	if err != nil {
		t.Fatalf("CreateBooking() error = %v, want nil", err)
	}

	if id != 101 {
		t.Fatalf("CreateBooking() id = %d, want 101", id)
	}
	if form.HasErrors() {
		t.Fatalf("CreateBooking() form errors = %v, want none", form.Errors)
	}
	if !bookingRepo.createCalled {
		t.Fatal("CreateBooking() did not persist booking")
	}
	if !notifier.called {
		t.Fatal("CreateBooking() did not call notifier")
	}
	if notifier.booking.ID != 101 {
		t.Fatalf("notified booking ID = %d, want 101", notifier.booking.ID)
	}
	if notifier.booking.EstimatedTotal != 180 {
		t.Fatalf("notified EstimatedTotal = %f, want 180", notifier.booking.EstimatedTotal)
	}
	if notifier.car.ID != testBookingCar().ID {
		t.Fatalf("notified car ID = %d, want %d", notifier.car.ID, testBookingCar().ID)
	}
}

func TestCreateBookingStillSucceedsWhenAdminNotificationFails(t *testing.T) {
	bookingRepo := &fakeBookingRepository{createID: 102}
	notifier := &fakeBookingNotifier{err: errors.New("smtp unavailable")}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	id, form, err := service.CreateBooking(context.Background(), testBookingCar(), validFutureBookingForm())
	if err != nil {
		t.Fatalf("CreateBooking() error = %v, want nil", err)
	}

	if id != 102 {
		t.Fatalf("CreateBooking() id = %d, want 102", id)
	}
	if form.HasErrors() {
		t.Fatalf("CreateBooking() form errors = %v, want none", form.Errors)
	}
	if !notifier.called {
		t.Fatal("CreateBooking() did not call notifier")
	}
}

func TestCreateBookingDoesNotNotifyWhenValidationFails(t *testing.T) {
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(&fakeBookingRepository{}, &fakeBookingCarRepository{}, notifier)

	id, form, err := service.CreateBooking(context.Background(), testBookingCar(), model.BookingForm{})
	if err != nil {
		t.Fatalf("CreateBooking() error = %v, want nil", err)
	}

	if id != 0 {
		t.Fatalf("CreateBooking() id = %d, want 0", id)
	}
	if !form.HasErrors() {
		t.Fatal("CreateBooking() form has no validation errors")
	}
	if notifier.called {
		t.Fatal("CreateBooking() called notifier on validation failure")
	}
}

func TestCreateBookingDoesNotNotifyWhenAvailabilityConflict(t *testing.T) {
	bookingRepo := &fakeBookingRepository{
		hasConflict:        true,
		nextAvailableAt:    time.Now().Add(72 * time.Hour),
		nextAvailableFound: true,
	}
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	id, form, err := service.CreateBooking(context.Background(), testBookingCar(), validFutureBookingForm())
	if err != nil {
		t.Fatalf("CreateBooking() error = %v, want nil", err)
	}

	if id != 0 {
		t.Fatalf("CreateBooking() id = %d, want 0", id)
	}
	if !form.HasErrors() {
		t.Fatal("CreateBooking() form has no conflict validation error")
	}
	if notifier.called {
		t.Fatal("CreateBooking() called notifier on conflict")
	}
}

func TestAdminBookingEmailNotifierBuildsAndSendsBookingCreatedMessage(t *testing.T) {
	notificationService := newTestEmailNotificationService(t)
	sender := &fakeEmailSender{}
	notifier := NewAdminBookingEmailNotifier(sender, notificationService)

	err := notifier.NotifyAdminBookingCreated(context.Background(), testEmailBooking(), testEmailCar())
	if err != nil {
		t.Fatalf("NotifyAdminBookingCreated() error = %v, want nil", err)
	}

	if !sender.called {
		t.Fatal("NotifyAdminBookingCreated() did not call email sender")
	}
	if sender.message.To != "admin@example.test" {
		t.Fatalf("sent To = %q, want admin@example.test", sender.message.To)
	}
	assertContains(t, sender.message.Subject, "New booking request #42")
	assertContains(t, sender.message.TextBody, "Toyota Corolla")
	assertContains(t, sender.message.HTMLBody, "Toyota Corolla")
}

func TestAdminBookingEmailNotifierWorksWithNoopSender(t *testing.T) {
	notificationService := newTestEmailNotificationService(t)
	notifier := NewAdminBookingEmailNotifier(NoopEmailSender{}, notificationService)

	if err := notifier.NotifyAdminBookingCreated(context.Background(), testEmailBooking(), testEmailCar()); err != nil {
		t.Fatalf("NotifyAdminBookingCreated() error = %v, want nil", err)
	}
}

func TestCustomerBookingEmailNotifierBuildsAndSendsStatusMessage(t *testing.T) {
	notificationService := newTestEmailNotificationService(t)
	sender := &fakeEmailSender{}
	notifier := NewAdminBookingEmailNotifier(sender, notificationService)

	err := notifier.NotifyCustomerBookingStatusChanged(context.Background(), testEmailBooking(), testEmailCar())
	if err != nil {
		t.Fatalf("NotifyCustomerBookingStatusChanged() error = %v, want nil", err)
	}

	if !sender.called {
		t.Fatal("NotifyCustomerBookingStatusChanged() did not call email sender")
	}
	if sender.message.To != "customer@example.test" {
		t.Fatalf("sent To = %q, want customer@example.test", sender.message.To)
	}
	assertContains(t, sender.message.Subject, "Your booking request #42 is confirmed")
	assertContains(t, sender.message.TextBody, "Toyota Corolla")
	assertContains(t, sender.message.HTMLBody, "confirmed")
}

func TestCustomerBookingEmailNotifierWorksWithNoopSender(t *testing.T) {
	notificationService := newTestEmailNotificationService(t)
	notifier := NewAdminBookingEmailNotifier(NoopEmailSender{}, notificationService)

	if err := notifier.NotifyCustomerBookingStatusChanged(context.Background(), testEmailBooking(), testEmailCar()); err != nil {
		t.Fatalf("NotifyCustomerBookingStatusChanged() error = %v, want nil", err)
	}
}

func TestUpdateBookingStatusSendsCustomerNotificationForConfirmed(t *testing.T) {
	bookingRepo := &fakeBookingRepository{
		getBooking: testBookingAdminView(model.BookingStatusConfirmed),
	}
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	err := service.UpdateBookingStatus(context.Background(), 42, model.BookingStatusConfirmed)
	if err != nil {
		t.Fatalf("UpdateBookingStatus() error = %v, want nil", err)
	}

	if !bookingRepo.updateCalled {
		t.Fatal("UpdateBookingStatus() did not persist status")
	}
	if bookingRepo.updateID != 42 {
		t.Fatalf("updated booking ID = %d, want 42", bookingRepo.updateID)
	}
	if bookingRepo.updateStatus != model.BookingStatusConfirmed {
		t.Fatalf("updated status = %q, want %q", bookingRepo.updateStatus, model.BookingStatusConfirmed)
	}
	if !bookingRepo.getCalled {
		t.Fatal("UpdateBookingStatus() did not load notification data")
	}
	if !notifier.customerCalled {
		t.Fatal("UpdateBookingStatus() did not call customer notifier")
	}
	if notifier.statusBooking.Status != model.BookingStatusConfirmed {
		t.Fatalf("notified status = %q, want %q", notifier.statusBooking.Status, model.BookingStatusConfirmed)
	}
	if notifier.statusCar.Year != 2024 {
		t.Fatalf("notified car year = %d, want 2024", notifier.statusCar.Year)
	}
}

func TestUpdateBookingStatusSucceedsWhenCustomerNotificationFails(t *testing.T) {
	bookingRepo := &fakeBookingRepository{
		getBooking: testBookingAdminView(model.BookingStatusCancelled),
	}
	notifier := &fakeBookingNotifier{statusErr: errors.New("smtp unavailable")}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	err := service.UpdateBookingStatus(context.Background(), 42, model.BookingStatusCancelled)
	if err != nil {
		t.Fatalf("UpdateBookingStatus() error = %v, want nil", err)
	}

	if !bookingRepo.updateCalled {
		t.Fatal("UpdateBookingStatus() did not persist status")
	}
	if !notifier.customerCalled {
		t.Fatal("UpdateBookingStatus() did not attempt customer notification")
	}
}

func TestUpdateBookingStatusDoesNotNotifyWhenRepositoryUpdateFails(t *testing.T) {
	bookingRepo := &fakeBookingRepository{
		updateErr: errors.New("database unavailable"),
	}
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	err := service.UpdateBookingStatus(context.Background(), 42, model.BookingStatusCompleted)
	if err == nil {
		t.Fatal("UpdateBookingStatus() error = nil, want error")
	}

	if bookingRepo.getCalled {
		t.Fatal("UpdateBookingStatus() loaded notification data after update failure")
	}
	if notifier.customerCalled {
		t.Fatal("UpdateBookingStatus() called customer notifier after update failure")
	}
}

func TestUpdateBookingStatusSkipsCustomerNotificationForPending(t *testing.T) {
	bookingRepo := &fakeBookingRepository{}
	notifier := &fakeBookingNotifier{}
	service := NewBookingService(bookingRepo, &fakeBookingCarRepository{}, notifier)

	err := service.UpdateBookingStatus(context.Background(), 42, model.BookingStatusPending)
	if err != nil {
		t.Fatalf("UpdateBookingStatus() error = %v, want nil", err)
	}

	if !bookingRepo.updateCalled {
		t.Fatal("UpdateBookingStatus() did not persist status")
	}
	if bookingRepo.getCalled {
		t.Fatal("UpdateBookingStatus() loaded notification data for pending status")
	}
	if notifier.customerCalled {
		t.Fatal("UpdateBookingStatus() called customer notifier for pending status")
	}
}

func TestFindAvailabilityWindowsNoBlockingBookings(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(48 * time.Hour)

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, nil, 90)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	assertAvailabilityWindow(t, windows[0], requestedPickup, requestedReturn, 2, 180)
}

func TestFindAvailabilityWindowsSingleConflictReturnsNextWindowAfterBuffer(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(48 * time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: requestedReturn,
			Status:   model.BookingStatusPending,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 75)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	wantStart := requestedReturn.Add(time.Duration(model.BookingReturnBufferHours) * time.Hour)
	assertAvailabilityWindow(t, windows[0], wantStart, wantStart.Add(48*time.Hour), 2, 150)
}

func TestFindAvailabilityWindowsUsesLargeGapBetweenConflicts(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	firstReturn := requestedPickup.Add(2 * time.Hour)
	secondPickup := firstReturn.Add(time.Duration(model.BookingReturnBufferHours)*time.Hour + 30*time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: firstReturn,
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: secondPickup,
			ReturnAt: secondPickup.Add(24 * time.Hour),
			Status:   model.BookingStatusConfirmed,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 120)

	if len(windows) != 2 {
		t.Fatalf("len(windows) = %d, want 2", len(windows))
	}

	wantFirstStart := firstReturn.Add(time.Duration(model.BookingReturnBufferHours) * time.Hour)
	assertAvailabilityWindow(t, windows[0], wantFirstStart, wantFirstStart.Add(24*time.Hour), 1, 120)

	wantSecondStart := secondPickup.Add(24*time.Hour + time.Duration(model.BookingReturnBufferHours)*time.Hour)
	assertAvailabilityWindow(t, windows[1], wantSecondStart, wantSecondStart.Add(24*time.Hour), 1, 120)
}

func TestFindAvailabilityWindowsSkipsGapTooSmall(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	firstReturn := requestedPickup.Add(2 * time.Hour)
	secondPickup := firstReturn.Add(time.Duration(model.BookingReturnBufferHours)*time.Hour + 23*time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: firstReturn,
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: secondPickup,
			ReturnAt: secondPickup.Add(24 * time.Hour),
			Status:   model.BookingStatusConfirmed,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 80)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	wantStart := secondPickup.Add(24*time.Hour + time.Duration(model.BookingReturnBufferHours)*time.Hour)
	assertAvailabilityWindow(t, windows[0], wantStart, wantStart.Add(24*time.Hour), 1, 80)
}

func TestFindAvailabilityWindowsLimitsSuggestionsToThree(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(24 * time.Hour)
	blockingBookings := []model.Booking{
		{
			PickupAt: requestedPickup,
			ReturnAt: requestedPickup.Add(2 * time.Hour),
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: requestedPickup.Add(48 * time.Hour),
			ReturnAt: requestedPickup.Add(50 * time.Hour),
			Status:   model.BookingStatusPending,
		},
		{
			PickupAt: requestedPickup.Add(96 * time.Hour),
			ReturnAt: requestedPickup.Add(98 * time.Hour),
			Status:   model.BookingStatusPending,
		},
	}

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, blockingBookings, 60)

	if len(windows) != 3 {
		t.Fatalf("len(windows) = %d, want 3", len(windows))
	}
}

func TestFindAvailabilityWindowsCalculatesCeilBillingDaysAndTotal(t *testing.T) {
	requestedPickup := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	requestedReturn := requestedPickup.Add(25 * time.Hour)

	windows := findAvailabilityWindows(requestedPickup, requestedReturn, nil, 45)

	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	assertAvailabilityWindow(t, windows[0], requestedPickup, requestedReturn, 2, 90)
}

func TestAlternativeVehiclePriceRange(t *testing.T) {
	minPrice, maxPrice := alternativeVehiclePriceRange(100)

	if minPrice != 80 {
		t.Fatalf("minPrice = %f, want 80", minPrice)
	}

	if maxPrice != 120 {
		t.Fatalf("maxPrice = %f, want 120", maxPrice)
	}
}

func TestBuildVehicleSuggestionsEmptyCars(t *testing.T) {
	pickupAt := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	returnAt := pickupAt.Add(24 * time.Hour)

	suggestions := buildVehicleSuggestions(nil, pickupAt, returnAt)

	if suggestions != nil {
		t.Fatalf("suggestions = %#v, want nil", suggestions)
	}
}

func TestBuildVehicleSuggestionsCalculatesBillingAndTotals(t *testing.T) {
	pickupAt := time.Date(2026, time.June, 1, 10, 0, 0, 0, time.UTC)
	returnAt := pickupAt.Add(25 * time.Hour)
	cars := []model.Car{
		{
			ID:          10,
			Brand:       "Hyundai",
			Model:       "Elantra",
			PricePerDay: 50,
		},
		{
			ID:          11,
			Brand:       "Toyota",
			Model:       "Camry",
			PricePerDay: 65,
		},
	}

	suggestions := buildVehicleSuggestions(cars, pickupAt, returnAt)

	if len(suggestions) != 2 {
		t.Fatalf("len(suggestions) = %d, want 2", len(suggestions))
	}

	if suggestions[0].Car.ID != 10 {
		t.Fatalf("suggestions[0].Car.ID = %d, want 10", suggestions[0].Car.ID)
	}

	if suggestions[0].BillingDays != 2 {
		t.Fatalf("suggestions[0].BillingDays = %d, want 2", suggestions[0].BillingDays)
	}

	if suggestions[0].EstimatedTotal != 100 {
		t.Fatalf("suggestions[0].EstimatedTotal = %f, want 100", suggestions[0].EstimatedTotal)
	}

	if suggestions[1].BillingDays != 2 {
		t.Fatalf("suggestions[1].BillingDays = %d, want 2", suggestions[1].BillingDays)
	}

	if suggestions[1].EstimatedTotal != 130 {
		t.Fatalf("suggestions[1].EstimatedTotal = %f, want 130", suggestions[1].EstimatedTotal)
	}
}

func assertAvailabilityWindow(t *testing.T, window model.AvailabilityWindow, startAt time.Time, endAt time.Time, billingDays int, estimatedTotal float64) {
	t.Helper()

	if !window.StartAt.Equal(startAt) {
		t.Fatalf("StartAt = %v, want %v", window.StartAt, startAt)
	}

	if !window.EndAt.Equal(endAt) {
		t.Fatalf("EndAt = %v, want %v", window.EndAt, endAt)
	}

	if window.BillingDays != billingDays {
		t.Fatalf("BillingDays = %d, want %d", window.BillingDays, billingDays)
	}

	if window.EstimatedTotal != estimatedTotal {
		t.Fatalf("EstimatedTotal = %f, want %f", window.EstimatedTotal, estimatedTotal)
	}
}

func validFutureBookingForm() model.BookingForm {
	pickupAt := time.Now().Add(48 * time.Hour).Truncate(time.Minute)
	returnAt := pickupAt.Add(48 * time.Hour)

	return model.BookingForm{
		CustomerName:  "Jane Customer",
		CustomerEmail: "jane@example.test",
		CustomerPhone: "555-0100",
		PickupAt:      pickupAt.Format(datetimeLocalLayout),
		ReturnAt:      returnAt.Format(datetimeLocalLayout),
		Message:       "Please prepare the car.",
	}
}

func testBookingCar() model.Car {
	return model.Car{
		ID:          7,
		CategoryID:  2,
		Brand:       "Toyota",
		Model:       "Corolla",
		Year:        2024,
		PricePerDay: 90,
	}
}

func testBookingAdminView(status string) model.BookingAdminView {
	return model.BookingAdminView{
		ID:             42,
		CarID:          7,
		CarBrand:       "Toyota",
		CarModel:       "Corolla",
		CarSlug:        "toyota-corolla",
		CarYear:        2024,
		CustomerName:   "Jane Customer",
		CustomerEmail:  "customer@example.test",
		CustomerPhone:  "555-0100",
		PickupAt:       time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC),
		ReturnAt:       time.Date(2026, time.July, 12, 11, 0, 0, 0, time.UTC),
		BillingDays:    3,
		EstimatedTotal: 270,
		Message:        "Please prepare a child seat.",
		Status:         status,
	}
}

type fakeBookingRepository struct {
	createCalled       bool
	createID           int64
	createdBooking     model.Booking
	hasConflict        bool
	nextAvailableAt    time.Time
	nextAvailableFound bool
	updateCalled       bool
	updateID           int64
	updateStatus       string
	updateErr          error
	getCalled          bool
	getBooking         model.BookingAdminView
	getErr             error
}

func (r *fakeBookingRepository) CreateBooking(ctx context.Context, booking model.Booking) (int64, error) {
	r.createCalled = true
	r.createdBooking = booking
	if r.createID == 0 {
		return 1, nil
	}

	return r.createID, nil
}

func (r *fakeBookingRepository) HasBookingConflict(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (bool, error) {
	return r.hasConflict, nil
}

func (r *fakeBookingRepository) FindNextAvailablePickupAt(ctx context.Context, carID int64, pickupAt time.Time, returnAt time.Time, bufferHours int) (time.Time, bool, error) {
	return r.nextAvailableAt, r.nextAvailableFound, nil
}

func (r *fakeBookingRepository) ListBlockingBookingsForCar(ctx context.Context, carID int64, from time.Time, to time.Time) ([]model.Booking, error) {
	return nil, nil
}

func (r *fakeBookingRepository) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeBookingRepository) CountBookings(ctx context.Context, filter model.AdminBookingFilter) (int, error) {
	return 0, nil
}

func (r *fakeBookingRepository) ListBookingsPage(ctx context.Context, filter model.AdminBookingFilter, pagination model.Pagination) ([]model.BookingAdminView, error) {
	return nil, nil
}

func (r *fakeBookingRepository) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	r.getCalled = true
	if r.getErr != nil {
		return model.BookingAdminView{}, r.getErr
	}

	return r.getBooking, nil
}

func (r *fakeBookingRepository) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	r.updateCalled = true
	r.updateID = id
	r.updateStatus = status
	return r.updateErr
}

type fakeBookingCarRepository struct{}

func (r *fakeBookingCarRepository) ListAvailableAlternativeCars(
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

type fakeBookingNotifier struct {
	called         bool
	booking        model.Booking
	car            model.Car
	err            error
	customerCalled bool
	statusBooking  model.Booking
	statusCar      model.Car
	statusErr      error
}

func (n *fakeBookingNotifier) NotifyAdminBookingCreated(ctx context.Context, booking model.Booking, car model.Car) error {
	n.called = true
	n.booking = booking
	n.car = car
	return n.err
}

func (n *fakeBookingNotifier) NotifyCustomerBookingStatusChanged(ctx context.Context, booking model.Booking, car model.Car) error {
	n.customerCalled = true
	n.statusBooking = booking
	n.statusCar = car
	return n.statusErr
}

type fakeEmailSender struct {
	called  bool
	message EmailMessage
}

func (s *fakeEmailSender) Send(ctx context.Context, message EmailMessage) error {
	s.called = true
	s.message = message
	return nil
}
