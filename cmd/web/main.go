package main

import (
	"context"
	"log"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/config"
	"github.com/MrLoony/car-rental-web/internal/database"
	"github.com/MrLoony/car-rental-web/internal/handler"
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

	router := handler.Routes(cfg.AppName)

	log.Printf("starting %s in %s mode on %s", cfg.AppName, cfg.AppEnv, addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
