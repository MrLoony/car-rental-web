package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
)

const datetimeLocalLayout = "2006-01-02T15:04"

type BookingService struct {
	repo *repository.BookingRepository
}

func NewBookingService(repo *repository.BookingRepository) *BookingService {
	return &BookingService{repo: repo}
}

func (s *BookingService) CreateBooking(ctx context.Context, car model.Car, form model.BookingForm) (int64, model.BookingForm, error) {
	form = normalizeBookingForm(form)

	pickupAt, returnAt := validateBookingForm(&form)
	if form.HasErrors() {
		return 0, form, nil
	}

	hasConflict, err := s.repo.HasBookingConflict(ctx, car.ID, pickupAt, returnAt, model.BookingReturnBufferHours)
	if err != nil {
		return 0, form, fmt.Errorf("check booking availability: %w", err)
	}

	if hasConflict {
		suggestedAt, found, err := s.repo.FindNextAvailablePickupAt(ctx, car.ID, pickupAt, returnAt, model.BookingReturnBufferHours)
		if err != nil {
			return 0, form, fmt.Errorf("find next available pickup time: %w", err)
		}
		if found {
			form.SuggestedPickupAt = suggestedAt.Format("Jan 02, 2006 15:04")
		}

		searchFrom := pickupAt.Add(-time.Duration(model.BookingReturnBufferHours) * time.Hour)
		searchTo := returnAt.AddDate(0, 0, 30)
		blockingBookings, err := s.repo.ListBlockingBookingsForCar(ctx, car.ID, searchFrom, searchTo)
		if err != nil {
			return 0, form, fmt.Errorf("list blocking bookings for availability windows: %w", err)
		}

		form.SuggestedAvailabilityWindows = findAvailabilityWindows(pickupAt, returnAt, blockingBookings, car.PricePerDay)
		form.Errors["pickup_at"] = "This car is unavailable for the selected period. Please choose another pickup or return time."
		return 0, form, nil
	}

	billingDays := calculateBillingDays(pickupAt, returnAt)
	booking := model.Booking{
		CarID:          car.ID,
		CustomerName:   form.CustomerName,
		CustomerEmail:  form.CustomerEmail,
		CustomerPhone:  form.CustomerPhone,
		PickupAt:       pickupAt,
		ReturnAt:       returnAt,
		BillingDays:    billingDays,
		EstimatedTotal: float64(billingDays) * car.PricePerDay,
		Message:        form.Message,
		Status:         model.BookingStatusPending,
	}

	id, err := s.repo.CreateBooking(ctx, booking)
	if err != nil {
		return 0, form, fmt.Errorf("create booking: %w", err)
	}

	return id, form, nil
}

func (s *BookingService) ListBookings(ctx context.Context) ([]model.BookingAdminView, error) {
	bookings, err := s.repo.ListBookings(ctx)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}

	return bookings, nil
}

func (s *BookingService) ListBookingsPage(ctx context.Context, filter model.AdminBookingFilter) ([]model.BookingAdminView, model.Pagination, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	filter.Status = model.NormalizeAdminBookingStatus(filter.Status)

	totalCount, err := s.repo.CountBookings(ctx, filter)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("count bookings: %w", err)
	}

	pagination := model.NewPagination(filter.Page, filter.PerPage, totalCount)
	bookings, err := s.repo.ListBookingsPage(ctx, filter, pagination)
	if err != nil {
		return nil, model.Pagination{}, fmt.Errorf("list bookings page: %w", err)
	}

	return bookings, pagination, nil
}

func (s *BookingService) GetBookingByID(ctx context.Context, id int64) (model.BookingAdminView, error) {
	booking, err := s.repo.GetBookingByID(ctx, id)
	if err != nil {
		return model.BookingAdminView{}, fmt.Errorf("get booking by id: %w", err)
	}

	return booking, nil
}

func (s *BookingService) UpdateBookingStatus(ctx context.Context, id int64, status string) error {
	if !model.IsValidBookingStatus(status) {
		return fmt.Errorf("invalid booking status: %s", status)
	}

	if err := s.repo.UpdateBookingStatus(ctx, id, status); err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}

	return nil
}

