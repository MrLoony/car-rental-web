package handler

import "net/http"

func (h *Handler) Home() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title:   "Home",
			AppName: h.appName,
		}

		if err := h.render(w, r, "home.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}
