package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

var clientsMap = make(map[*client]bool)
var broadcastChannel = make(chan string)

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

	logrus.Info("pwd: %s", wd)
	logPath := filepath.Join(wd, "server.log")

	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.Info("logPath: %s", logPath)
		logrus.SetOutput(file)
	} else {
		logrus.Info("Failed to log to file, using default stderr")
	}

	// DB setting
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", HOST, USER, PASSWORD, DATABASE),
	)
	if err != nil {
		panic(err)
	}

	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Successfully created connection to database")

	// handler setting
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Error(err)
			return
		}

		// generate uuid by pointer
		clientID := fmt.Sprintf("%p", conn)

		logrus.Info("New client connected: %s\n", clientID)

		c := &client{conn: conn, clientID: clientID}
		clientsMap[c] = true

		// a separate goroutine to listen on client
		go listenToClient(c)
	})

	// one goroutine to handle broadcast
	go broadcast()

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
		broadcastChannel <- fmt.Sprintf("%s: %s", c.clientID, message)

		logrus.Info("listenToClient() end...")
	}
}

func broadcast() {
	logrus.Info("broadcast() start...")
	for {
		message := <-broadcastChannel

		for c := range clientsMap {
			logrus.Info("start writing to client %s with message ---> %s\n", c.clientID, message)

			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				logrus.Error("error in broadcast: %v", err)
				c.conn.Close()
				delete(clientsMap, c)
			}
		}

		logrus.Info("broadcast() end...")
	}
}
