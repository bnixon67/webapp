// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webutil

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
)

// MockSMTPServerStart starts a mock SMTP server using the provided address.
// The server will signal on the ready channel when setup is complete.
func MockSMTPServerStart(ready chan<- bool, address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to start mock SMTP server",
			slog.String("address", address),
			slog.Any("error", err))
		os.Exit(1) // TODO: Handle error more gracefully

	}
	defer listener.Close()

	slog.Info("mock SMTP Server running", slog.String("address", address))

	ready <- true // Notify that the server is ready

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection",
				slog.Any("error", err))
			continue
		}
		go handleSMTPConnection(conn)
	}
}

// handleSMTPConnection handles a single SMTP server connection.
func handleSMTPConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "220 mock.smtp.server\r\n")

	scanner := bufio.NewScanner(conn)
	inDataMode := false // Track if we are in data mode

	for scanner.Scan() {
		line := scanner.Text()
		slog.Debug("received", slog.String("line", line))

		if inDataMode {
			if line == "." { // end of data marker
				fmt.Fprintf(conn, "250 OK: Message accepted for delivery\r\n")
				inDataMode = false // Reset data mode
			}
			continue
		}

		// Handle SMTP commands outside data mode
		switch {
		case strings.HasPrefix(line, "HELO"), strings.HasPrefix(line, "EHLO"):
			fmt.Fprintf(conn, "250-Hello\r\n250-AUTH PLAIN\r\n250 OK\r\n")
		case strings.HasPrefix(line, "AUTH"):
			fmt.Fprintf(conn, "235 OK\r\n")
		case strings.HasPrefix(line, "MAIL FROM:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "RCPT TO:"):
			fmt.Fprintf(conn, "250 OK\r\n")
		case strings.HasPrefix(line, "DATA"):
			fmt.Fprintf(conn, "354 Start mail input; end with <CRLF>.<CRLF>\r\n")
			inDataMode = true // Enter data mode
		case strings.HasPrefix(line, "QUIT"):
			fmt.Fprintf(conn, "221 Bye\r\n")
			return
		default:
			fmt.Fprintf(conn, "502 Command not implemented\r\n")
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Error("error reading from connection",
			slog.Any("error", err))
	}
}
