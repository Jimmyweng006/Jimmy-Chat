package usecase

import (
	"context"

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

func (m *messageUsecase) Store(ctx context.Context, message *db.Message) error {
	err := m.messageRepository.Store(ctx, message)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}

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
