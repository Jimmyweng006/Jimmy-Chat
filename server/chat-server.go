package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	mq "github.com/Jimmyweng006/Jimmy-Chat/server/messageQueue"
	"github.com/Jimmyweng006/Jimmy-Chat/server/user/delivery"
	userRepository "github.com/Jimmyweng006/Jimmy-Chat/server/user/repository"
	UserUsecase "github.com/Jimmyweng006/Jimmy-Chat/server/user/usecase"
	"github.com/segmentio/kafka-go"

	messageRepository "github.com/Jimmyweng006/Jimmy-Chat/server/message/repository"
	messageUsecase "github.com/Jimmyweng006/Jimmy-Chat/server/message/usecase"

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
	HOST       = "db"
	DATABASE   = "postgres"
	USER       = "postgres"
	PASSWORD   = "root"
	kafkaTopic = "public-room"
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

	// Kagka setting
	// 設定 Kafka 連線相關設定
	brokers := []string{"broker:29091"}
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	// 建立 Kafka Reader
	readerConfig := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   kafkaTopic,
		Dialer:  dialer,
	}
	// 建立 Kafka Writer
	writerConfig := kafka.WriterConfig{
		Brokers: brokers,
		Topic:   kafkaTopic,
		Dialer:  dialer,
	}

	// Clean Architecture injection
	queryObject := db.New(dbConnection)
	kafkaMessageQueue := mq.NewKafka(readerConfig, writerConfig)
	messageQueueWrapper := mq.NewMessageQueueWrapper(kafkaMessageQueue)

	userRepository := userRepository.NewUserRepository(queryObject)
	userUsercase := UserUsecase.NewUserUsecase(userRepository)

	messageRepository := messageRepository.NewMessageRepository(queryObject)
	messageUsecase := messageUsecase.NewMessageUsecase(messageRepository, messageQueueWrapper)

	userHandler := delivery.NewHandler(userUsercase, messageUsecase)

	// handler setting
	http.HandleFunc("/signIn", userHandler.SignInHandler)
	http.HandleFunc("/logIn", userHandler.LogInHandler)
	// http.HandleFunc("/logout", logOutHandler)
	http.HandleFunc("/chat", userHandler.ChatHandler)

	logrus.Info("server start on port: 8080")
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
