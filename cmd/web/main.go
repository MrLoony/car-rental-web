package main

import (
	"context"
	"log"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/config"
	"github.com/MrLoony/car-rental-web/internal/database"
	"github.com/MrLoony/car-rental-web/internal/handler"
	"github.com/MrLoony/car-rental-web/internal/repository"
	"github.com/MrLoony/car-rental-web/internal/service"
	"github.com/gorilla/sessions"
)

func main() {
	cfg := config.Load()
	addr := ":" + cfg.AppPort

	ctx := context.Background()
	dbpool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer dbpool.Close()

	log.Println("database connection established")

	carRepository := repository.NewCarRepository(dbpool)
	carService := service.NewCarService(carRepository)

	categoryRepository := repository.NewCategoryRepository(dbpool)
	categoryService := service.NewCategoryService(categoryRepository)

	bookingRepository := repository.NewBookingRepository(dbpool)
	bookingService := service.NewBookingService(bookingRepository, carRepository)

	bookingPrefillRepository := repository.NewBookingPrefillRepository(dbpool)
	bookingPrefillService := service.NewBookingPrefillService(bookingPrefillRepository)

	adminUserRepository := repository.NewAdminUserRepository(dbpool)
	loginAttemptLimiter := service.NewLoginAttemptLimiter()
	authService := service.NewAuthService(adminUserRepository, loginAttemptLimiter)

	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options = sessionOptions(cfg.IsProduction)

	appHandler := handler.New(cfg.AppName, carService, categoryService, bookingService, bookingPrefillService, authService, sessionStore, cfg.IsProduction)
	router := appHandler.Routes()

	log.Printf("starting %s in %s mode on %s", cfg.AppName, cfg.AppEnv, addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
