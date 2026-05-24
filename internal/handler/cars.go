package handler

import (
	"errors"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) CarsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := model.CarFilter{
			Search:       r.URL.Query().Get("search"),
			CategorySlug: r.URL.Query().Get("category"),
			FuelType:     r.URL.Query().Get("fuel"),
			Transmission: r.URL.Query().Get("transmission"),
			Sort:         model.NormalizeCarSort(r.URL.Query().Get("sort")),
		}

		cars, err := h.carService.ListCars(r.Context(), filter)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		categories, err := h.categoryService.ListCategories(r.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := TemplateData{
			Title:            "Cars",
			AppName:          h.appName,
			Cars:             cars,
			Filter:           filter,
			Categories:       categories,
			FuelTypes:        []string{"Gasoline", "Hybrid", "Diesel"},
			Transmissions:    []string{"Automatic", "Manual"},
			HasActiveFilters: hasActiveCarFilters(filter),
		}

		if err := h.render(w, r, "cars/index.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func hasActiveCarFilters(filter model.CarFilter) bool {
	return filter.Search != "" ||
		filter.CategorySlug != "" ||
		filter.FuelType != "" ||
		filter.Transmission != "" ||
		(filter.Sort != "" && filter.Sort != model.SortNewest)
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

		if err := h.render(w, r, "cars/show.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
