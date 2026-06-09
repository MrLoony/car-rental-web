package handler

import "net/http"

const contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data: https: http:; font-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'self'; object-src 'none';"

func (h *Handler) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)

		if h.isProduction {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}
