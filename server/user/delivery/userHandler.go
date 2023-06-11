package delivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
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

// JWT config
var secretKey = []byte("your-secret-key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
			// 生成 JWT
			token := jwt.New(jwt.SigningMethodHS256)
			claims := token.Claims.(jwt.MapClaims)
			claims["user_id"] = userData.ID
			claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // 設定過期時間為 1 天

			// 簽署 JWT
			signedToken, err := token.SignedString(secretKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// 將 JWT 發送給客戶端
			w.Write([]byte(signedToken))

			logrus.Infof("user:%s login successfully", userData.Username)
			return
		} else {
			http.Error(w, "password not match", http.StatusBadRequest)
			return
		}

	}
}

func (h *UserHandler) ChatHandler(w http.ResponseWriter, r *http.Request) {
	tokenFromURL := extractTokenFromURL(r.URL)

	// 驗證 JWT
	token, err := jwt.Parse(tokenFromURL, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	// 提取使用者 ID
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	clientID := claims["user_id"].(float64)

	logrus.Infof("New client: %f connected\n", clientID)
	c := &client{
		conn:     conn,
		clientID: strconv.Itoa(int(clientID)),
	}

	clientsMap[c] = true

	// a separate goroutine to listen on client
	go listenToClient(c)

	// one goroutine to handle broadcast
	go broadcast(h)

}

func extractTokenFromURL(url *url.URL) string {
	token := ""
	query := url.Query()
	if tokens, ok := query["token"]; ok && len(tokens) > 0 {
		token = tokens[0]
	}

	return token
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
			logrus.Error("error in listening: ", err)
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

		senderID, err := strconv.Atoi(messageObject.Sender)
		if err != nil {
			logrus.Error(err)
			return
		}

		err = h.MessageUsecase.Store(context.Background(), &db.Message{
			RoomID:         1, // 1 mean public room
			ReplyMessageID: sql.NullInt64{Valid: false},
			SenderID:       int64(senderID),
			MessageText:    string(messageObject.Content),
		})
		if err != nil {
			logrus.Error(err)
			return
		}

		for c := range clientsMap {
			logrus.Infof("start writing to client %s with message ---> %s\n", c.clientID, string(messageObject.Content))

			messageSend := fmt.Sprintf("User %d: %s", senderID, messageObject.Content)
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(messageSend))
			if err != nil {
				logrus.Error("error in broadcast: ", err)
				c.conn.Close()
				delete(clientsMap, c)
				return
			}
		}

		logrus.Info("broadcast() end...")
	}
}
