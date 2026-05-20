package main

import (
	"log"
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/config"
	"github.com/MrLoony/car-rental-web/internal/handler"
)

func main() {
	cfg := config.Load()
	addr := ":" + cfg.AppPort

	router := handler.Routes(cfg.AppName)

	log.Printf("starting %s in %s mode on %s", cfg.AppName, cfg.AppEnv, addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
