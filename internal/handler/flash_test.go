package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
	"github.com/gorilla/sessions"
)

func TestSetAndGetFlash(t *testing.T) {
	session := sessions.NewSession(sessions.NewCookieStore([]byte("flash-test-secret")), sessionName)
	SetFlash(session, model.FlashMessage{
		Type:    model.FlashSuccess,
		Message: "Saved successfully.",
	})

	flash, ok := GetFlash(session)
	if !ok {
		t.Fatal("GetFlash() ok = false, want true")
	}
	if flash.Type != model.FlashSuccess {
		t.Fatalf("flash.Type = %q, want %q", flash.Type, model.FlashSuccess)
	}
	if flash.Message != "Saved successfully." {
		t.Fatalf("flash.Message = %q, want %q", flash.Message, "Saved successfully.")
	}

	if _, ok := GetFlash(session); ok {
		t.Fatal("GetFlash() ok = true after first read, want false")
	}
}

func TestGetFlashMissing(t *testing.T) {
	session := sessions.NewSession(sessions.NewCookieStore([]byte("flash-test-secret")), sessionName)

	if _, ok := GetFlash(session); ok {
		t.Fatal("GetFlash() ok = true, want false")
	}
}

func TestRenderReceivesAndConsumesFlash(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()

	setRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	setResponse := httptest.NewRecorder()
	err := handler.setFlash(setResponse, setRequest, model.FlashMessage{
		Type:    model.FlashWarning,
		Message: "Read once.",
	})
	if err != nil {
		t.Fatalf("setFlash() error = %v, want nil", err)
	}

	firstRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	firstRequest.AddCookie(latestCookie(t, setResponse.Result().Cookies(), sessionName))
	firstResponse := httptest.NewRecorder()

	if err := handler.render(firstResponse, firstRequest, "home.html", TemplateData{Title: "Home", AppName: "Test App"}); err != nil {
		t.Fatalf("render() first error = %v, want nil", err)
	}

	firstBody := firstResponse.Body.String()
	if !strings.Contains(firstBody, "Warning") {
		t.Fatalf("first render body does not contain flash label:\n%s", firstBody)
	}
	if !strings.Contains(firstBody, "Read once.") {
		t.Fatalf("first render body does not contain flash message:\n%s", firstBody)
	}

	secondRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	secondRequest.AddCookie(latestCookie(t, firstResponse.Result().Cookies(), sessionName))
	secondResponse := httptest.NewRecorder()

	if err := handler.render(secondResponse, secondRequest, "home.html", TemplateData{Title: "Home", AppName: "Test App"}); err != nil {
		t.Fatalf("render() second error = %v, want nil", err)
	}

	secondBody := secondResponse.Body.String()
	if strings.Contains(secondBody, "Read once.") {
		t.Fatalf("second render body contains consumed flash message:\n%s", secondBody)
	}
}

func testFlashHandler() *Handler {
	store := sessions.NewCookieStore([]byte("flash-test-secret"))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	}

	return &Handler{sessionStore: store}
}

func latestCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for i := len(cookies) - 1; i >= 0; i-- {
		if cookies[i].Name == name {
			return cookies[i]
		}
	}

	t.Fatalf("cookie %q not found", name)
	return nil
}

func chdirProjectRoot(t *testing.T) {
	t.Helper()

	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := os.Chdir("../.."); err != nil {
		t.Fatalf("Chdir project root error = %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore working directory error = %v", err)
		}
	})
}
