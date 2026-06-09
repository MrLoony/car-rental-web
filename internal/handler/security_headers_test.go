package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeadersDevelopment(t *testing.T) {
	handler := &Handler{isProduction: false}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sample", nil)

	handler.SecurityHeaders(sampleSecurityHeadersHandler()).ServeHTTP(response, request)

	assertSecurityHeaders(t, response)
	if got := response.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("Strict-Transport-Security = %q, want empty in development", got)
	}
}

func TestSecurityHeadersProduction(t *testing.T) {
	handler := &Handler{isProduction: true}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sample", nil)

	handler.SecurityHeaders(sampleSecurityHeadersHandler()).ServeHTTP(response, request)

	assertSecurityHeaders(t, response)
	if got := response.Header().Get("Strict-Transport-Security"); got != "max-age=31536000; includeSubDomains" {
		t.Fatalf("Strict-Transport-Security = %q, want production HSTS header", got)
	}
}

func assertSecurityHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	if got := response.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want %q", got, "nosniff")
	}
	if got := response.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options = %q, want %q", got, "DENY")
	}
	if got := response.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Fatalf("Referrer-Policy = %q, want %q", got, "strict-origin-when-cross-origin")
	}

	csp := response.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Fatal("Content-Security-Policy is empty")
	}

	for _, directive := range []string{
		"default-src 'self'",
		"script-src 'self'",
		"style-src 'self'",
		"img-src 'self' data: https: http:",
		"frame-ancestors 'none'",
		"object-src 'none'",
	} {
		if !strings.Contains(csp, directive) {
			t.Fatalf("Content-Security-Policy = %q, missing directive %q", csp, directive)
		}
	}
}

func sampleSecurityHeadersHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
}
