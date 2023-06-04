package usecase

import (
	"context"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/sirupsen/logrus"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type messageUsecase struct {
	messageRepository domain.MessageRepository
}

func NewMessageUsecase(repository domain.MessageRepository) domain.MessageUsecase {
	return &messageUsecase{repository}
}

func (m *messageUsecase) Store(ctx context.Context, message *db.Message) error {
	err := m.messageRepository.Store(ctx, message)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}
