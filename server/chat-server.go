package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	delivery "github.com/Jimmyweng006/Jimmy-Chat/server/user/delivery"
	repository "github.com/Jimmyweng006/Jimmy-Chat/server/user/repository"
	usecase "github.com/Jimmyweng006/Jimmy-Chat/server/user/usecase"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

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

const (
	// HOST value should equal to service name in yaml.services
	HOST     = "db"
	DATABASE = "postgres"
	USER     = "postgres"
	PASSWORD = "root"
)

func main() {
	// log setting
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	logrus.Info("pwd: ", wd)
	logPath := filepath.Join(wd, "server.log")

	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.Info("logPath: ", logPath)
		logrus.SetOutput(file)
	} else {
		logrus.Info("Failed to log to file, using default stderr")
	}

	// DB setting
	dbConnection, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", HOST, USER, PASSWORD, DATABASE),
	)
	if err != nil {
		logrus.Error(err)
	}

	if err = dbConnection.Ping(); err != nil {
		logrus.Error(err)
	}

	defer dbConnection.Close()

	logrus.Info("Successfully created connection to database")

	// Clean Architecture injection
	queryObject := db.New(dbConnection)
	userRepository := repository.NewUserRepository(queryObject)
	userUsercase := usecase.NewUserUsecase(userRepository)
	userHandler := delivery.NewHandler(userUsercase)

	// handler setting
	http.HandleFunc("/signin", userHandler.SignInHandler)
	// http.HandleFunc("/login", logInHandler)
	// http.HandleFunc("/logout", logOutHandler)
	// http.HandleFunc("/dashboard", dashboardHandler)

	// http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
	// 	conn, err := upgrader.Upgrade(w, r, nil)
	// 	if err != nil {
	// 		logrus.Error(err)
	// 		return
	// 	}

	// 	// generate uuid by pointer
	// 	clientID := fmt.Sprintf("%p", conn)

	// 	createdUser, err := queryObject.CreateUser(context.Background(), clientID)
	// 	if err != nil {
	// 		logrus.Error(err)
	// 	}

	// 	logrus.Info("createdUser info %v", createdUser)
	// 	logrus.Info("New client connected: %s\n", clientID)

	// 	c := &client{conn: conn, clientID: clientID}
	// 	clientsMap[c] = true

	// 	// a separate goroutine to listen on client
	// 	go listenToClient(c)
	// })

	// // one goroutine to handle broadcast
	// go broadcast(queryObject)

	logrus.Info("server start on port: 8080")
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}

func listenToClient(c *client) {
	logrus.Info("listenToClient() start...")
	defer func() {
		logrus.Info("Client disconnected: %s\n", c.clientID)
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

		logrus.Info("start reading from client %s with message ---> %s \n", c.clientID, string(message))
		broadcastChannel <- Message{
			Sender:  c.clientID,
			Content: message,
		}

		logrus.Info("listenToClient() end...")
	}
}

func broadcast(queryObject *db.Queries) {
	logrus.Info("broadcast() start...")
	for {
		messageObject := <-broadcastChannel
		logrus.Info("messageObject info %v", messageObject)

		senderObject, err := queryObject.FindUser(context.Background(), messageObject.Sender)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Info("senderObject info %v", senderObject)

		createMessageParams := db.CreateMessageParams{
			RoomID:         123,
			ReplyMessageID: sql.NullInt64{Valid: false},
			SenderID:       senderObject.ID,
			MessageText:    string(messageObject.Content),
		}

		sendMessage, err := queryObject.CreateMessage(context.Background(), createMessageParams)
		if err != nil {
			logrus.Error(err)
		}
		logrus.Info("sendMessage info %v", sendMessage)

		for c := range clientsMap {
			logrus.Info("start writing to client %s with message ---> %s\n", c.clientID, string(messageObject.Content))

			messageSend := fmt.Sprintf(senderObject.Username + ": " + string(messageObject.Content))
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
