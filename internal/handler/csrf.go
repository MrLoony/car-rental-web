package handler

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	csrfSessionKey = "csrf_token"
	csrfFormField  = "csrf_token"
	csrfTokenBytes = 32
)

func generateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func (h *Handler) getCSRFToken(w http.ResponseWriter, r *http.Request) (string, error) {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return "", err
	}

	if token, ok := session.Values[csrfSessionKey].(string); ok && token != "" {
		return token, nil
	}

	token, err := generateCSRFToken()
	if err != nil {
		return "", err
	}

	session.Values[csrfSessionKey] = token
	if err := session.Save(r, w); err != nil {
		return "", err
	}

	return token, nil
}

func (h *Handler) validateCSRFToken(r *http.Request) bool {
	session, err := h.sessionStore.Get(r, sessionName)
	if err != nil {
		return false
	}

	sessionToken, ok := session.Values[csrfSessionKey].(string)
	if !ok || sessionToken == "" {
		return false
	}

	submittedToken := r.FormValue(csrfFormField)
	if submittedToken == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(sessionToken), []byte(submittedToken)) == 1
}

func (h *Handler) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		if !h.validateCSRFToken(r) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
