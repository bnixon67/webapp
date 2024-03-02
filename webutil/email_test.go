// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil_test

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

// Define a test struct to hold the test case data
type sendEmailTest struct {
	name       string
	smtpConfig webutil.SMTPConfig
	to         string
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
	go webutil.MockSMTPServerStart(ready,
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
			name:       "empty smtp config",
			smtpConfig: webutil.SMTPConfig{},
			wantErr:    webutil.ErrEmailInvalidSMTPConfig,
		},
		{
			name: "invalid smtp server",
			smtpConfig: webutil.SMTPConfig{
				Host:     "smtp.example.com",
				Port:     "587",
				User:     "smtpuser@example.com",
				Password: "password",
			},
			to:      "recipient@example.com",
			subject: "Greetings",
			body:    "Hello, How are you?",
			wantErr: webutil.ErrEmailSendFailed,
		},
		{
			name: "valid smtp server",
			smtpConfig: webutil.SMTPConfig{
				Host:     MockSMTPHost,
				Port:     MockSMTPPort,
				User:     "smtpuser@example.com",
				Password: "password",
			},
			to:      "recipient@example.com",
			subject: "Greetings",
			body:    "Hello, How are you?",
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.smtpConfig.SendMessage(tc.to, tc.subject, tc.body)
			if !errors.Is(err, tc.wantErr) {
				t.Errorf("SendEmail() got error = %q, want error %q", err, tc.wantErr)
			}
		})
	}
}

func TestSMTPConfigMarshalJSON(t *testing.T) {
	testCases := []struct {
		name  string
		input webutil.SMTPConfig
		want  string
	}{
		{
			name: "Password",
			input: webutil.SMTPConfig{
				Password: "supersecret",
			},
			want: `{"Host":"","Port":"","User":"","Password":"[REDACTED]"}`,
		},
		{
			name: "Host",
			input: webutil.SMTPConfig{
				Host: "host",
			},
			want: `{"Host":"host","Port":"","User":"","Password":""}`,
		},
		{
			name: "All",
			input: webutil.SMTPConfig{
				Host:     "host",
				Port:     "25",
				User:     "user",
				Password: "supersecret",
			},
			want: `{"Host":"host","Port":"25","User":"user","Password":"[REDACTED]"}`,
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
		input webutil.SMTPConfig
		want  string
	}{
		{
			name: "Password",
			input: webutil.SMTPConfig{
				Password: "supersecret",
			},
			want: `{Host: Port: User: Password:[REDACTED]}`,
		},
		{
			name: "Host",
			input: webutil.SMTPConfig{
				Host: "host",
			},
			want: `{Host:host Port: User: Password:}`,
		},
		{
			name: "All",
			input: webutil.SMTPConfig{
				Host:     "host",
				Port:     "25",
				User:     "user",
				Password: "supersecret",
			},
			want: `{Host:host Port:25 User:user Password:[REDACTED]}`,
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
