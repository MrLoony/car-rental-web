package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/MrLoony/car-rental-web/internal/model"
)

type TemplateData struct {
	Title                        string
	AppName                      string
	Cars                         []model.Car
	Car                          model.Car
	Filter                       model.CarFilter
	Categories                   []model.Category
	FuelTypes                    []string
	Transmissions                []string
	HasActiveFilters             bool
	Pagination                   model.Pagination
	PaginationPreviousURL        string
	PaginationNextURL            string
	BookingForm                  model.BookingForm
	BookingID                    int64
	AdminBookings                []model.BookingAdminView
	AdminBooking                 model.BookingAdminView
	AdminBookingFilter           model.AdminBookingFilter
	HasActiveAdminBookingFilters bool
	AdminCars                    []model.Car
	AdminCar                     model.Car
	AdminCarFilter               model.AdminCarFilter
	HasActiveAdminCarFilters     bool
	BookingStatuses              []string
	LoginForm                    model.LoginForm
	CarForm                      model.CarForm
	IsAdminAuthenticated         bool
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, page string, data TemplateData) error {
	return h.renderWithStatus(w, r, page, data, http.StatusOK)
}

func (h *Handler) renderWithStatus(w http.ResponseWriter, r *http.Request, page string, data TemplateData, status int) error {
	data.IsAdminAuthenticated = h.isAdminAuthenticated(r)
	return renderWithStatus(w, page, data, status)
}

func render(w http.ResponseWriter, page string, data TemplateData) error {
	return renderWithStatus(w, page, data, http.StatusOK)
}

func renderWithStatus(w http.ResponseWriter, page string, data TemplateData, status int) error {
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatDateTime":    formatDateTime,
		"formatMoney":       formatMoney,
		"bookingPrefillURL": bookingPrefillURL,
	}).ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/"+page,
	)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	return err
}

func formatDateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format("Jan 02, 2006 15:04")
}

func formatMoney(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func bookingPrefillURL(slug string, form model.BookingForm) string {
	path := "/cars/" + url.PathEscape(slug) + "/book"
	values := url.Values{}

	if form.CustomerName != "" {
		values.Set("name", form.CustomerName)
	}
	if form.CustomerEmail != "" {
		values.Set("email", form.CustomerEmail)
	}
	if form.CustomerPhone != "" {
		values.Set("phone", form.CustomerPhone)
	}
	if form.PickupAt != "" {
		values.Set("pickup_at", form.PickupAt)
	}
	if form.ReturnAt != "" {
		values.Set("return_at", form.ReturnAt)
	}
	if form.Message != "" {
		values.Set("message", form.Message)
	}

	encoded := values.Encode()
	if encoded == "" {
		return path
	}

	return path + "?" + encoded
}
