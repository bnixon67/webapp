package webauth_test

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func loginBody(data webauth.LoginPageData) string {
	// Get path to template file.
	assetDir := assets.AssetPath()
	tmplFile := filepath.Join(assetDir, "tmpl", "login.html")

	// Parse the HTML template from a file.
	tmpl := template.Must(template.ParseFiles(tmplFile))

	// Create a buffer to store the rendered HTML.
	var body bytes.Buffer

	// Execute the template with the data and write result to the buffer.
	tmpl.Execute(&body, data)

	return body.String()
}

func TestLoginGetHandler(t *testing.T) {
	app := AppForTest(t)

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/login",
			RequestMethod: http.MethodPost,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
		{
			Name:          "Valid GET Request",
			Target:        "/login",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusOK,
			WantBody: loginBody(webauth.LoginPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
			}),
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.LoginGetHandler, tests)
}

func TestLoginPostHandler(t *testing.T) {
	app := AppForTest(t)

	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	d, err := time.ParseDuration(app.Cfg.LoginExpires)
	if err != nil {
		t.Fatalf("cannot parse duration")
	}
	expires := time.Now().Add(d)

	loginDontRememberCookie := webauth.LoginCookie("value", time.Time{})
	loginRememberCookie := webauth.LoginCookie("value", expires)

	tests := []webhandler.TestCase{
		{
			Name:          "Invalid Method",
			Target:        "/login",
			RequestMethod: http.MethodGet,
			WantStatus:    http.StatusMethodNotAllowed,
			WantBody:      "Error: Method Not Allowed\n",
		},
		{
			Name:           "Missing username and password",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: loginBody(webauth.LoginPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgMissingUsernameAndPassword,
			}),
		},
		{
			Name:           "Missing username",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"password": {"foo"}}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: loginBody(webauth.LoginPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgMissingUsername,
			}),
		},
		{
			Name:           "Missing password",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody:    url.Values{"username": {"foo"}}.Encode(),
			WantStatus:     http.StatusOK,
			WantBody: loginBody(webauth.LoginPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgMissingPassword,
			}),
		},
		{
			Name:           "Invalid Login",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{"username": {"foo"},
				"password": {"bar"}}.Encode(),
			WantStatus: http.StatusOK,
			WantBody: loginBody(webauth.LoginPageData{
				CommonPageData: webauth.CommonPageData{
					Title: app.Cfg.App.Name,
				},
				Message: webauth.MsgLoginFailed,
			}),
		},
		{
			Name:           "Valid Login - Don't Remember",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{"username": {"test"},
				"password": {"password"}}.Encode(),
			WantStatus:  http.StatusSeeOther,
			WantBody:    "",
			WantCookies: []http.Cookie{*loginDontRememberCookie},
			WantCookiesCmpOpts: []cmp.Option{
				cmpopts.IgnoreFields(http.Cookie{}, "Value"),
				cmpopts.IgnoreFields(http.Cookie{}, "Raw"),
			},
		},
		{
			Name:           "Valid Login - Remember",
			Target:         "/login",
			RequestMethod:  http.MethodPost,
			RequestHeaders: header,
			RequestBody: url.Values{"username": {"test"},
				"password": {"password"},
				"remember": {"on"}}.Encode(),
			WantStatus:  http.StatusSeeOther,
			WantBody:    "",
			WantCookies: []http.Cookie{*loginRememberCookie},
			WantCookiesCmpOpts: []cmp.Option{
				cmpopts.IgnoreFields(http.Cookie{}, "Value"),
				cmpopts.IgnoreFields(http.Cookie{}, "Raw"),
				cmpopts.IgnoreFields(http.Cookie{}, "RawExpires"),
				cmpopts.EquateApproxTime(5 * time.Second),
			},
		},
	}

	// Test the handler using the utility function.
	webhandler.HandlerTestWithCases(t, app.LoginPostHandler, tests)
}
