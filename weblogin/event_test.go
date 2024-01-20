package weblogin_test

import (
	"errors"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

func TestWriteEvent(t *testing.T) {
	app := AppForTest(t)

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
				Name:      weblogin.EventLogin,
				Succeeded: true,
				Username:  "writevent",
			},
			wantErr: nil,
		},
		{
			name: "Username too long",
			db:   app.DB,
			event: weblogin.Event{
				Name:      weblogin.EventLogin,
				Succeeded: true,
				Username:  "1234567890123456789012345678901",
			},
			wantErr: weblogin.ErrWriteEventFailed,
		},
		{
			name:    "InvalidDB",
			db:      nil,
			event:   weblogin.Event{},
			wantErr: weblogin.ErrWriteEventDBNil,
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
