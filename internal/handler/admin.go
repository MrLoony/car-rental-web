package handler

import "net/http"

func (h *Handler) AdminIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title:   "Admin",
			AppName: h.appName,
		}

		if err := h.render(w, r, "admin/index.html", data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) AdminCleanupPrefills() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.bookingPrefillService.CleanupExpiredPrefills(r.Context()); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}
