package model

import "testing"

func TestFlashMessageCanHoldTypeAndMessage(t *testing.T) {
	flash := FlashMessage{
		Type:    FlashSuccess,
		Message: "Saved successfully.",
	}

	if flash.Type != FlashSuccess {
		t.Fatalf("Type = %q, want %q", flash.Type, FlashSuccess)
	}

	if flash.Message != "Saved successfully." {
		t.Fatalf("Message = %q, want %q", flash.Message, "Saved successfully.")
	}
}
