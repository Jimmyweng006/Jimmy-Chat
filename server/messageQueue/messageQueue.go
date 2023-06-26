package messageQueue

import (
	"context"

	"github.com/segmentio/kafka-go"
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

func (k *Kafka) WriteMessages(ctx context.Context, messages [][]byte) error {
	kafkaMessages := make([]kafka.Message, len(messages))
	for i := 0; i < len(messages); i++ {
		kafkaMessages[i] = kafka.Message{
			Value: messages[i],
		}
	}

	err := k.writer.WriteMessages(ctx, kafkaMessages...)

	return err
}

func (k *Kafka) CloseReader() error {
	return k.reader.Close()
}

func (k *Kafka) CloseWriter() error {
	return k.writer.Close()
}
