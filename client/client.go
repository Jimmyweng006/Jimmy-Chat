package main

import (
	"bufio"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/chat"}
	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Error("dial erro:", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	go handleUserInput(connection, scanner)

	go listenToServer(connection)

	// Block the main goroutine to keep the program running
	select {}
}

func handleUserInput(connection *websocket.Conn, scanner *bufio.Scanner) {
	logrus.Info("handleUserInput() start...")

	for {
		// fmt.Print("\nEnter message: ")
		scanner.Scan()
		message := scanner.Text()

		err := connection.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			logrus.Error("write:", err)
			return
		}

		logrus.Info("handleUserInput() end...")
	}
}

func listenToServer(connection *websocket.Conn) {
	logrus.Info("listenToServer() start...")
	defer func() {
		connection.Close()
	}()

	for {
		logrus.Info("start reading from server with message")
		_, echoMessage, err := connection.ReadMessage()
		if err != nil {
			logrus.Error("error in reading from server: %v", err)
			break
		}
		logrus.Info(string(echoMessage))

		logrus.Info("listenToServer() end...")
	}
}
