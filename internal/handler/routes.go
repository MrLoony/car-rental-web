package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(appName string) http.Handler {
	r := chi.NewRouter()

	fileServer := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", Home(appName))
	r.Get("/health", Health())

	return r
}
