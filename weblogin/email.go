// Copyright 2023 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package weblogin

import (
	"bytes"
	"fmt"
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

// SendEmail sends an email using the values provided.
func SendEmail(smtpConfig ConfigSMTP, mailMessage MailMessage) error {
	mailMessage.From = smtpConfig.User

	// Fill message template.
	message := &bytes.Buffer{}
	err := emailTmpl.Execute(message, mailMessage)
	if err != nil {
		return fmt.Errorf("failed to execute mail template: %w", err)
	}

	// Authenticate to SMTP server
	auth := smtp.PlainAuth("", smtpConfig.User, smtpConfig.Password, smtpConfig.Host)

	// send email
	err = smtp.SendMail(smtpConfig.Host+":"+smtpConfig.Port, auth, mailMessage.From, []string{mailMessage.To}, message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return err
}
