package domain

import (
	"context"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type UserRepository interface {
	GetByUsername(ctx context.Context, Username string) (*db.User, error)
	Store(ctx context.Context, u *db.User) error
}

type UserUsecase interface {
	GetByUsername(ctx context.Context, Username string) (*db.User, error)
	Store(ctx context.Context, u *db.User) error
}
