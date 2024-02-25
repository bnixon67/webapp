// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth

import (
	"bytes"
	"fmt"
	"net"
	"net/smtp"
	"text/template"
)

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

var emailTmpl = template.Must(template.New("email").Parse(emailTmplText))

// SendMessage sends an email using the values provided.
func (cfg ConfigSMTP) SendMessage(to, subject, body string) error {
	mailMessage := MailMessage{
		From:    cfg.User,
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

	// Authenticate to SMTP server
	auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)

	// send email
	srv := net.JoinHostPort(cfg.Host, cfg.Port)
	tos := []string{mailMessage.To}
	msg := message.Bytes()
	err = smtp.SendMail(srv, auth, mailMessage.From, tos, msg)
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return err
}
