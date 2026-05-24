package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

type Handler struct {
	appName         string
	carService      *service.CarService
	categoryService *service.CategoryService
	bookingService  *service.BookingService
	authService     *service.AuthService
	sessionStore    *sessions.CookieStore
}

func New(appName string, carService *service.CarService, categoryService *service.CategoryService, bookingService *service.BookingService, authService *service.AuthService, sessionStore *sessions.CookieStore) *Handler {
	return &Handler{
		appName:         appName,
		carService:      carService,
		categoryService: categoryService,
		bookingService:  bookingService,
		authService:     authService,
		sessionStore:    sessionStore,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.Home())
	r.Get("/health", h.Health())
	r.Get("/login", h.LoginNew())
	r.Post("/login", h.LoginCreate())
	r.Post("/logout", h.Logout())
	r.Get("/cars", h.CarsIndex())
	r.Get("/cars/{slug}/book", h.BookingNew())
	r.Post("/cars/{slug}/book", h.BookingCreate())
	r.Get("/cars/{slug}", h.CarsShow())
	r.Get("/bookings/success", h.BookingSuccess())

	r.Group(func(r chi.Router) {
		r.Use(h.RequireAdminAuth)
		r.Get("/admin", h.AdminIndex())
		r.Get("/admin/bookings", h.AdminBookingsIndex())
		r.Get("/admin/bookings/{id}", h.AdminBookingsShow())
		r.Post("/admin/bookings/{id}/status", h.AdminBookingStatusUpdate())
	})

	return r
}
