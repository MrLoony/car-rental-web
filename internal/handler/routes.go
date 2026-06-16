package handler

import (
	"context"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

type bookingPrefillService interface {
	CreateFromBookingForm(ctx context.Context, form model.BookingForm) (string, error)
	GetFormByToken(ctx context.Context, token string) (model.BookingForm, error)
	CleanupExpiredPrefills(ctx context.Context) error
}

type Handler struct {
	appName               string
	carService            *service.CarService
	categoryService       *service.CategoryService
	bookingService        *service.BookingService
	bookingPrefillService bookingPrefillService
	authService           *service.AuthService
	sessionStore          *sessions.CookieStore
	isProduction          bool
}

func New(appName string, carService *service.CarService, categoryService *service.CategoryService, bookingService *service.BookingService, bookingPrefillService *service.BookingPrefillService, authService *service.AuthService, sessionStore *sessions.CookieStore, isProduction bool) *Handler {
	return &Handler{
		appName:               appName,
		carService:            carService,
		categoryService:       categoryService,
		bookingService:        bookingService,
		bookingPrefillService: bookingPrefillService,
		authService:           authService,
		sessionStore:          sessionStore,
		isProduction:          isProduction,
	}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(h.SecurityHeaders)
	r.NotFound(h.NotFound())

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.Home())
	r.Get("/health", h.Health())
	r.Get("/login", h.LoginNew())
	r.With(h.RequireCSRF).Post("/login", h.LoginCreate())
	r.With(h.RequireCSRF).Post("/logout", h.Logout())
	r.Get("/cars", h.CarsIndex())
	r.Get("/cars/{slug}/book", h.BookingNew())
	r.With(h.RequireCSRF).Post("/cars/{slug}/book", h.BookingCreate())
	r.Get("/cars/{slug}", h.CarsShow())
	r.Get("/bookings/success", h.BookingSuccess())

	r.Group(func(r chi.Router) {
		r.Use(h.RequireAdminAuth)
		r.Get("/admin", h.AdminIndex())
		r.With(h.RequireCSRF).Post("/admin/cleanup/prefills", h.AdminCleanupPrefills())
		r.Get("/admin/cars", h.AdminCarsIndex())
		r.Get("/admin/cars/new", h.AdminCarsNew())
		r.With(h.RequireCSRF).Post("/admin/cars", h.AdminCarsCreate())
		r.Get("/admin/cars/{id}/edit", h.AdminCarsEdit())
		r.With(h.RequireCSRF).Post("/admin/cars/{id}", h.AdminCarsUpdate())
		r.With(h.RequireCSRF).Post("/admin/cars/{id}/availability", h.AdminCarAvailabilityUpdate())
		r.With(h.RequireCSRF).Post("/admin/cars/{id}/archive", h.AdminCarArchive())
		r.With(h.RequireCSRF).Post("/admin/cars/{id}/unarchive", h.AdminCarUnarchive())
		r.Get("/admin/cars/{id}", h.AdminCarsShow())
		r.Get("/admin/bookings", h.AdminBookingsIndex())
		r.Get("/admin/bookings/export.csv", h.AdminBookingsExport())
		r.Get("/admin/bookings/{id}", h.AdminBookingsShow())
		r.With(h.RequireCSRF).Post("/admin/bookings/{id}/status", h.AdminBookingStatusUpdate())
	})

	return r
}
