package messageQueue

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MessageQueue interface {
	ReadMessage(ctx context.Context) ([]byte, error)
	WriteMessage(ctx context.Context, message []byte) error
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

func (m *MessageQueueWrapper) WriteMessage(ctx context.Context, message []byte) error {
	err := m.messageQueue.WriteMessage(ctx, message)

	return err
}

func (m *MessageQueueWrapper) CloseReader() error {
	return m.messageQueue.CloseReader()
}

func (m *MessageQueueWrapper) CloseWriter() error {
	return m.messageQueue.CloseWriter()
}

// Kafka struct implement
type Kafka struct {
	reader *kafka.Reader
	writer *kafka.Writer
}

func NewKafka(readerConfig kafka.ReaderConfig, writerConfig kafka.WriterConfig) MessageQueue {
	reader := kafka.NewReader(readerConfig)
	writer := kafka.NewWriter(writerConfig)

	return &Kafka{
		reader: reader,
		writer: writer,
	}
}

func (k *Kafka) ReadMessage(ctx context.Context) ([]byte, error) {
	message, err := k.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	return message.Value, nil
}

func (k *Kafka) WriteMessage(ctx context.Context, message []byte) error {
	err := k.writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})

	return err
}

func (k *Kafka) CloseReader() error {
	return k.reader.Close()
}

func (k *Kafka) CloseWriter() error {
	return k.writer.Close()
}
