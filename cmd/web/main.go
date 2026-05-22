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
	appHandler := handler.New(cfg.AppName, carService)
	router := appHandler.Routes()

	log.Printf("starting %s in %s mode on %s", cfg.AppName, cfg.AppEnv, addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
