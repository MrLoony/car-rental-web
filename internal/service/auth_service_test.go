package service

import (
	"testing"

	"github.com/MrLoony/car-rental-web/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateLoginForm(t *testing.T) {
	form := normalizeLoginForm(model.LoginForm{})
	validateLoginForm(&form)

	if form.Errors["email"] == "" {
		t.Fatal("validateLoginForm() did not validate required email")
	}

	if form.Errors["password"] == "" {
		t.Fatal("validateLoginForm() did not validate required password")
	}
}

func TestPasswordMatchesHashWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if passwordMatchesHash("wrong-password", string(hash)) {
		t.Fatal("passwordMatchesHash() = true for wrong password")
	}
}

func TestPasswordMatchesHashCorrectPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	if !passwordMatchesHash("admin123", string(hash)) {
		t.Fatal("passwordMatchesHash() = false for correct password")
	}
}
