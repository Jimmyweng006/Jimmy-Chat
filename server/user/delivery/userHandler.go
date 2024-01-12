package delivery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type client struct {
	conn     *websocket.Conn
	clientID string
}

type Message struct {
	Sender  string
	Content string
}

var clientsMap = make(map[*client]bool)
var broadcastChannel = make(chan Message)

// JWT config
var secretKey = []byte("your-secret-key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Kafka config
var kafkaTopic = "public-room"

var waitGroupReader sync.WaitGroup
var waitGroupWriter sync.WaitGroup

type UserHandler struct {
	UserUsecase    domain.UserUsecase
	MessageUsecase domain.MessageUsecase
	Upgrader       websocket.Upgrader
}

type LoginResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func NewHandler(userUsercase domain.UserUsecase, messageUsecase domain.MessageUsecase, upgrader websocket.Upgrader) *UserHandler {
	return &UserHandler{
		UserUsecase:    userUsercase,
		MessageUsecase: messageUsecase,
		Upgrader:       upgrader,
	}
}

func (h *UserHandler) SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// retrieve username and password
		var requestBody LoginRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, LoginResponse{Error: "Failed to parse request body"})
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
			respondJSON(w, http.StatusForbidden, map[string]string{"error": "username had been used"})
			return
		}

		// then create user in db
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Fatal(err)
		}

		err = h.UserUsecase.Store(context.Background(), &db.User{
			Username: username,
			Password: string(hashedPassword),
		})
		if err != nil {
			respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "create account fail"})
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{"message": "create account successfully"})
	}
}

func (h *UserHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// retrieve username and password
		var requestBody LoginRequest
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, LoginResponse{Error: "Failed to parse request body"})
			return
		}

		username := requestBody.Username
		password := requestBody.Password

		// check username & password satisfy data in DB
		userData, err := h.UserUsecase.GetByUsername(context.Background(), username)
		if err != nil {
			respondJSON(w, http.StatusBadRequest, LoginResponse{Error: "check username error"})
			return
		}

		if userData.Username != username {
			respondJSON(w, http.StatusBadRequest, LoginResponse{Error: fmt.Sprintf("can not find username: %s", username)})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(password)); err != nil {
			respondJSON(w, http.StatusBadRequest, LoginResponse{Error: "password not match"})
			return
		}

		// 生成 JWT
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = userData.ID
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // 設定過期時間為 1 天

		// 簽署 JWT
		signedToken, err := token.SignedString(secretKey)
		if err != nil {
			respondJSON(w, http.StatusInternalServerError, LoginResponse{Error: err.Error()})
			return
		}

		// 將 JWT 發送給客戶端
		respondJSON(w, http.StatusOK, LoginResponse{Token: signedToken})
		logrus.Infof("user:%s login successfully", userData.Username)
		return

	}
}

func (h *UserHandler) ChatHandler(w http.ResponseWriter, r *http.Request) {
	tokenFromURL := extractTokenFromURL(r.URL)
	// JWTtoken := r.Header["Authorization"][0]

	// 驗證 JWT
	token, err := jwt.Parse(tokenFromURL, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		respondJSON(w, http.StatusUnauthorized, LoginResponse{Error: "Not able to start chat, please re login."})
		return
	}

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		respondJSON(w, http.StatusUnauthorized, LoginResponse{Error: err.Error()})
		return
	}

	// 提取使用者 ID
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, LoginResponse{Error: err.Error()})
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
	go listenToClient(h, c)

	// // one goroutine to handle broadcast
	// go broadcast(h)
}

func respondJSON(w http.ResponseWriter, status int, responseMap interface{}) {
	response, err := json.Marshal(responseMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func extractTokenFromURL(url *url.URL) string {
	token := ""
	query := url.Query()
	if tokens, ok := query["token"]; ok && len(tokens) > 0 {
		token = tokens[0]
	}

	return token
}

func listenToClient(h *UserHandler, c *client) {
	logrus.Info("listenToClient() start...")
	defer func() {
		logrus.Infof("Client disconnected: %s\n", c.clientID)
		c.conn.Close()
		delete(clientsMap, c)
	}()

	// defer h.MessageUsecase.CloseMessageQueueWriter()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			logrus.Error("error in listening: ", err)
			delete(clientsMap, c)
			break
		}

		logrus.Infof("start reading from client %s with message ---> %s \n", c.clientID, string(message))

		messageObj := Message{
			Sender:  c.clientID,
			Content: string(message),
		}
		messageForKafka, err := json.Marshal(messageObj)
		if err != nil {
			logrus.Error("construct user signIn body error:", err)
			return
		}

		// 將訊息寫入 Kafka
		messagesForKafka := make([][]byte, 1)
		messagesForKafka[0] = messageForKafka

		logrus.Infof("Send message to Kafka: %s\n", messageForKafka)
		err = h.MessageUsecase.WriteMessagesToMessageQueue(context.Background(), messagesForKafka)
		if err != nil {
			logrus.Error("Error writing message: ", err)
			return
		}

		logrus.Info("listenToClient() end...")
	}
}

func Broadcast(h *UserHandler) {
	logrus.Info("broadcast() start...")

	for {
		m, err := h.MessageUsecase.ReadMessageFromMessageQueue(context.Background())
		if err != nil {
			break
		}
		// logrus.Printf("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		storeAndSendMessageToOtherUsersExceptSender(h, m)
	}
}

func storeAndSendMessageToOtherUsersExceptSender(h *UserHandler, message []byte) {
	var messageObject Message
	logrus.Infof("message from processMessage(): %s", string(message))
	if err := json.Unmarshal(message, &messageObject); err != nil {
		logrus.Error("Parse message from Kafka error: ", err)
		return
	}

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

	user, err := h.UserUsecase.GetByUserID(context.Background(), int64(senderID))
	if err != nil {
		logrus.Error(err)
		return
	}

	for c := range clientsMap {
		if c.clientID != strconv.Itoa(senderID) {
			MessagePushToFronted, err := json.Marshal(Message{
				Sender:  user.Username,
				Content: messageObject.Content,
			})
			if err != nil {
				logrus.Error("parse structure error when sending web socket message to frontend")
				return
			}

			err = c.conn.WriteMessage(websocket.TextMessage, MessagePushToFronted)
			if err != nil {
				logrus.Error("error in broadcast: ", err)
				c.conn.Close()
				delete(clientsMap, c)
				return
			}
		}
	}
}
