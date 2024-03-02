// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"net"
	"os"
	"testing"

	"github.com/bnixon67/webapp/webutil"
)

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
