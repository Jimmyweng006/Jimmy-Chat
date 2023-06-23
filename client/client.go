package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	waitForChat := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		for {
			// menu functionality
			fmt.Println("possible command as below, please type exactly")
			fmt.Println("signIn: will prompt input username and password")
			fmt.Println("logIn")
			fmt.Println("logOut")
			fmt.Println("chat: please logIn before start chating")
			fmt.Println("exit")
			fmt.Print("please input available command: ")

			scanner.Scan()
			command := scanner.Text()

			switch command {
			case "signIn":
				signIn(scanner)
			case "logIn":
				logIn(scanner)
			case "chat":
				wg.Add(1) // 增加等待計數器

				go func() {
					chat(scanner)
					// wg.Done(), not execute to block main goroutine
				}()

				wg.Wait() // 等待 chat 函式完成
			case "exit":
				fmt.Println("client program exit")
				return
			default:
				fmt.Println("invalid command")
			}
		}
	}()

	<-waitForChat // 等待 chat 函式完成後才結束主程式
}

type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var token = ""

func signIn(scanner *bufio.Scanner) {
	// retrieve username and password(hashed)
	fmt.Println("please type in username")
	fmt.Print("username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Println()

	fmt.Println("please type in password")
	fmt.Print("password: ")
	scanner.Scan()
	password := scanner.Text()

	// start request
	data := UserData{
		Username: username,
		Password: password,
	}

	reqBody, err := json.Marshal(data)
	if err != nil {
		log.Fatal("construct user signIn body error:", err)
		return
	}

	client := &http.Client{}

	request, err := http.NewRequest("POST", "http://localhost:8080/signIn", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal("generate request error:", err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal("send request error:", err)
		return
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read response error:", err)
		return
	}

	fmt.Println("respBody:", string(respBody))
}

func logIn(scanner *bufio.Scanner) {
	// retrieve username and password(hashed)
	fmt.Println("please type in username")
	fmt.Print("username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Println()

	fmt.Println("please type in password")
	fmt.Print("password: ")
	scanner.Scan()
	password := scanner.Text()

	// start request
	data := UserData{
		Username: username,
		Password: password,
	}

	reqBody, err := json.Marshal(data)
	if err != nil {
		log.Fatal("construct user signIn body error:", err)
		return
	}

	client := &http.Client{}

	request, err := http.NewRequest("POST", "http://localhost:8080/logIn", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal("generate request error:", err)
		return
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatal("send request error:", err)
		return
	}

	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read response error:", err)
		return
	}

	fmt.Println("respBody:", string(respBody))
	token = string(respBody)
}

func chat(scanner *bufio.Scanner) {
	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost:8080",
		Path:     "/chat",
		RawQuery: fmt.Sprintf("token=%s", token),
	}

	logrus.Infof("url: %s", u.String())
	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.Error("dial error:", err)
		return
	}

	go handleUserInput(connection, scanner)

	go listenToServer(connection)
}

func handleUserInput(connection *websocket.Conn, scanner *bufio.Scanner) {
	logrus.Info("handleUserInput() start...")

	for {
		fmt.Print("\nEnter message: ")
		scanner.Scan()
		message := scanner.Text()

		err := connection.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			logrus.Error("write:", err)
			return
		}

		// logrus.Info("handleUserInput() end...")
	}
}

func listenToServer(connection *websocket.Conn) {
	logrus.Info("listenToServer() start...")
	defer func() {
		connection.Close()
	}()

	for {
		// logrus.Info("start reading from server with message")
		_, echoMessage, err := connection.ReadMessage()
		if err != nil {
			logrus.Error("error in reading from server: ", err)
			break
		}
		// logrus.Info(string(echoMessage))
		fmt.Printf("\n%s\n", echoMessage)

		// logrus.Info("listenToServer() end...")
	}
}
