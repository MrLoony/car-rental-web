package handler

import (
	"fmt"
	"net/http"
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	}
}
