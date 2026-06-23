package model

import "testing"

func TestNewCarForm(t *testing.T) {
	form := NewCarForm()

	if form.Errors == nil {
		t.Fatal("NewCarForm() did not initialize Errors")
	}

	if form.IsAvailable {
		t.Fatal("NewCarForm().IsAvailable = true, want false")
	}

	if form.HasErrors() {
		t.Fatal("NewCarForm().HasErrors() = true, want false")
	}
}

func TestCarFormHasErrors(t *testing.T) {
	form := NewCarForm()
	form.Errors["brand"] = "Brand is required."

	if !form.HasErrors() {
		t.Fatal("CarForm.HasErrors() = false, want true")
	}
}
