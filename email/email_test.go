// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package email_test

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/bnixon67/webapp/email"
)

// Define a test struct to hold the test case data
type sendEmailTest struct {
	name       string
	smtpConfig email.SMTPConfig
	from       string
	to         []string
	subject    string
	body       string
	wantErr    error
}

const (
	MockSMTPHost = "localhost"
	MockSMTPPort = "2525"
)

func TestMain(m *testing.M) {
	// Create a channel to signal when the server is ready.
	ready := make(chan bool)

	// Start the mock SMTP server in a goroutine.
	go email.MockSMTPServerStart(ready,
		net.JoinHostPort(MockSMTPHost, MockSMTPPort))

	// Wait for the server to signal it is ready.
	<-ready

	// Continue with test.
	os.Exit(m.Run())
}

// TestSendEmail runs table-driven tests for the SendEmail function
func TestSendEmail(t *testing.T) {
	tests := []sendEmailTest{
		{
			name:       "emptyConfig",
			smtpConfig: email.SMTPConfig{},
			wantErr:    email.ErrEmailInvalidConfig,
		},
		{
			name: "invalidServer",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			from:    "from@example.com",
			to:      []string{"to@example.com"},
			wantErr: email.ErrEmailSendFailed,
		},
		{
			name: "emptyFrom",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			to:      []string{"to@example.com"},
			wantErr: email.ErrEmailInvalidFrom,
		},
		{
			name: "invalidFrom",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			to:      []string{"to"},
			wantErr: email.ErrEmailInvalidFrom,
		},
		{
			name: "emptyTo",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			from:    "from@example.com",
			wantErr: email.ErrEmailNoRecipients,
		},
		{
			name: "invalidTo0",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			from:    "from@example.com",
			to:      []string{"to"},
			wantErr: email.ErrEmailInvalidRecipient,
		},
		{
			name: "invalidTo1",
			smtpConfig: email.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				Username: "smtpuser@example.com",
				Password: "password",
			},
			from:    "from@example.com",
			to:      []string{"from@exmaple.com", "to"},
			wantErr: email.ErrEmailInvalidRecipient,
		},
		{
			name: "validMessage",
			smtpConfig: email.SMTPConfig{
				Host:     MockSMTPHost,
				Port:     MockSMTPPort,
				Username: "smtpuser@example.com",
				Password: "password",
			},
			from:    "from@example.com",
			to:      []string{"recipient@example.com"},
			subject: "Greetings",
			body:    "Hello, How are you?",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.smtpConfig.SendMessage(tc.from, tc.to, tc.subject, tc.body)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("SendEmail() got error = %q, want error %q", err, tc.wantErr)
			}
		})
	}
}

func TestSMTPConfigMarshalJSON(t *testing.T) {
	testCases := []struct {
		name  string
		input email.SMTPConfig
		want  string
	}{
		{
			name: "Password",
			input: email.SMTPConfig{
				Password: "supersecret",
			},
			want: `{"Host":"","Port":"","Username":"","Password":"[REDACTED]"}`,
		},
		{
			name: "Host",
			input: email.SMTPConfig{
				Host: "host",
			},
			want: `{"Host":"host","Port":"","Username":"","Password":""}`,
		},
		{
			name: "All",
			input: email.SMTPConfig{
				Host:     "host",
				Port:     "25",
				Username: "user",
				Password: "supersecret",
			},
			want: `{"Host":"host","Port":"25","Username":"user","Password":"[REDACTED]"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.input.MarshalJSON()
			if err != nil {
				t.Fatalf("Error during MarshalJSON: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("unequal output\ngot  %q\nwant %q\n",
					string(got), tc.want)
			}
		})
	}
}

func TestSMTPConfigString(t *testing.T) {
	testCases := []struct {
		name  string
		input email.SMTPConfig
		want  string
	}{
		{
			name: "Password",
			input: email.SMTPConfig{
				Password: "supersecret",
			},
			want: `{Host: Port: Username: Password:[REDACTED]}`,
		},
		{
			name: "Host",
			input: email.SMTPConfig{
				Host: "host",
			},
			want: `{Host:host Port: Username: Password:}`,
		},
		{
			name: "All",
			input: email.SMTPConfig{
				Host:     "host",
				Port:     "25",
				Username: "user",
				Password: "supersecret",
			},
			want: `{Host:host Port:25 Username:user Password:[REDACTED]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			if got != tc.want {
				t.Errorf("unequal output\ngot  %q\nwant %q\n",
					got, tc.want)
			}
		})
	}
}
