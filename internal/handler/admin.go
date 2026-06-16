package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func (h *Handler) AdminIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title:   "Admin",
			AppName: h.appName,
		}

		if err := h.render(w, r, "admin/index.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func (h *Handler) AdminCleanupPrefills() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.bookingPrefillService.CleanupExpiredPrefills(r.Context()); err != nil {
			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin", model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Expired prefill tokens cleaned successfully.",
		})
	}
}
