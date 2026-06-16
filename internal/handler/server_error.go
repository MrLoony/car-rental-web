package handler

import (
	"log"
	"net/http"
)

func (h *Handler) renderServerError(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		log.Printf("internal server error: %v", err)
	}

	data := TemplateData{
		Title:   "Something went wrong",
		AppName: h.appName,
	}

	if renderErr := h.renderWithStatus(w, r, "server_error.html", data, http.StatusInternalServerError); renderErr != nil {
		log.Printf("internal server error page failed: %v", renderErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
