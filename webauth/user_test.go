package webauth_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/bnixon67/webapp/webauth"
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
		t.Fatalf("could not login user to get login token")
	}

	user, err := app.DB.UserForLoginToken(token.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}

	tests := []struct {
		name       string
		loginToken string
		wantUser   webauth.User
		wantError  bool
	}{
		{
			name:       "Empty login token",
			loginToken: "",
			wantUser:   webauth.User{},
			wantError:  false,
		},
		{
			name:       "Valid login token",
			loginToken: token.Value,
			wantUser:   user,
			wantError:  false,
		},
		{
			name:       "Invalid login token",
			loginToken: "invalid",
			wantUser:   webauth.User{},
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.loginToken != "" {
				req.AddCookie(&http.Cookie{Name: webauth.LoginTokenCookieName, Value: tt.loginToken})
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

func TestUserForLoginToken(t *testing.T) {
	app := AppForTest(t)
	db := app.DB

	validToken, err := app.LoginUser("test", "password")
	if err != nil {
		t.Fatalf("could not login user to get login token")
	}

	validUser, err := app.DB.UserForLoginToken(validToken.Value)
	if err != nil {
		t.Fatalf("could not get user")
	}

	expiredToken, err := db.CreateToken(webauth.LoginTokenKind, "test", webauth.LoginTokenSize, "0s")

	if err != nil {
		t.Fatalf("could not get user")
	}

	tests := []struct {
		name       string
		loginToken string
		wantUser   webauth.User
		wantErr    error
	}{
		{
			name:       "valid token",
			loginToken: validToken.Value,
			wantUser:   validUser,
			wantErr:    nil,
		},
		{
			name:       "invalid token",
			loginToken: "invalid",
			wantUser:   webauth.EmptyUser,
			wantErr:    webauth.ErrUserLoginTokenNotFound,
		},
		{
			name:       "expired token",
			loginToken: expiredToken.Value,
			wantUser:   webauth.EmptyUser,
			wantErr:    webauth.ErrUserLoginTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			gotUser, gotErr := db.UserForLoginToken(tt.loginToken)

			if gotErr != tt.wantErr {
				t.Errorf("UserForLoginToken(%q) gotErr = %v, wantErr %v", tt.loginToken, gotErr, tt.wantErr)

			}

			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("UserFromRequest() gotUser = %v, want %v", gotUser, tt.wantUser)
			}

		})
	}
}
