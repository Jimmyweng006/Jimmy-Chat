package usecase

import (
	"context"
	"time"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	queue "github.com/Jimmyweng006/Jimmy-Chat/server/messageQueue"
	"github.com/sirupsen/logrus"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type messageUsecase struct {
	messageRepository domain.MessageRepository
	messageEventQueue queue.MessageQueue
}

func NewMessageUsecase(repository domain.MessageRepository, messageEventQueue queue.MessageQueue) domain.MessageUsecase {
	return &messageUsecase{
		messageRepository: repository,
		messageEventQueue: messageEventQueue,
	}
}

// message repository
func (m *messageUsecase) Store(ctx context.Context, message *db.Message) error {
	err := m.messageRepository.Store(ctx, message)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}

func (m *messageUsecase) RetriveHistoryMessage(ctx context.Context, roomID int64, u domain.UserUsecase) (*[]domain.MessageDTO, error) {
	// read room history messages and send to current user directly
	start := time.Now()
	historyMessages, err := m.messageRepository.GetByRoomID(context.Background(), roomID)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	end := time.Now()
	elapsed := end.Sub(start)
	logrus.Infof("it takes %s to retrive %d history data from DB\n", elapsed, len(*historyMessages))

	// convert to DTO(transport layer to encode response to frontend)
	var data []domain.MessageDTO
	for _, message := range *historyMessages {
		user, err := u.GetByUserID(context.Background(), int64(message.SenderID))
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		data = append(data, domain.MessageDTO{
			Sender:  user.Username,
			Content: message.MessageText,
		})
	}

	return &data, nil
}

// messageQueue
func (m *messageUsecase) ReadMessageFromMessageQueue(ctx context.Context) ([]byte, error) {
	message, err := m.messageEventQueue.ReadMessage(ctx)
	if err != nil {
		logrus.Error(err)
	}

	return message, nil
}

func (m *messageUsecase) WriteMessagesToMessageQueue(ctx context.Context, messages [][]byte) error {
	return m.messageEventQueue.WriteMessages(ctx, messages)
}

func (m *messageUsecase) CloseMessageQueueReader() error {
	return m.messageEventQueue.CloseReader()
}

func (m *messageUsecase) CloseMessageQueueWriter() error {
	return m.messageEventQueue.CloseWriter()
}
