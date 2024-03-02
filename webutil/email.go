// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"text/template"
)

// SMTPConfig holds SMTP server settings for email functionality.
type SMTPConfig struct {
	Host     string // Host address.
	Port     string // Port number.
	User     string // Server username.
	Password string // Server password.
}

// MailMessage contains data to include in the email template.
type MailMessage struct {
	From    string
	To      string
	Subject string
	Body    string
}

const emailTmplText = `From: {{ .From }}
To: {{ .To }}
Subject: {{ .Subject }}

{{ .Body }}
`

var (
	emailTmpl = template.Must(template.New("email").Parse(emailTmplText))

	ErrEmailInvalidSMTPConfig = errors.New("invalid SMTP configuration")
	ErrEmailSendFailed        = errors.New("failed to send email")
)

// RedactedSMTPConfig is copy of SMTPConfig that doesn't expose sensitve fields.
type RedactedSMTPConfig SMTPConfig

// redact creates a copy of s that hides sensitive data.
func (s SMTPConfig) redact() RedactedSMTPConfig {
	r := RedactedSMTPConfig(s)
	if s.Password != "" {
		r.Password = "[REDACTED]"
	}
	return r
}

// MarshalJSON customizes JSON marshalling to redact sensitive data.
func (s SMTPConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.redact())
}

// String returns a string representation of s with sensitive data redacted.
func (s SMTPConfig) String() string {
	return fmt.Sprintf("%+v", s.redact())
}

// Valid checks s and returns true if required fields are non-empty.
func (s SMTPConfig) Valid() bool {
	if s.Host == "" || s.Port == "" || s.User == "" || s.Password == "" {
		return false
	}
	return true
}

// SendMessage sends an email using the values provided.
func (s SMTPConfig) SendMessage(to, subject, body string) error {
	if !s.Valid() {
		return ErrEmailInvalidSMTPConfig
	}

	mailMessage := MailMessage{
		From:    s.User,
		To:      to,
		Subject: subject,
		Body:    body,
	}

	// Fill message template.
	message := &bytes.Buffer{}
	err := emailTmpl.Execute(message, mailMessage)
	if err != nil {
		return fmt.Errorf("failed to execute mail template: %w", err)
	}

	// Authenticate to SMTP server.
	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)

	// Send email.
	addr := net.JoinHostPort(s.Host, s.Port)
	tos := []string{mailMessage.To}
	msg := message.Bytes()
	err = smtp.SendMail(addr, auth, mailMessage.From, tos, msg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmailSendFailed, err)
	}

	return err
}
