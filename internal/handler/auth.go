package handler

import (
	"net/http"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func (h *Handler) LoginNew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.isAdminAuthenticated(r) {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		data := TemplateData{
			Title:     "Login",
			AppName:   h.appName,
			LoginForm: model.NewLoginForm(),
		}

		if err := h.render(w, r, "auth/login.html", data); err != nil {
			h.renderServerError(w, r, err)
		}
	}
}

func (h *Handler) LoginCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		form := model.LoginForm{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		adminUser, form, err := h.authService.Authenticate(r.Context(), form)
		if err != nil {
			h.renderServerError(w, r, err)
			return
		}

		if form.HasErrors() {
			data := TemplateData{
				Title:     "Login",
				AppName:   h.appName,
				LoginForm: form,
			}
			if message := form.Errors["credentials"]; message != "" {
				data.Flash = &model.FlashMessage{
					Type:    model.FlashError,
					Message: message,
				}
			}

			if err := h.renderWithStatus(w, r, "auth/login.html", data, http.StatusUnprocessableEntity); err != nil {
				h.renderServerError(w, r, err)
			}
			return
		}

		if err := h.setAdminSession(w, r, adminUser.ID); err != nil {
			h.renderServerError(w, r, err)
			return
		}

		h.redirectWithFlash(w, r, "/admin", model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "Welcome back.",
		})
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.clearAdminSessionWithFlash(w, r, model.FlashMessage{
			Type:    model.FlashSuccess,
			Message: "You have been logged out.",
		}); err != nil {
			h.renderServerError(w, r, err)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
