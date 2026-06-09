package main

import (
	"net/http"
	"testing"

	"github.com/gorilla/sessions"
)

func TestSessionOptionsDevelopment(t *testing.T) {
	options := sessionOptions(false)

	if options.Secure {
		t.Fatal("Secure = true, want false in development")
	}

	assertCommonSessionOptions(t, options)
}

func TestSessionOptionsProduction(t *testing.T) {
	options := sessionOptions(true)

	if !options.Secure {
		t.Fatal("Secure = false, want true in production")
	}

	assertCommonSessionOptions(t, options)
}

func assertCommonSessionOptions(t *testing.T, options *sessions.Options) {
	t.Helper()

	if options.Path != "/" {
		t.Fatalf("Path = %q, want %q", options.Path, "/")
	}
	if !options.HttpOnly {
		t.Fatal("HttpOnly = false, want true")
	}
	if options.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want %v", options.SameSite, http.SameSiteLaxMode)
	}
}