func normalizeBookingForm(form model.BookingForm) model.BookingForm {
	if form.Errors == nil {
		form.Errors = make(map[string]string)
	}

	form.CustomerName = strings.TrimSpace(form.CustomerName)
	form.CustomerEmail = strings.TrimSpace(form.CustomerEmail)
	form.CustomerPhone = strings.TrimSpace(form.CustomerPhone)
	form.PickupAt = strings.TrimSpace(form.PickupAt)
	form.ReturnAt = strings.TrimSpace(form.ReturnAt)
	form.Message = strings.TrimSpace(form.Message)

	return form
}

func validateBookingForm(form *model.BookingForm) (time.Time, time.Time) {
	if form.CustomerName == "" {
		form.Errors["customer_name"] = "Name is required."
	}

	if form.CustomerEmail == "" {
		form.Errors["customer_email"] = "Email is required."
	} else if !strings.Contains(form.CustomerEmail, "@") || !strings.Contains(form.CustomerEmail, ".") {
		form.Errors["customer_email"] = "Enter a valid email address."
	}

	if form.CustomerPhone == "" {
		form.Errors["customer_phone"] = "Phone number is required."
	}

	pickupAt, pickupOK := parseRequiredDatetime(form.PickupAt, "pickup_at", "Pickup time is required.", form.Errors)
	returnAt, returnOK := parseRequiredDatetime(form.ReturnAt, "return_at", "Return time is required.", form.Errors)

	if pickupOK && pickupAt.Before(time.Now()) {
		form.Errors["pickup_at"] = "Pickup time cannot be in the past."
	}

	if pickupOK && returnOK && !returnAt.After(pickupAt) {
		form.Errors["return_at"] = "Return time must be after pickup time."
	}

	return pickupAt, returnAt
}

func parseRequiredDatetime(value, field, requiredMessage string, errors map[string]string) (time.Time, bool) {
	if value == "" {
		errors[field] = requiredMessage
		return time.Time{}, false
	}

	parsed, err := time.ParseInLocation(datetimeLocalLayout, value, time.Local)
	if err != nil {
		errors[field] = "Enter a valid date and time."
		return time.Time{}, false
	}

	return parsed, true
}

func calculateBillingDays(pickupAt, returnAt time.Time) int {
	duration := returnAt.Sub(pickupAt)
	billingDays := int(math.Ceil(duration.Hours() / 24))
	if billingDays < 1 {
		return 1
	}

	return billingDays
}

func findAvailabilityWindows(requestedPickup time.Time, requestedReturn time.Time, blockingBookings []model.Booking, pricePerDay float64) []model.AvailabilityWindow {
	const maxSuggestions = 3

	requestedDuration := requestedReturn.Sub(requestedPickup)
	if requestedDuration <= 0 {
		return nil
	}

	windows := make([]model.AvailabilityWindow, 0, maxSuggestions)
	cursor := requestedPickup

	addWindow := func(start time.Time) {
		if len(windows) >= maxSuggestions {
			return
		}

		end := start.Add(requestedDuration)
		billingDays := calculateBillingDays(start, end)
		windows = append(windows, model.AvailabilityWindow{
			StartAt:        start,
			EndAt:          end,
			BillingDays:    billingDays,
			EstimatedTotal: float64(billingDays) * pricePerDay,
		})
	}

	for _, booking := range blockingBookings {
		if len(windows) >= maxSuggestions {
			break
		}

		blockedStart := booking.PickupAt
		blockedEnd := booking.ReturnAt.Add(time.Duration(model.BookingReturnBufferHours) * time.Hour)

		if cursor.Before(blockedStart) {
			gapDuration := blockedStart.Sub(cursor)
			if gapDuration >= requestedDuration {
				addWindow(cursor)
			}
		}

		if blockedEnd.After(cursor) {
			cursor = blockedEnd
		}
	}

	addWindow(cursor)

	return windows
}
