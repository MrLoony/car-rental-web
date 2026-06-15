package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/smtp"
	"strings"

	"github.com/MrLoony/car-rental-web/internal/config"
)

type EmailMessage struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

type EmailSender interface {
	Send(ctx context.Context, message EmailMessage) error
}

type NoopEmailSender struct{}

func (NoopEmailSender) Send(ctx context.Context, message EmailMessage) error {
	return nil
}

type SMTPSender struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

func NewEmailSender(cfg config.Config) EmailSender {
	if !cfg.EmailEnabled {
		return NoopEmailSender{}
	}

	return NewSMTPSender(cfg)
}

func NewSMTPSender(cfg config.Config) *SMTPSender {
	return &SMTPSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.SMTPFrom,
		fromName: cfg.SMTPFromName,
	}
}

func (s *SMTPSender) Send(ctx context.Context, message EmailMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validateEmailMessage(message); err != nil {
		return err
	}

	body, err := s.buildMessage(message)
	if err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(address, auth, s.from, []string{strings.TrimSpace(message.To)}, body); err != nil {
		return fmt.Errorf("send email via smtp: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

func (s *SMTPSender) buildMessage(message EmailMessage) ([]byte, error) {
	if err := validateEmailMessage(message); err != nil {
		return nil, err
	}

	var builder strings.Builder
	writeCommonEmailHeaders(&builder, s, message)

	switch {
	case message.TextBody != "" && message.HTMLBody != "":
		writeMultipartAlternativeEmail(&builder, message)
	case message.HTMLBody != "":
		writeSinglePartEmail(&builder, "text/html", message.HTMLBody)
	default:
		writeSinglePartEmail(&builder, "text/plain", message.TextBody)
	}

	return []byte(builder.String()), nil
}

func validateEmailMessage(message EmailMessage) error {
	if strings.TrimSpace(message.To) == "" {
		return errors.New("email recipient is required")
	}

	if strings.TrimSpace(message.Subject) == "" {
		return errors.New("email subject is required")
	}

	if message.TextBody == "" && message.HTMLBody == "" {
		return errors.New("email body is required")
	}

	return nil
}

func writeCommonEmailHeaders(builder *strings.Builder, sender *SMTPSender, message EmailMessage) {
	writeEmailHeader(builder, "From", formatEmailFromHeader(sender.fromName, sender.from))
	writeEmailHeader(builder, "To", strings.TrimSpace(message.To))
	writeEmailHeader(builder, "Subject", encodeEmailHeader(message.Subject))
	writeEmailHeader(builder, "MIME-Version", "1.0")
}

func writeSinglePartEmail(builder *strings.Builder, contentType string, body string) {
	writeEmailHeader(builder, "Content-Type", contentType+"; charset=UTF-8")
	writeEmailHeader(builder, "Content-Transfer-Encoding", "quoted-printable")
	builder.WriteString("\r\n")
	builder.WriteString(encodeQuotedPrintable(body))
}

func writeMultipartAlternativeEmail(builder *strings.Builder, message EmailMessage) {
	const boundary = "car-rental-web-email-boundary"

	writeEmailHeader(builder, "Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	builder.WriteString("\r\n")

	writeEmailPart(builder, boundary, "text/plain", message.TextBody)
	writeEmailPart(builder, boundary, "text/html", message.HTMLBody)
	builder.WriteString("--" + boundary + "--\r\n")
}

func writeEmailPart(builder *strings.Builder, boundary string, contentType string, body string) {
	builder.WriteString("--" + boundary + "\r\n")
	writeEmailHeader(builder, "Content-Type", contentType+"; charset=UTF-8")
	writeEmailHeader(builder, "Content-Transfer-Encoding", "quoted-printable")
	builder.WriteString("\r\n")
	builder.WriteString(encodeQuotedPrintable(body))
	builder.WriteString("\r\n")
}

func writeEmailHeader(builder *strings.Builder, key string, value string) {
	builder.WriteString(key)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\r\n")
}

func formatEmailFromHeader(fromName string, from string) string {
	from = strings.TrimSpace(from)
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		return from
	}

	return fmt.Sprintf("%s <%s>", encodeEmailHeader(fromName), from)
}

func encodeEmailHeader(value string) string {
	return mime.QEncoding.Encode("utf-8", strings.TrimSpace(value))
}

func encodeQuotedPrintable(value string) string {
	var buffer bytes.Buffer
	writer := quotedprintable.NewWriter(&buffer)
	_, _ = writer.Write([]byte(value))
	_ = writer.Close()

	return buffer.String()
}
