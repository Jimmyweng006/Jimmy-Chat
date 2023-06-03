package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	model "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
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

type UserHandler struct {
	UserUsecase domain.UserUsecase
}

func NewHandler(userUsercase domain.UserUsecase) *UserHandler {
	return &UserHandler{
		UserUsecase: userUsercase,
	}
}

func (h *UserHandler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// retrieve username and password
		var requestBody LoginRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Failed to parse request body", http.StatusBadRequest)
			return
		}

		username := requestBody.Username
		password := requestBody.Password

		logrus.Info("username: %s, password: %s", username, password)

		// check username not exist in db
		user, err := h.UserUsecase.GetByUsername(context.Background(), username)
		if err != nil {
			logrus.Error(err)
			return
		}
		if user.Username != "" {
			http.Error(w, "username had been used", http.StatusForbidden)
			return
		}

		// then create user in db
		err = h.UserUsecase.Store(context.Background(), &model.User{
			Username: username,
			Password: password,
		})
		if err != nil {
			http.Error(w, "create", http.StatusUnauthorized)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "create account successfully")
		return
	}
}
