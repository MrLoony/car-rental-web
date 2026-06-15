package service

import (
	"context"
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
		EmailEnabled: true,
		SMTPHost:     "smtp.example.test",
		SMTPPort:     587,
		SMTPFrom:     "no-reply@example.test",
		SMTPFromName: "Car Rental Web",
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
