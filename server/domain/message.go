package domain

import (
	"context"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type MessageRepository interface {
	Store(ctx context.Context, m *db.Message) error
}

type MessageUsecase interface {
	Store(ctx context.Context, m *db.Message) error
	ReadMessageFromMessageQueue(ctx context.Context) ([]byte, error)
	WriteMessageToMessageQueue(ctx context.Context, message []byte) error
	CloseMessageQueueReader() error
	CloseMessageQueueWriter() error
}
