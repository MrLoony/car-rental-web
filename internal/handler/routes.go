package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	appName    string
	carService *service.CarService
}

func New(appName string, carService *service.CarService) *Handler {
	return &Handler{
		appName:    appName,
		carService: carService,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.Home())
	r.Get("/health", h.Health())
	r.Get("/cars", h.CarsIndex())
	r.Get("/cars/{slug}", h.CarsShow())

	return r
}
