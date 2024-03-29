// Copyright 2024 Bill Nixon. All rights reserved.
// Use of this source code is governed by the license found in the LICENSE file.

package webauth_test

import (
	"errors"
	"testing"

	"github.com/bnixon67/webapp/webauth"
)

func TestWriteEvent(t *testing.T) {
	app := AppForTest(t)

	testCases := []struct {
		name    string
		db      *webauth.AuthDB
		event   webauth.Event
		wantErr error
	}{
		{
			name: "Success",
			db:   app.DB,
			event: webauth.Event{
				Name:      webauth.EventLogin,
				Succeeded: true,
				Username:  "writevent",
			},
			wantErr: nil,
		},
		{
			name: "Username too long",
			db:   app.DB,
			event: webauth.Event{
				Name:      webauth.EventLogin,
				Succeeded: true,
				Username:  "1234567890123456789012345678901",
			},
			wantErr: webauth.ErrWriteEventFailed,
		},
		{
			name:    "InvalidDB",
			db:      nil,
			event:   webauth.Event{},
			wantErr: webauth.ErrWriteEventDBNil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.db.WriteEvent(tc.event.Name, tc.event.Succeeded, tc.event.Username, tc.event.Message)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}
