package repository

import (
	"context"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/sirupsen/logrus"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type messageRepository struct {
	query *db.Queries
}

func NewMessageRepository(query *db.Queries) domain.MessageRepository {
	return &messageRepository{query}
}

func (m *messageRepository) Store(ctx context.Context, message *db.Message) error {
	params := db.CreateMessageParams{
		RoomID:         message.RoomID,
		ReplyMessageID: message.ReplyMessageID,
		SenderID:       message.SenderID,
		MessageText:    message.MessageText,
	}

	_, err := m.query.CreateMessage(ctx, params)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}

func (m *messageRepository) GetByRoomID(ctx context.Context, id int64) (*[]db.Message, error) {
	message, err := m.query.FindMessagesByRoomID(ctx, id)
	if err != nil {
		logrus.Error(err)
	}

	return &message, nil
}
