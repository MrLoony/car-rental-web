package handler

import "net/http"

func (h *Handler) NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.renderNotFound(w, r)
	}
}

func (h *Handler) renderNotFound(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Title:   "Page Not Found",
		AppName: h.appName,
	}

	if err := h.renderWithStatus(w, r, "not_found.html", data, http.StatusNotFound); err != nil {
		h.renderServerError(w, r, err)
	}
}
