package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var requestBody LoginRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			// 处理解析错误
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// 获取字段的值
		username := requestBody.Username
		password := requestBody.Password

		logrus.Info("username: %s, password: %s", username, password)

		// check username & password satisfy pair in DB
		if username == "admin" && password == "password" {
			session, _ := store.Get(r, sessionName)
			session.Values[sessionUserID] = 123 // 存储用户ID或其他信息
			session.Save(r, w)

			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)

	// assertion to check value in session.Values[sessionUserID] is int -> if ok means user has logined
	if userID, ok := session.Values[sessionUserID].(int); ok {
		dashboard := fmt.Sprintf("Welcome, User ID: %d", userID)
		logrus.Info(w, dashboard)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)

	delete(session.Values, sessionUserID)
	session.Save(r, w)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
