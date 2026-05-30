package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) BookingNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		car, ok := h.loadBookingCar(w, r)
		if !ok {
			return
		}

		data := TemplateData{
			Title:       "Book " + car.Brand + " " + car.Model,
			AppName:     h.appName,
			Car:         car,
			BookingForm: bookingFormFromQuery(r),
		}

		if err := h.render(w, r, "bookings/new.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) BookingCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		car, ok := h.loadBookingCar(w, r)
		if !ok {
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := model.BookingForm{
			CustomerName:  r.FormValue("customer_name"),
			CustomerEmail: r.FormValue("customer_email"),
			CustomerPhone: r.FormValue("customer_phone"),
			PickupAt:      r.FormValue("pickup_at"),
			ReturnAt:      r.FormValue("return_at"),
			Message:       r.FormValue("message"),
		}

		bookingID, form, err := h.bookingService.CreateBooking(r.Context(), car, form)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if form.HasErrors() {
			data := TemplateData{
				Title:       "Book " + car.Brand + " " + car.Model,
				AppName:     h.appName,
				Car:         car,
				BookingForm: form,
			}

			if err := h.renderWithStatus(w, r, "bookings/new.html", data, http.StatusUnprocessableEntity); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, "/bookings/success?id="+strconv.FormatInt(bookingID, 10), http.StatusSeeOther)
	}
}

func (h *Handler) BookingSuccess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var bookingID int64
		if id := r.URL.Query().Get("id"); id != "" {
			parsedID, err := strconv.ParseInt(id, 10, 64)
			if err == nil {
				bookingID = parsedID
			}
		}

		data := TemplateData{
			Title:     "Booking request sent",
			AppName:   h.appName,
			BookingID: bookingID,
		}

		if err := h.render(w, r, "bookings/success.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) loadBookingCar(w http.ResponseWriter, r *http.Request) (model.Car, bool) {
	slug := chi.URLParam(r, "slug")

	car, err := h.carService.GetCarBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrCarNotFound) {
			http.Error(w, "car not found", http.StatusNotFound)
			return model.Car{}, false
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return model.Car{}, false
	}

	return car, true
}

func bookingFormFromQuery(r *http.Request) model.BookingForm {
	form := model.NewBookingForm()
	query := r.URL.Query()

	form.CustomerName = query.Get("name")
	form.CustomerEmail = query.Get("email")
	form.CustomerPhone = query.Get("phone")
	form.PickupAt = query.Get("pickup_at")
	form.ReturnAt = query.Get("return_at")
	form.Message = query.Get("message")

	return form
}
