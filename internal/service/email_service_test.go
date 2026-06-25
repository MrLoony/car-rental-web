package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrLoony/car-rental-web/internal/config"
)

func TestNoopEmailSenderReturnsNil(t *testing.T) {
	sender := NoopEmailSender{}

	if err := sender.Send(context.Background(), EmailMessage{}); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}
}

func TestNewEmailSenderReturnsNoopWhenDisabled(t *testing.T) {
	sender := NewEmailSender(config.Config{EmailEnabled: false})

	if _, ok := sender.(NoopEmailSender); !ok {
		t.Fatalf("NewEmailSender() = %T, want NoopEmailSender", sender)
	}
}

func TestNewEmailSenderReturnsSMTPWhenEnabled(t *testing.T) {
	sender := NewEmailSender(validEmailConfig())

	if _, ok := sender.(*SMTPSender); !ok {
		t.Fatalf("NewEmailSender() = %T, want *SMTPSender", sender)
	}
}

func TestNewEmailSenderReturnsBrevoWhenConfigured(t *testing.T) {
	sender := NewEmailSender(validBrevoEmailConfig())

	if _, ok := sender.(*BrevoEmailSender); !ok {
		t.Fatalf("NewEmailSender() = %T, want *BrevoEmailSender", sender)
	}
}

