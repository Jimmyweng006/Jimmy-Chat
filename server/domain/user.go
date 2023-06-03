package domain

import (
	"context"

	model "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, Username string) (*model.User, error)
	Store(ctx context.Context, u *model.User) error
}

type UserUsecase interface {
	GetByUsername(ctx context.Context, Username string) (*model.User, error)
	Store(ctx context.Context, u *model.User) error
}
