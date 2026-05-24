package handler

import "net/http"

const (
	sessionName           = "car_rental_admin_session"
	sessionAdminUserIDKey = "admin_user_id"
)

func (h *Handler) setAdminSession(w http.ResponseWriter, r *http.Request, adminUserID int64) error {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	session.Values[sessionAdminUserIDKey] = adminUserID
	return session.Save(r, w)
}

func (h *Handler) clearAdminSession(w http.ResponseWriter, r *http.Request) error {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return err
	}

	delete(session.Values, sessionAdminUserIDKey)
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func (h *Handler) isAdminAuthenticated(r *http.Request) bool {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return false
	}

	switch value := session.Values[sessionAdminUserIDKey].(type) {
	case int:
		return value > 0
	case int64:
		return value > 0
	case int32:
		return value > 0
	case float64:
		return value > 0 && value == float64(int64(value))
	default:
		return false
	}
}

func (h *Handler) RequireAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.isAdminAuthenticated(r) {
			next.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}
