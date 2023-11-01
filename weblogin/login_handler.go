package weblogin

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

// LoginPageData contains data passed to the HTML template.
type LoginPageData struct {
	Title   string
	Message string
}

// LoginHandler handles /login requests.
func (app *LoginApp) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Get logger with request info from request context and add calling function name.
	logger := webhandler.LoggerFromContext(r.Context()).With(slog.String("func", webhandler.FuncName()))

	// Check if the HTTP method is valid.
	if !webutil.ValidMethod(w, r, http.MethodGet, http.MethodPost) {
		logger.Error("invalid method")
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := webutil.RenderTemplate(app.Tmpl, w, "login.html",
			LoginPageData{Title: app.Cfg.Title})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		logger.Info("LoginHandler")

	case http.MethodPost:
		app.loginPost(w, r)
	}
}

const (
	MsgMissingUserNameAndPassword = "Missing username and password"
	MsgMissingUserName            = "Missing username"
	MsgMissingPassword            = "Missing password"
	MsgLoginFailed                = "Login Failed"
)

// loginPost is called for the POST method of the LoginHandler.
func (app *LoginApp) loginPost(w http.ResponseWriter, r *http.Request) {
	// get form values
	userName := strings.TrimSpace(r.PostFormValue("username"))
	password := strings.TrimSpace(r.PostFormValue("password"))

	logger := slog.With(slog.Group("form",
		"userName", userName,
		"password empty", password == ""))

	// check for missing values
	var msg string
	switch {
	case userName == "" && password == "":
		msg = MsgMissingUserNameAndPassword
	case userName == "":
		msg = MsgMissingUserName
	case password == "":
		msg = MsgMissingPassword
	}
	if msg != "" {
		logger.Info("error", "display", msg)
		err := webutil.RenderTemplate(app.Tmpl, w, "login.html",
			LoginPageData{Title: app.Cfg.Title, Message: msg})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	// attempt to login the given userName with the given password
	token, err := app.LoginUser(userName, password)
	if err != nil {
		logger.Error("failed to LoginUser", "err", err)
		err := webutil.RenderTemplate(app.Tmpl, w, "login.html",
			LoginPageData{
				Title:   app.Cfg.Title,
				Message: MsgLoginFailed,
			})
		if err != nil {
			logger.Error("unable to RenderTemplate", "err", err)
			return
		}
		return
	}

	// login successful, so create a cookie for the session Token
	http.SetCookie(w, &http.Cookie{
		Name:     SessionTokenCookieName,
		Value:    token.Value,
		Expires:  token.Expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	redirect := r.URL.Query().Get("r")
	if redirect == "" {
		redirect = "/"
	}

	// redirect from login page
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	logger.Info("login successful")
}
