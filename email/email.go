// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/bnixon67/required"
)

// SMTPConfig holds configuration for an SMTP server for sending emails.
type SMTPConfig struct {
	Host     string `required:"true"` // Host address.
	Port     string `required:"true"` // Port number.
	Username string `required:"true"` // Username for authentication.
	Password string `required:"true"` // Password for authentication.
}

// RedactedSMTPConfig is a copy of SMTPConfig to hide sensitive information.
type RedactedSMTPConfig SMTPConfig

// redact creates a copy of SMTPConfig with the password field redacted.
func (s SMTPConfig) redact() RedactedSMTPConfig {
	r := RedactedSMTPConfig(s)
	if s.Password != "" {
		r.Password = "[REDACTED]"
	}
	return r
}

// MarshalJSON redacts sensitive information when marshalling to JSON.
func (s SMTPConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.redact())
}

// String returns a string for SMTPConfig with sensitive data redacted.
func (s SMTPConfig) String() string {
	return fmt.Sprintf("%+v", s.redact())
}

// IsValid verifies that SMTPConfig has all required fields populated.
func (s SMTPConfig) IsValid() (bool, error) {
	return required.ArePresent(s)
}

var (
	ErrEmailInvalidConfig    = errors.New("invalid SMTP configuration")
	ErrEmailInvalidFrom      = errors.New("invalid 'from' address")
	ErrEmailNoRecipients     = errors.New("failed to provide one recipient")
	ErrEmailInvalidRecipient = errors.New("invalid 'recipient' address")
	ErrEmailSendFailed       = errors.New("failed to send email")
)

// SendMessage sends an email using the configured SMTP server settings.
func (s SMTPConfig) SendMessage(from string, recipients []string, subject, body string) error {
	if isValid, err := s.IsValid(); !isValid || err != nil {
		return ErrEmailInvalidConfig
	}

	if _, err := mail.ParseAddress(from); err != nil {
		return fmt.Errorf("%w: %q", ErrEmailInvalidFrom, from)
	}

	if len(recipients) == 0 {
		return ErrEmailNoRecipients
	}
	for _, recipient := range recipients {
		if _, err := mail.ParseAddress(recipient); err != nil {
			return fmt.Errorf("%w: %q", ErrEmailInvalidRecipient, recipient)
		}
	}

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n",
		from, strings.Join(recipients, ", "), subject)
	message := []byte(headers + body)

	serverAddr := net.JoinHostPort(s.Host, s.Port)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)
	err := smtp.SendMail(serverAddr, auth, from, recipients, message)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrEmailSendFailed, err)
	}

	return err
}
