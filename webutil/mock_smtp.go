// Copyright 2023 Bill Nixon. All rights reserved.
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

// MockSMTPServerStart starts a mock SMTP server using the provided addr.
// The server will signal on ready when setup is complete.
func MockSMTPServerStart(ready chan<- bool, addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("failed to start mock SMTP server",
			slog.String("addr", addr),
			slog.Any("err", err))
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("mock SMTP Server running", slog.String("addr", addr))

	ready <- true

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept", slog.Any("err", err))
			continue
		}
		go MockSMTPServerConnection(conn)
	}
}

// MockSMTPServerConnection handles a single mock SMTP server connection.
func MockSMTPServerConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Fprintf(conn, "220 mock.smtp.server\r\n")

	scanner := bufio.NewScanner(conn)
	dataMode := false // Flag to track if we are in data mode

	for scanner.Scan() {
		line := scanner.Text()
		slog.Debug("received", slog.String("line", line))

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
