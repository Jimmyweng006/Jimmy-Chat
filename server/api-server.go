package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	store *sessions.CookieStore
)

const (
	sessionName   = "my-session"
	sessionUserID = "user-id"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func init() {
	store = sessions.NewCookieStore([]byte("secret-key"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1天
		HttpOnly: true,
	}
}

func logOutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)

	delete(session.Values, sessionUserID)
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
