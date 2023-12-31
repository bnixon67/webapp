package weblogin_test

import (
	"errors"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

func TestWriteEvent(t *testing.T) {
	app := AppForTest(t)

	// Test cases
	testCases := []struct {
		name    string
		db      *weblogin.LoginDB
		event   weblogin.Event
		wantErr error
	}{
		{
			name: "Success",
			db:   app.DB,
			event: weblogin.Event{
				Name:     weblogin.EventLogin,
				Success:  true,
				UserName: "writevent",
			},
			wantErr: nil,
		},
		{
			name: "UserName too long",
			db:   app.DB,
			event: weblogin.Event{
				Name:     weblogin.EventLogin,
				Success:  true,
				UserName: "1234567890123456789012345678901",
			},
			wantErr: weblogin.ErrWriteEventFailed,
		},
		{
			name:    "InvalidDB",
			db:      nil,
			event:   weblogin.Event{},
			wantErr: weblogin.ErrWriteEventInvalidDB,
		},
		// You can add more test cases here
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.db.WriteEvent(tc.event.Name, tc.event.Success, tc.event.UserName, tc.event.Message)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}
