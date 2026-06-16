package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) CarsIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := parsePositiveInt(r.URL.Query().Get("page"), model.DefaultPage)
		filter := model.CarFilter{
			Search:       r.URL.Query().Get("search"),
			CategorySlug: r.URL.Query().Get("category"),
			FuelType:     r.URL.Query().Get("fuel"),
			Transmission: r.URL.Query().Get("transmission"),
			Sort:         model.NormalizeCarSort(r.URL.Query().Get("sort")),
			Page:         page,
			PerPage:      6,
		}

		cars, pagination, err := h.carService.ListCarsPage(r.Context(), filter)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		categories, err := h.categoryService.ListCategories(r.Context())
		if err != nil {
			h.renderServerError(w, r, err)
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
			Pagination:       pagination,
		}
		if pagination.HasPrevious {
			data.PaginationPreviousURL = paginationURL(r, pagination.PreviousPage)
		}
		if pagination.HasNext {
			data.PaginationNextURL = paginationURL(r, pagination.NextPage)
		}

		if err := h.render(w, r, "cars/index.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func parsePositiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}

	return parsed
}

func paginationURL(r *http.Request, page int) string {
	values := r.URL.Query()
	values.Set("page", strconv.Itoa(page))
	return r.URL.Path + "?" + values.Encode()
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
				h.renderNotFound(w, r)
				return
			}

			h.renderServerError(w, r, err)
			return
		}

		data := TemplateData{
			Title:   car.Brand + " " + car.Model,
			AppName: h.appName,
			Car:     car,
		}

		if err := h.render(w, r, "cars/show.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}
