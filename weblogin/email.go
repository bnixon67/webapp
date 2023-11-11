// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

// TODO: move template to file
const emailTmpl = `From: {{ .From }}
To: {{ .To }}
Subject: {{ .Subject }}

{{ .Body }}
`

// MailMessage contains data to include in the email template.
type MailMessage struct {
	From    string
	To      string
	Subject string
	Body    string
}

// SendEmail will send an email using the values provided.
func SendEmail(smtpUser, smtpPassword, smtpHost, smtpPort, to, subject, body string) error {
	mailMessage := MailMessage{
		From:    smtpUser,
		To:      to,
		Subject: subject,
		Body:    body,
	}

	// TODO: cache template
	t, err := template.New("email").Parse(emailTmpl)
	if err != nil {
		return fmt.Errorf("SendEmail: failed to parse template: %w", err)
	}

	// fill message template
	message := &bytes.Buffer{}
	err = t.Execute(message, mailMessage)
	if err != nil {
		return fmt.Errorf("SendEmail: failed to execute template: %w", err)
	}

	// authenticate to SMTP server
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	// send email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, mailMessage.From, []string{mailMessage.To}, message.Bytes())
	if err != nil {
		return fmt.Errorf("SendEmail: failed to send mail: %w", err)
	}

	return err
}
