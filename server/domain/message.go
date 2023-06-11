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
}