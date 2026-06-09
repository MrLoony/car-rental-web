package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func sessionOptions(isProduction bool) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isProduction,
	}
}
