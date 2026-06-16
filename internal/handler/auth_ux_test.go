package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
)

func TestLoginCredentialsFlashDoesNotDuplicateInlineError(t *testing.T) {
	chdirProjectRoot(t)
	handler := testFlashHandler()

	data := TemplateData{
		Title:   "Login",
		AppName: "Test App",
		LoginForm: model.LoginForm{
			Email: "admin@example.com",
			Errors: map[string]string{
				"credentials": "Invalid email or password.",
			},
		},
		Flash: &model.FlashMessage{
			Type:    model.FlashError,
			Message: "Invalid email or password.",
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	response := httptest.NewRecorder()

	if err := handler.renderWithStatus(response, request, "auth/login.html", data, http.StatusUnprocessableEntity); err != nil {
		t.Fatalf("renderWithStatus() error = %v, want nil", err)
	}

	if count := strings.Count(response.Body.String(), "Invalid email or password."); count != 1 {
		t.Fatalf("credentials message count = %d, want 1", count)
	}
}
