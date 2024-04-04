package main

import (
	"net/http"
	"path/filepath"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
	"github.com/bnixon67/webapp/webutil"
)

func AddRoutes(mux *http.ServeMux, app *webauth.AuthApp) {
	assetDir := assets.AssetPath()
	cssFile := filepath.Join(assetDir, "css", "w3.css")
	icoFile := filepath.Join(assetDir, "ico", "favicon.ico")

	mux.Handle("/",
		http.RedirectHandler("/user", http.StatusFound))
	mux.HandleFunc("/w3.css", webutil.ServeFileHandler(cssFile))
	mux.HandleFunc("/favicon.ico", webutil.ServeFileHandler(icoFile))
	mux.HandleFunc("GET /user", app.UserGetHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/userscsv", app.UsersCSVHandler)
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("/confirm_request", app.ConfirmRequestHandler)
	mux.HandleFunc("GET /confirm", app.ConfirmHandlerGet)
	mux.HandleFunc("POST /confirm", app.ConfirmHandlerPost)
	mux.HandleFunc("/reset", app.ResetHandler)
	mux.HandleFunc("/events", app.EventsHandler)
	mux.HandleFunc("/eventscsv", app.EventsCSVHandler)
	mux.HandleFunc("GET /login", app.LoginGetHandler)
	mux.HandleFunc("POST /login", app.LoginPostHandler)

	// https://www.w3.org/TR/change-password-url/
	mux.Handle("/.well-known/change-password",
		http.RedirectHandler("/forgot", http.StatusFound))
}

func AddMiddleware(h http.Handler) http.Handler {
	h = webhandler.AddSecurityHeaders(h)
	h = webhandler.LogRequest(h)
	h = webhandler.AddLogger(h)
	h = webhandler.AddRequestID(h)

	return h
}
