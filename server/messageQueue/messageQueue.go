package messageQueue

import (
	"context"
)

type MessageQueue interface {
	ReadMessage(ctx context.Context) ([]byte, error)
	WriteMessages(ctx context.Context, message [][]byte) error
	CloseReader() error
	CloseWriter() error
}

// MessageQueueWrapper implement
type MessageQueueWrapper struct {
	messageQueue MessageQueue
}

func NewMessageQueueWrapper(messageQueue MessageQueue) *MessageQueueWrapper {
	return &MessageQueueWrapper{
		messageQueue: messageQueue,
	}
}

func (m *MessageQueueWrapper) ReadMessage(ctx context.Context) ([]byte, error) {
	message, err := m.messageQueue.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (m *MessageQueueWrapper) WriteMessages(ctx context.Context, messages [][]byte) error {
	err := m.messageQueue.WriteMessages(ctx, messages)

	return err
}

func (m *MessageQueueWrapper) CloseReader() error {
	return m.messageQueue.CloseReader()
}

func (m *MessageQueueWrapper) CloseWriter() error {
	return m.messageQueue.CloseWriter()
}
