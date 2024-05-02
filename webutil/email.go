// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// SMTPConfig holds SMTP server settings for email functionality.
type SMTPConfig struct {
	Host     string `required:"true"` // Host address.
	Port     string `required:"true"` // Port number.
	User     string `required:"true"` // Server username.
	Password string `required:"true"` // Server password.
}

var (
	ErrEmailInvalidFrom       = errors.New("invalid from address")
	ErrEmailInvalidTo         = errors.New("invalid to address")
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
func (s SMTPConfig) SendMessage(from string, to []string, subject, body string) error {
	if !s.Valid() {
		return ErrEmailInvalidSMTPConfig
	}

	if from == "" || !strings.Contains(from, "@") {
		return fmt.Errorf("%w: %q", ErrEmailInvalidFrom, from)
	}

	if len(to) == 0 || to[0] == "" || !strings.Contains(to[0], "@") {
		return fmt.Errorf("%w: %q", ErrEmailInvalidTo, to)
	}

	// Authenticate to SMTP server.
	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n",
		from, strings.Join(to, ", "), subject)

	msg := []byte(headers + body)

	addr := net.JoinHostPort(s.Host, s.Port)

	err := smtp.SendMail(addr, auth, from, to, msg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmailSendFailed, err)
	}

	return err
}
