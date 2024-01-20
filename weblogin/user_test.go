package weblogin_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/bnixon67/webapp/weblogin"
)

func TestLastLoginForUser(t *testing.T) {
	var zeroTime time.Time

	dt := time.Date(2023, time.January, 15, 1, 0, 0, 0, time.UTC)

	cases := []struct {
		username string
		want     time.Time
		err      error
	}{
		{"no such user", zeroTime, nil},
		{"test1", zeroTime, nil},
		{"test2", dt, nil},
		{"test3", dt.Add(time.Hour), nil},
		{"test4", dt.Add(time.Hour * 2), nil},
	}

	app := AppForTest(t)

	for _, tc := range cases {
		got, _, err := app.DB.LastLoginForUser(tc.username)
		if !errors.Is(err, tc.err) {
			t.Errorf("LastLoginForUser(db, %q)\ngot err '%v' want '%v'", tc.username, err, tc.err)
		}
		if got != tc.want {
			t.Errorf("LastLoginForUser(db, %q)\n got '%v'\nwant '%v'", tc.username, got, tc.want)
		}
	}
}

func TestUserFromRequest(t *testing.T) {
	app := AppForTest(t)

	token, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get session token")
	}

	user, err := app.DB.UserForSessionToken(token.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}

	tests := []struct {
		name         string
		sessionToken string
		wantUser     weblogin.User
		wantError    bool
	}{
		{
			name:         "Empty session",
			sessionToken: "",
			wantUser:     weblogin.User{},
			wantError:    false,
		},
		{
			name:         "Valid session",
			sessionToken: token.Value,
			wantUser:     user,
			wantError:    false,
		},
		{
			name:         "Invalid session",
			sessionToken: "invalid",
			wantUser:     weblogin.User{},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{Name: weblogin.SessionTokenCookieName, Value: tt.sessionToken})
			}

			w := httptest.NewRecorder()

			gotUser, gotErr := app.DB.UserFromRequest(w, req)

			// Validate the returned user and error.
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("UserFromRequest() gotUser = %v, want %v", gotUser, tt.wantUser)
			}
			if (gotErr != nil) != tt.wantError {
				t.Errorf("UserFromRequest() gotErr = %v, wantErr %v", gotErr, tt.wantError)
			}

			// TODO: check for expired cookie
		})
	}
}