func TestSMTPSenderSendRejectsInvalidMessages(t *testing.T) {
	tests := []struct {
		name    string
		message EmailMessage
		wantErr string
	}{
		{
			name: "empty recipient",
			message: EmailMessage{
				Subject:  "Subject",
				TextBody: "Body",
			},
			wantErr: "recipient",
		},
		{
			name: "empty subject",
			message: EmailMessage{
				To:       "customer@example.test",
				TextBody: "Body",
			},
			wantErr: "subject",
		},
		{
			name: "empty body",
			message: EmailMessage{
				To:      "customer@example.test",
				Subject: "Subject",
			},
			wantErr: "body",
		},
	}

	sender := NewSMTPSender(validEmailConfig())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sender.Send(context.Background(), tt.message)
			if err == nil {
				t.Fatal("Send() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Send() error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestSMTPSenderBuildsTextOnlyMessage(t *testing.T) {
	sender := NewSMTPSender(validEmailConfig())

	raw := buildTestEmailMessage(t, sender, EmailMessage{
		To:       "customer@example.test",
		Subject:  "Plain subject",
		TextBody: "Plain body",
	})

	assertContains(t, raw, "Content-Type: text/plain; charset=UTF-8")
	assertContains(t, raw, "Plain body")
	assertNotContains(t, raw, "multipart/alternative")
	assertNotContains(t, raw, "Content-Type: text/html")
}

func TestSMTPSenderBuildsHTMLOnlyMessage(t *testing.T) {
	sender := NewSMTPSender(validEmailConfig())

	raw := buildTestEmailMessage(t, sender, EmailMessage{
		To:       "customer@example.test",
		Subject:  "HTML subject",
		HTMLBody: "<p>HTML body</p>",
	})

	assertContains(t, raw, "Content-Type: text/html; charset=UTF-8")
	assertContains(t, raw, "<p>HTML body</p>")
	assertNotContains(t, raw, "multipart/alternative")
}

func TestSMTPSenderBuildsMultipartAlternativeMessage(t *testing.T) {
	sender := NewSMTPSender(validEmailConfig())

	raw := buildTestEmailMessage(t, sender, EmailMessage{
		To:       "customer@example.test",
		Subject:  "Multipart subject",
		TextBody: "Plain body",
		HTMLBody: "<p>HTML body</p>",
	})

	assertContains(t, raw, `Content-Type: multipart/alternative; boundary="car-rental-web-email-boundary"`)
	assertContains(t, raw, "Content-Type: text/plain; charset=UTF-8")
	assertContains(t, raw, "Content-Type: text/html; charset=UTF-8")
	assertContains(t, raw, "Plain body")
	assertContains(t, raw, "<p>HTML body</p>")
	assertContains(t, raw, "--car-rental-web-email-boundary--")
}

func TestSMTPSenderEncodesUTF8SubjectAndFromName(t *testing.T) {
	cfg := validEmailConfig()
	cfg.SMTPFromName = "ქირავდება მანქანა"
	sender := NewSMTPSender(cfg)

	raw := buildTestEmailMessage(t, sender, EmailMessage{
		To:       "customer@example.test",
		Subject:  "ჯავშანი მიღებულია",
		TextBody: "Body",
	})

	assertContains(t, raw, "Subject: =?utf-8?")
	assertContains(t, raw, "From: =?utf-8?")
	assertContains(t, raw, "<no-reply@example.test>")
}

func TestBrevoEmailSenderSendsExpectedRequest(t *testing.T) {
	var gotRequest brevoEmailPayload
	var gotAPIKey string
	var gotAccept string
	var gotContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}

		gotAPIKey = r.Header.Get("api-key")
		gotAccept = r.Header.Get("accept")
		gotContentType = r.Header.Get("content-type")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode Brevo payload: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"messageId":"message-id"}`))
	}))
	defer server.Close()

	sender := NewBrevoEmailSender("secret-api-key", "no-reply@example.test", "Car Rental Web", server.Client())
	sender.endpoint = server.URL

	err := sender.Send(context.Background(), EmailMessage{
		To:       "customer@example.test",
		Subject:  "Booking update",
		TextBody: "Plain text body",
		HTMLBody: "<p>HTML body</p>",
	})
	if err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if gotAPIKey != "secret-api-key" {
		t.Fatalf("api-key header = %q, want configured key", gotAPIKey)
	}
	if gotAccept != "application/json" {
		t.Fatalf("accept header = %q, want application/json", gotAccept)
	}
	if gotContentType != "application/json" {
		t.Fatalf("content-type header = %q, want application/json", gotContentType)
	}
	if gotRequest.Sender.Email != "no-reply@example.test" {
		t.Fatalf("sender email = %q, want no-reply@example.test", gotRequest.Sender.Email)
	}
	if gotRequest.Sender.Name != "Car Rental Web" {
		t.Fatalf("sender name = %q, want Car Rental Web", gotRequest.Sender.Name)
	}
	if len(gotRequest.To) != 1 || gotRequest.To[0].Email != "customer@example.test" {
		t.Fatalf("to = %#v, want customer recipient", gotRequest.To)
	}
	if gotRequest.Subject != "Booking update" {
		t.Fatalf("subject = %q, want Booking update", gotRequest.Subject)
	}
	if gotRequest.TextContent != "Plain text body" {
		t.Fatalf("textContent = %q, want text body", gotRequest.TextContent)
	}
	if gotRequest.HTMLContent != "<p>HTML body</p>" {
		t.Fatalf("htmlContent = %q, want HTML body", gotRequest.HTMLContent)
	}
}

func TestBrevoEmailSenderTreatsNon2xxAsErrorWithoutAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid key secret-api-key", http.StatusUnauthorized)
	}))
	defer server.Close()

	sender := NewBrevoEmailSender("secret-api-key", "no-reply@example.test", "Car Rental Web", server.Client())
	sender.endpoint = server.URL

	err := sender.Send(context.Background(), EmailMessage{
		To:       "customer@example.test",
		Subject:  "Booking update",
		TextBody: "Plain text body",
	})
	if err == nil {
		t.Fatal("Send() error = nil, want non-2xx error")
	}
	if !strings.Contains(err.Error(), "unexpected status 401") {
		t.Fatalf("Send() error = %q, want status code", err.Error())
	}
	if strings.Contains(err.Error(), "secret-api-key") {
		t.Fatalf("Send() error exposes API key: %q", err.Error())
	}
}

func TestBrevoEmailSenderRejectsInvalidMessages(t *testing.T) {
	sender := NewBrevoEmailSender("secret-api-key", "no-reply@example.test", "Car Rental Web", nil)

	err := sender.Send(context.Background(), EmailMessage{
		Subject:  "Missing recipient",
		TextBody: "Body",
	})
	if err == nil {
		t.Fatal("Send() error = nil, want validation error")
	}
	if !strings.Contains(err.Error(), "recipient") {
		t.Fatalf("Send() error = %q, want recipient validation", err.Error())
	}
}

func buildTestEmailMessage(t *testing.T, sender *SMTPSender, message EmailMessage) string {
	t.Helper()

	raw, err := sender.buildMessage(message)
	if err != nil {
		t.Fatalf("buildMessage() error = %v, want nil", err)
	}

	return string(raw)
}

func validEmailConfig() config.Config {
	return config.Config{
		EmailEnabled:  true,
		EmailProvider: "smtp",
		SMTPHost:      "smtp.example.test",
		SMTPPort:      587,
		SMTPFrom:      "no-reply@example.test",
		SMTPFromName:  "Car Rental Web",
	}
}

func validBrevoEmailConfig() config.Config {
	return config.Config{
		EmailEnabled:   true,
		EmailProvider:  "brevo",
		BrevoAPIKey:    "secret-api-key",
		BrevoFromEmail: "no-reply@example.test",
		BrevoFromName:  "Car Rental Web",
	}
}

func assertContains(t *testing.T, value string, expected string) {
	t.Helper()

	if !strings.Contains(value, expected) {
		t.Fatalf("value does not contain %q:\n%s", expected, value)
	}
}

func assertNotContains(t *testing.T, value string, unexpected string) {
	t.Helper()

	if strings.Contains(value, unexpected) {
		t.Fatalf("value contains %q:\n%s", unexpected, value)
	}
}
