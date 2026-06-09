package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
)

func TestGenerateCSRFToken(t *testing.T) {
	token, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("generateCSRFToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("generateCSRFToken() returned empty token")
	}

	pattern := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	if !pattern.MatchString(token) {
		t.Fatalf("generateCSRFToken() = %q, want URL-safe token", token)
	}
}

func TestGenerateCSRFTokensDiffer(t *testing.T) {
	first, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("generateCSRFToken() first error = %v", err)
	}

	second, err := generateCSRFToken()
	if err != nil {
		t.Fatalf("generateCSRFToken() second error = %v", err)
	}

	if first == second {
		t.Fatal("generateCSRFToken() returned duplicate tokens")
	}
}

func TestValidateCSRFTokenMatchingToken(t *testing.T) {
	handler := testCSRFHandler()
	token, cookies := issueCSRFToken(t, handler)

	request := csrfPostRequest(token, cookies)
	if !handler.validateCSRFToken(request) {
		t.Fatal("validateCSRFToken() = false, want true")
	}
}

func TestValidateCSRFTokenMissingToken(t *testing.T) {
	handler := testCSRFHandler()
	_, cookies := issueCSRFToken(t, handler)

	request := csrfPostRequest("", cookies)
	if handler.validateCSRFToken(request) {
		t.Fatal("validateCSRFToken() = true, want false")
	}
}

func TestValidateCSRFTokenMismatchedToken(t *testing.T) {
	handler := testCSRFHandler()
	_, cookies := issueCSRFToken(t, handler)

	request := csrfPostRequest("wrong-token", cookies)
	if handler.validateCSRFToken(request) {
		t.Fatal("validateCSRFToken() = true, want false")
	}
}

func TestValidateCSRFTokenMissingSessionToken(t *testing.T) {
	handler := testCSRFHandler()

	request := csrfPostRequest("submitted-token", nil)
	if handler.validateCSRFToken(request) {
		t.Fatal("validateCSRFToken() = true, want false")
	}
}

func TestRequireCSRFSafeMethodReachesNext(t *testing.T) {
	handler := testCSRFHandler()
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.RequireCSRF(next).ServeHTTP(response, request)

	if !called {
		t.Fatal("next handler was not called")
	}
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRequireCSRFMissingTokenReturnsForbidden(t *testing.T) {
	handler := testCSRFHandler()
	_, cookies := issueCSRFToken(t, handler)

	request := csrfPostRequest("", cookies)
	response := httptest.NewRecorder()

	handler.RequireCSRF(unexpectedCSRFNextHandler(t)).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestRequireCSRFWrongTokenReturnsForbidden(t *testing.T) {
	handler := testCSRFHandler()
	_, cookies := issueCSRFToken(t, handler)

	request := csrfPostRequest("wrong-token", cookies)
	response := httptest.NewRecorder()

	handler.RequireCSRF(unexpectedCSRFNextHandler(t)).ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestRequireCSRFValidTokenReachesNext(t *testing.T) {
	handler := testCSRFHandler()
	token, cookies := issueCSRFToken(t, handler)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	})

	request := csrfPostRequest(token, cookies)
	response := httptest.NewRecorder()

	handler.RequireCSRF(next).ServeHTTP(response, request)

	if !called {
		t.Fatal("next handler was not called")
	}
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
}

func testCSRFHandler() *Handler {
	store := sessions.NewCookieStore([]byte("csrf-test-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	}

	return &Handler{sessionStore: store}
}

func issueCSRFToken(t *testing.T, handler *Handler) (string, []*http.Cookie) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	token, err := handler.getCSRFToken(response, request)
	if err != nil {
		t.Fatalf("getCSRFToken() error = %v", err)
	}

	return token, response.Result().Cookies()
}

func csrfPostRequest(token string, cookies []*http.Cookie) *http.Request {
	form := url.Values{}
	if token != "" {
		form.Set(csrfFormField, token)
	}

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}

	return request
}

func unexpectedCSRFNextHandler(t *testing.T) http.Handler {
	t.Helper()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})
}
