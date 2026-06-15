package service

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"
	texttemplate "text/template"
	"time"

	"github.com/MrLoony/car-rental-web/internal/config"
	"github.com/MrLoony/car-rental-web/internal/model"
)

const emailTemplateDir = "web/templates/emails"

type EmailNotificationService struct {
	adminEmail    string
	textTemplates *texttemplate.Template
	htmlTemplates *htmltemplate.Template
}

type BookingNotifier interface {
	NotifyAdminBookingCreated(ctx context.Context, booking model.Booking, car model.Car) error
	NotifyCustomerBookingStatusChanged(ctx context.Context, booking model.Booking, car model.Car) error
}

type AdminBookingEmailNotifier struct {
	sender       EmailSender
	notification *EmailNotificationService
}

type emailTemplateData struct {
	BookingID      int64
	CarBrand       string
	CarModel       string
	CarYear        int
	CustomerName   string
	CustomerEmail  string
	CustomerPhone  string
	PickupAt       string
	ReturnAt       string
	BillingDays    int
	EstimatedTotal string
	Message        string
	Status         string
}

func NewEmailNotificationService(cfg config.Config) (*EmailNotificationService, error) {
	return NewEmailNotificationServiceWithAdminEmail(cfg.AdminNotificationEmail)
}

func NewEmailNotificationServiceWithAdminEmail(adminEmail string) (*EmailNotificationService, error) {
	templateDir, err := resolveEmailTemplateDir()
	if err != nil {
		return nil, err
	}

	textTemplates, err := texttemplate.ParseFiles(
		filepath.Join(templateDir, "booking_created_admin_subject.txt"),
		filepath.Join(templateDir, "booking_created_admin_text.txt"),
		filepath.Join(templateDir, "booking_status_customer_subject.txt"),
		filepath.Join(templateDir, "booking_status_customer_text.txt"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse text email templates: %w", err)
	}

	htmlTemplates, err := htmltemplate.ParseFiles(
		filepath.Join(templateDir, "booking_created_admin_html.html"),
		filepath.Join(templateDir, "booking_status_customer_html.html"),
	)
	if err != nil {
		return nil, fmt.Errorf("parse html email templates: %w", err)
	}

	return &EmailNotificationService{
		adminEmail:    adminEmail,
		textTemplates: textTemplates,
		htmlTemplates: htmlTemplates,
	}, nil
}

func resolveEmailTemplateDir() (string, error) {
	candidates := []string{
		emailTemplateDir,
		filepath.Join("..", "..", emailTemplateDir),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("email template directory not found: %s", emailTemplateDir)
}

func NewAdminBookingEmailNotifier(sender EmailSender, notification *EmailNotificationService) *AdminBookingEmailNotifier {
	return &AdminBookingEmailNotifier{
		sender:       sender,
		notification: notification,
	}
}

func (n *AdminBookingEmailNotifier) NotifyAdminBookingCreated(ctx context.Context, booking model.Booking, car model.Car) error {
	if n == nil || n.sender == nil || n.notification == nil {
		return nil
	}

	message, err := n.notification.BuildBookingCreatedAdminMessage(booking, car)
	if err != nil {
		return err
	}

	return n.sender.Send(ctx, message)
}

func (n *AdminBookingEmailNotifier) NotifyCustomerBookingStatusChanged(ctx context.Context, booking model.Booking, car model.Car) error {
	if n == nil || n.sender == nil || n.notification == nil {
		return nil
	}

	message, err := n.notification.BuildBookingStatusCustomerMessage(booking, car)
	if err != nil {
		return err
	}

	return n.sender.Send(ctx, message)
}

func (s *EmailNotificationService) BuildBookingCreatedAdminMessage(booking model.Booking, car model.Car) (EmailMessage, error) {
	data := newEmailTemplateData(booking, car)

	subject, err := s.renderTextTemplate("booking_created_admin_subject.txt", data)
	if err != nil {
		return EmailMessage{}, err
	}

	textBody, err := s.renderTextTemplate("booking_created_admin_text.txt", data)
	if err != nil {
		return EmailMessage{}, err
	}

	htmlBody, err := s.renderHTMLTemplate("booking_created_admin_html.html", data)
	if err != nil {
		return EmailMessage{}, err
	}

	return EmailMessage{
		To:       s.adminEmail,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

func (s *EmailNotificationService) BuildBookingStatusCustomerMessage(booking model.Booking, car model.Car) (EmailMessage, error) {
	data := newEmailTemplateData(booking, car)

	subject, err := s.renderTextTemplate("booking_status_customer_subject.txt", data)
	if err != nil {
		return EmailMessage{}, err
	}

	textBody, err := s.renderTextTemplate("booking_status_customer_text.txt", data)
	if err != nil {
		return EmailMessage{}, err
	}

	htmlBody, err := s.renderHTMLTemplate("booking_status_customer_html.html", data)
	if err != nil {
		return EmailMessage{}, err
	}

	return EmailMessage{
		To:       booking.CustomerEmail,
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody,
	}, nil
}

func (s *EmailNotificationService) renderTextTemplate(name string, data emailTemplateData) (string, error) {
	var buffer bytes.Buffer
	if err := s.textTemplates.ExecuteTemplate(&buffer, name, data); err != nil {
		return "", fmt.Errorf("render text email template %s: %w", name, err)
	}

	return buffer.String(), nil
}

func (s *EmailNotificationService) renderHTMLTemplate(name string, data emailTemplateData) (string, error) {
	var buffer bytes.Buffer
	if err := s.htmlTemplates.ExecuteTemplate(&buffer, name, data); err != nil {
		return "", fmt.Errorf("render html email template %s: %w", name, err)
	}

	return buffer.String(), nil
}

func newEmailTemplateData(booking model.Booking, car model.Car) emailTemplateData {
	return emailTemplateData{
		BookingID:      booking.ID,
		CarBrand:       car.Brand,
		CarModel:       car.Model,
		CarYear:        car.Year,
		CustomerName:   booking.CustomerName,
		CustomerEmail:  booking.CustomerEmail,
		CustomerPhone:  booking.CustomerPhone,
		PickupAt:       formatEmailDateTime(booking.PickupAt),
		ReturnAt:       formatEmailDateTime(booking.ReturnAt),
		BillingDays:    booking.BillingDays,
		EstimatedTotal: formatEmailMoney(booking.EstimatedTotal),
		Message:        booking.Message,
		Status:         booking.Status,
	}
}

func formatEmailDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format("02 Jan 2006 15:04")
}

func formatEmailMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}
