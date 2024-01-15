package domain

import (
	"context"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type MessageRepository interface {
	Store(ctx context.Context, m *db.Message) error
	GetByRoomID(ctx context.Context, id int64) (*[]db.Message, error)
}

type MessageUsecase interface {
	Store(ctx context.Context, m *db.Message) error
	RetriveHistoryMessage(ctx context.Context, roomID int64, u UserUsecase) (*[]MessageDTO, error)
	ReadMessageFromMessageQueue(ctx context.Context) ([]byte, error)
	WriteMessagesToMessageQueue(ctx context.Context, messages [][]byte) error
	CloseMessageQueueReader() error
	CloseMessageQueueWriter() error
}

type MessageDTO struct {
	Sender  string
	Content string
}
