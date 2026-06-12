package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
)

const (
	bookingPrefillTokenBytes = 32
	bookingPrefillTTL        = 30 * time.Minute
)

type BookingPrefillService struct {
	repo *repository.BookingPrefillRepository
}

func NewBookingPrefillService(repo *repository.BookingPrefillRepository) *BookingPrefillService {
	return &BookingPrefillService{repo: repo}
}

func (s *BookingPrefillService) CreateFromBookingForm(ctx context.Context, form model.BookingForm) (string, error) {
	if err := s.repo.DeleteExpiredBookingPrefills(ctx); err != nil {
		return "", fmt.Errorf("delete expired booking prefills: %w", err)
	}

	token, err := generateBookingPrefillToken()
	if err != nil {
		return "", fmt.Errorf("generate booking prefill token: %w", err)
	}

	prefill := bookingFormToPrefill(form, token, time.Now().Add(bookingPrefillTTL))
	if err := s.repo.CreateBookingPrefill(ctx, prefill); err != nil {
		return "", fmt.Errorf("create booking prefill: %w", err)
	}

	return token, nil
}

func (s *BookingPrefillService) GetFormByToken(ctx context.Context, token string) (model.BookingForm, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return model.NewBookingForm(), fmt.Errorf("get booking prefill by token: %w", repository.ErrBookingPrefillNotFound)
	}

	prefill, err := s.repo.GetBookingPrefillByToken(ctx, token)
	if err != nil {
		return model.NewBookingForm(), fmt.Errorf("get booking prefill by token: %w", err)
	}

	return bookingPrefillToForm(prefill), nil
}

func (s *BookingPrefillService) DeleteExpired(ctx context.Context) error {
	if err := s.repo.DeleteExpiredBookingPrefills(ctx); err != nil {
		return fmt.Errorf("delete expired booking prefills: %w", err)
	}

	return nil
}

func (s *BookingPrefillService) CleanupExpiredPrefills(ctx context.Context) error {
	return s.DeleteExpired(ctx)
}

func generateBookingPrefillToken() (string, error) {
	bytes := make([]byte, bookingPrefillTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func bookingFormToPrefill(form model.BookingForm, token string, expiresAt time.Time) model.BookingPrefill {
	return model.BookingPrefill{
		Token:     token,
		Name:      form.CustomerName,
		Email:     form.CustomerEmail,
		Phone:     form.CustomerPhone,
		PickupAt:  form.PickupAt,
		ReturnAt:  form.ReturnAt,
		Message:   form.Message,
		ExpiresAt: expiresAt,
	}
}

func bookingPrefillToForm(prefill model.BookingPrefill) model.BookingForm {
	form := model.NewBookingForm()
	form.CustomerName = prefill.Name
	form.CustomerEmail = prefill.Email
	form.CustomerPhone = prefill.Phone
	form.PickupAt = prefill.PickupAt
	form.ReturnAt = prefill.ReturnAt
	form.Message = prefill.Message

	return form
}
