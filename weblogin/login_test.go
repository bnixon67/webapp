package weblogin_test

import (
	"errors"
	"testing"

	"github.com/bnixon67/webapp/weblogin"
)

func TestLoginUser(t *testing.T) {
	testCases := []struct {
		name      string
		username  string
		password  string
		wantToken weblogin.Token
		wantErr   error
	}{
		{
			name:     "Successful login",
			username: "test",
			password: "password",
			wantErr:  nil,
		},
		{
			name:     "Incorrect password",
			username: "test",
			password: "invalid",
			wantErr:  weblogin.ErrIncorrectPassword,
		},
		// Add more test cases for different scenarios
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := AppForTest(t)

			// gotToken, err := app.LoginUser(tc.username, tc.password)
			_, err := app.LoginUser(tc.username, tc.password)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("LoginUser() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// TODO: test token
			// if !reflect.DeepEqual(gotToken, tc.wantToken) {
			// 	t.Errorf("LoginUser() gotToken = %v, want %v", gotToken, tc.wantToken)
			// }
		})
	}
}
