package main

import (
	"net/http"
	"path/filepath"

	"github.com/bnixon67/webapp/assets"
	"github.com/bnixon67/webapp/webauth"
	"github.com/bnixon67/webapp/webhandler"
)

func AddRoutes(mux *http.ServeMux, app *webauth.AuthApp) {
	assetDir := assets.AssetPath()
	cssFile := filepath.Join(assetDir, "css", "w3.css")
	icoFile := filepath.Join(assetDir, "ico", "favicon.ico")

	mux.Handle("/",
		http.RedirectHandler("/user", http.StatusFound))
	mux.HandleFunc("/events", app.EventsHandler)
	mux.HandleFunc("/eventscsv", app.EventsCSVHandler)
	mux.HandleFunc("/favicon.ico", webhandler.ServeFileHandler(icoFile))
	mux.HandleFunc("/forgot", app.ForgotHandler)
	mux.HandleFunc("GET /confirm", app.ConfirmHandlerGet)
	mux.HandleFunc("GET /confirmed", app.ConfirmedHandlerGet)
	mux.HandleFunc("GET /confirm_request", app.ConfirmRequestHandlerGet)
	mux.HandleFunc("GET /confirm_request_sent", app.ConfirmRequestSentHandlerGet)
	mux.HandleFunc("GET /login", app.LoginGetHandler)
	mux.HandleFunc("GET /user", app.UserGetHandler)
	mux.HandleFunc("/logout", app.LogoutHandler)
	mux.HandleFunc("POST /confirm", app.ConfirmHandlerPost)
	mux.HandleFunc("POST /confirm_request", app.ConfirmRequestHandlerPost)
	mux.HandleFunc("POST /login", app.LoginPostHandler)
	mux.HandleFunc("/register", app.RegisterHandler)
	mux.HandleFunc("/reset", app.ResetHandler)
	mux.HandleFunc("/users", app.UsersHandler)
	mux.HandleFunc("/userscsv", app.UsersCSVHandler)
	mux.HandleFunc("/w3.css", webhandler.ServeFileHandler(cssFile))

	// https://www.w3.org/TR/change-password-url/
	mux.Handle("/.well-known/change-password",
		http.RedirectHandler("/forgot", http.StatusFound))
}

func AddMiddleware(h http.Handler) http.Handler {
	h = webhandler.AddSecurityHeaders(h)
	h = webhandler.LogRequest(h)
	h = webhandler.MiddlewareLogger(h)
	h = webhandler.AddRequestID(h)

	return h
}
