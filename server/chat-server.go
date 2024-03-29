package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	mq "github.com/Jimmyweng006/Jimmy-Chat/server/messageQueue"
	"github.com/Jimmyweng006/Jimmy-Chat/server/user/delivery"
	userRepository "github.com/Jimmyweng006/Jimmy-Chat/server/user/repository"
	UserUsecase "github.com/Jimmyweng006/Jimmy-Chat/server/user/usecase"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/viper"

	messageRepository "github.com/Jimmyweng006/Jimmy-Chat/server/message/repository"
	messageUsecase "github.com/Jimmyweng006/Jimmy-Chat/server/message/usecase"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Message struct {
	Sender  string
	Content []byte
}

const (
	// HOST value should equal to service name in yaml.services
	HOST           = "db"
	DATABASE       = "postgres"
	USER           = "postgres"
	PASSWORD       = "root"
	KAFKA_TOPIC    = "public-room"
	KAFKA_GROUP_ID = "public-room-group"
)

func main() {
	// Log setting
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}

	logPath := filepath.Join(wd, "server.log")

	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.Info("logPath: ", logPath)
		logrus.SetOutput(file)
	} else {
		logrus.Info("Failed to log to file, using default stderr")
	}

	// get config
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	viper.SetConfigName("config." + env) // 例如 config.development 或 config.production
	viper.AddConfigPath(".")             // 可以添加多個搜索路徑
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("Error reading config file, %s", err)
	}

	allowedOrigins := viper.GetString("allowedOrigins")

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
	brokers := []string{"broker:29091"}
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	// Kafka Reader
	readerConfig := kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   KAFKA_TOPIC,
		Dialer:  dialer,
		GroupID: KAFKA_GROUP_ID,
	}
	// Kafka Writer
	writerConfig := kafka.WriterConfig{
		Brokers: brokers,
		Topic:   KAFKA_TOPIC,
		Dialer:  dialer,
	}
	// chat service related config
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")

			for _, allowedOrigin := range strings.Split(allowedOrigins, ",") {
				if origin == allowedOrigin {
					return true
				}
			}

			return false
		},
	}

	// Clean Architecture injection
	queryObject := db.New(dbConnection)
	kafkaMessageQueue := mq.NewKafka(readerConfig, writerConfig)
	messageQueueWrapper := mq.NewMessageQueueWrapper(kafkaMessageQueue)

	userRepository := userRepository.NewUserRepository(queryObject)
	userUsercase := UserUsecase.NewUserUsecase(userRepository)

	messageRepository := messageRepository.NewMessageRepository(queryObject)
	messageUsecase := messageUsecase.NewMessageUsecase(messageRepository, messageQueueWrapper)

	userHandler := delivery.NewHandler(userUsercase, messageUsecase, upgrader)
	defer userHandler.MessageUsecase.CloseMessageQueueWriter()

	// one goroutine to handle broadcast
	go delivery.Broadcast(userHandler)

	delivery.SetRouter(userHandler, allowedOrigins)
}
