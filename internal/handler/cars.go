package handler

import (
	"errors"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) CarsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cars, err := h.carService.ListAvailableCars(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := TemplateData{
			Title:   "Cars",
			AppName: h.appName,
			Cars:    cars,
		}

		if err := render(w, "cars/index.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) CarsShow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")

		car, err := h.carService.GetCarBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, repository.ErrCarNotFound) {
				http.Error(w, "car not found", http.StatusNotFound)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := TemplateData{
			Title:   car.Brand + " " + car.Model,
			AppName: h.appName,
			Car:     car,
		}

		if err := render(w, "cars/show.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
