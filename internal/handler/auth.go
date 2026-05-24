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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if form.HasErrors() {
			data := TemplateData{
				Title:     "Login",
				AppName:   h.appName,
				LoginForm: form,
			}

			if err := h.renderWithStatus(w, r, "auth/login.html", data, http.StatusUnprocessableEntity); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		if err := h.setAdminSession(w, r, adminUser.ID); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.clearAdminSession(w, r); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
