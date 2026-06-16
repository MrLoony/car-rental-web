package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) AdminBookingsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := model.NormalizeAdminBookingStatus(r.URL.Query().Get("status"))
		page := parsePositiveInt(r.URL.Query().Get("page"), model.DefaultPage)
		filter := model.AdminBookingFilter{
			Search:  strings.TrimSpace(r.URL.Query().Get("search")),
			Status:  status,
			Page:    page,
			PerPage: model.DefaultPerPage,
		}

		bookings, pagination, err := h.bookingService.ListBookingsPage(r.Context(), filter)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:                        "Booking requests",
			AppName:                      h.appName,
			AdminBookings:                bookings,
			AdminBookingFilter:           filter,
			HasActiveAdminBookingFilters: hasActiveAdminBookingFilters(filter),
			Pagination:                   pagination,
		}
		if pagination.HasPrevious {
			data.PaginationPreviousURL = paginationURL(r, pagination.PreviousPage)
		}
		if pagination.HasNext {
			data.PaginationNextURL = paginationURL(r, pagination.NextPage)
		}

		if err := h.render(w, r, "admin/bookings/index.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func hasActiveAdminBookingFilters(filter model.AdminBookingFilter) bool {
	return filter.Search != "" ||
		model.NormalizeAdminBookingStatus(filter.Status) != model.AdminBookingStatusAll
}

func (h *Handler) AdminBookingsShow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseBookingID(w, r)
		if !ok {
			return
		}

		booking, err := h.bookingService.GetBookingByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, repository.ErrBookingNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:           "Booking #" + strconv.FormatInt(booking.ID, 10),
			AppName:         h.appName,
			AdminBooking:    booking,
			BookingStatuses: bookingStatusOptions(),
		}

		if err := h.render(w, r, "admin/bookings/show.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func (h *Handler) AdminBookingStatusUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseBookingID(w, r)
		if !ok {
			return
		}

		status := r.FormValue("status")
		err := h.bookingService.UpdateBookingStatus(r.Context(), id, status)
		if err != nil {
			if !model.IsValidBookingStatus(status) {
				http.Error(w, "Invalid booking status.", http.StatusBadRequest)
				return
			}

			if errors.Is(err, repository.ErrBookingNotFound) {
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin/bookings/"+strconv.FormatInt(id, 10), model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Booking status updated to " + status + ".",
		})
	}
}

func parseBookingID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return 0, false
	}

	return id, true
}

func bookingStatusOptions() []string {
	return []string{
		model.BookingStatusPending,
		model.BookingStatusConfirmed,
		model.BookingStatusCancelled,
		model.BookingStatusCompleted,
	}
}
