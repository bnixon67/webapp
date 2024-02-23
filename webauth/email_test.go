package webauth_test

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bnixon67/webapp/webauth"
)

// Define a test struct to hold the test case data
type sendEmailTest struct {
	name        string
	smtpConfig  webauth.ConfigSMTP
	mailMessage webauth.MailMessage
	wantErr     bool
}

const (
	mockHost = "localhost"
	mockPort = "2525"
)

func TestMain(m *testing.M) {
	var wg sync.WaitGroup
	wg.Add(1)

	go startMockSMTPServer(&wg, mockHost+":"+mockPort)
	wg.Wait()

	os.Exit(m.Run())
}

func startMockSMTPServer(wg *sync.WaitGroup, hostPort string) {
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	slog.Debug("mock SMTP Server running", "host:port", hostPort)
	wg.Done()

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept", "err", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "220 mock.smtp.server\r\n")

	scanner := bufio.NewScanner(conn)
	dataMode := false // Flag to track if we are in data mode

	for scanner.Scan() {
		line := scanner.Text()
		slog.Debug("received", "line", line)

		if dataMode {
			// Check for end of data marker
			if line == "." {
				fmt.Fprintf(conn, "250 OK: Message accepted for delivery\r\n")
				dataMode = false // Reset data mode
			}
			continue // Keep reading data until end of data marker
		}

		// Handle SMTP commands
		switch {
		case strings.HasPrefix(line, "HELO") || strings.HasPrefix(line, "EHLO"):
			fmt.Fprintf(conn, "250-Hello\r\n250-AUTH PLAIN\r\n250 OK\r\n")
		case strings.HasPrefix(line, "AUTH"):
			fmt.Fprintf(conn, "235 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "DATA"):
			fmt.Fprintf(conn, "354 Start mail input; end with <CRLF>.<CRLF>\r\n")
			dataMode = true // Enter data mode
		case strings.HasPrefix(line, "QUIT"):
			fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprintf(conn, "502 Command not implemented\r\n")
		}
	}
}

// TestSendEmail runs table-driven tests for the SendEmail function
func TestSendEmail(t *testing.T) {
	tests := []sendEmailTest{
		{
			name: "invalid smtp server",
			smtpConfig: webauth.ConfigSMTP{
				Host:     "smtp.example.com",
				Port:     "587",
				User:     "smtpuser@example.com",
				Password: "password",
			},
			mailMessage: webauth.MailMessage{
				To:      "recipient@example.com",
				Subject: "Greetings",
				Body:    "Hello, How are you?",
			},
			wantErr: true,
		},
		{
			name: "valid smtp server",
			smtpConfig: webauth.ConfigSMTP{
				Host:     mockHost,
				Port:     mockPort,
				User:     "smtpuser@example.com",
				Password: "password",
			},
			mailMessage: webauth.MailMessage{
				To:      "recipient@example.com",
				Subject: "Greetings",
				Body:    "Hello, How are you?",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := webauth.SendEmail(tc.smtpConfig, tc.mailMessage)
			if (err != nil) != tc.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
