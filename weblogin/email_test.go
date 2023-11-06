package weblogin_test

import (
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

// Define a test struct to hold the test case data
type sendEmailTest struct {
	smtpUser     string
	smtpPassword string
	smtpHost     string
	smtpPort     string
	to           string
	subject      string
	body         string
	wantErr      bool
}

// TestSendEmail runs table-driven tests for the SendEmail function
func TestSendEmail(t *testing.T) {
	tests := []sendEmailTest{
		{
			smtpUser:     "smtpuser@example.com",
			smtpPassword: "password",
			smtpHost:     "smtp.example.com",
			smtpPort:     "587",
			to:           "recipient@example.com",
			subject:      "Greetings",
			body:         "Hello, How are you?",
			wantErr:      true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			err := weblogin.SendEmail(tt.smtpUser, tt.smtpPassword, tt.smtpHost, tt.smtpPort, tt.to, tt.subject, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
