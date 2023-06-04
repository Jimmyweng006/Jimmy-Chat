package delivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// chat service related config
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type client struct {
	conn     *websocket.Conn
	clientID string
}

type Message struct {
	Sender  string
	Content []byte
}

var clientsMap = make(map[*client]bool)
var broadcastChannel = make(chan Message)

// session related config
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
	UserUsecase    domain.UserUsecase
	MessageUsecase domain.MessageUsecase
}

func NewHandler(userUsercase domain.UserUsecase, messageUsecase domain.MessageUsecase) *UserHandler {
	return &UserHandler{
		UserUsecase:    userUsercase,
		MessageUsecase: messageUsecase,
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

		logrus.Infof("username: %s, password: %s", username, password)

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
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Fatal(err)
		}

		fmt.Println("Hashed Password:", string(hashedPassword))

		err = h.UserUsecase.Store(context.Background(), &db.User{
			Username: username,
			Password: string(hashedPassword),
		})
		if err != nil {
			http.Error(w, "create", http.StatusUnauthorized)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "create account successfully")
		return
	}
}

func (h *UserHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
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

		logrus.Infof("username: %s, password: %s", username, password)

		// check username & password satisfy data in DB
		userData, err := h.UserUsecase.GetByUsername(context.Background(), username)
		if err != nil {
			http.Error(w, "check username error", http.StatusBadRequest)
			return
		} else if userData.Username != username {
			http.Error(w, "username not match", http.StatusBadRequest)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(password))
		if err == nil {
			session, _ := store.Get(r, sessionName)
			session.Values[sessionUserID] = userData.ID
			logrus.Infof("user session: %v", session)
			logrus.Infof("user session value: %v", session.Values)
			session.Save(r, w)

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "login successfully")
			return
		} else {
			http.Error(w, "password not match", http.StatusBadRequest)
			return
		}

	}
}

func (h *UserHandler) ChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	session, _ := store.Get(r, sessionName)

	// assertion to check value in session.Values[sessionUserID] is int -> if ok means user has logined
	if clientID, ok := session.Values[sessionUserID].(int); !ok {
		logrus.Infof("what is session %v", session.Values)
		logrus.Infof("what is session value %d", session.Values[sessionUserID])
		logrus.Error("user authentication fail")
		// http.Redirect(w, r, "/login", http.StatusSeeOther)
		http.Error(w, "start chating error", http.StatusBadRequest)
		return
	} else {
		// conn, err := upgrader.Upgrade(w, r, nil)
		// if err != nil {
		// 	logrus.Error(err)
		// 	return
		// }

		logrus.Infof("New client: %s connected\n", clientID)

		c := &client{conn: conn, clientID: strconv.Itoa(clientID)}
		clientsMap[c] = true

		// a separate goroutine to listen on client
		go listenToClient(c)

		// one goroutine to handle broadcast
		go broadcast(h)
	}

}

func listenToClient(c *client) {
	logrus.Info("listenToClient() start...")
	defer func() {
		logrus.Infof("Client disconnected: %s\n", c.clientID)
		c.conn.Close()
		delete(clientsMap, c)
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			logrus.Error("error in listening: %v", err)
			delete(clientsMap, c)
			break
		}

		logrus.Infof("start reading from client %s with message ---> %s \n", c.clientID, string(message))
		broadcastChannel <- Message{
			Sender:  c.clientID,
			Content: message,
		}

		logrus.Info("listenToClient() end...")
	}
}

func broadcast(h *UserHandler) {
	logrus.Info("broadcast() start...")
	for {
		messageObject := <-broadcastChannel
		logrus.Infof("messageObject info %v", messageObject)

		senderData, err := h.UserUsecase.GetByUsername(context.Background(), messageObject.Sender)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Infof("senderObject info %v", senderData)

		err = h.MessageUsecase.Store(context.Background(), &db.Message{
			RoomID:         123,
			ReplyMessageID: sql.NullInt64{Valid: false},
			SenderID:       senderData.ID,
			MessageText:    string(messageObject.Content),
		})
		if err != nil {
			logrus.Error(err)
		}

		for c := range clientsMap {
			logrus.Infof("start writing to client %s with message ---> %s\n", c.clientID, string(messageObject.Content))

			messageSend := fmt.Sprintf(senderData.Username + ": " + string(messageObject.Content))
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(messageSend))
			if err != nil {
				logrus.Error("error in broadcast: %v", err)
				c.conn.Close()
				delete(clientsMap, c)
			}
		}

		logrus.Info("broadcast() end...")
	}
}
