package model

import "testing"

func TestNewLoginForm(t *testing.T) {
	form := NewLoginForm()

	if form.Errors == nil {
		t.Fatal("NewLoginForm() did not initialize Errors")
	}

	if form.HasErrors() {
		t.Fatal("NewLoginForm().HasErrors() = true, want false")
	}
}

func TestLoginFormHasErrors(t *testing.T) {
	form := NewLoginForm()
	form.Errors["email"] = "Email is required."

	if !form.HasErrors() {
		t.Fatal("LoginForm.HasErrors() = false, want true")
	}
}
