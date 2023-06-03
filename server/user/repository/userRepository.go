package repository

import (
	"context"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/sirupsen/logrus"

	model "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
	query "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type userRepository struct {
	query *query.Queries
}

func NewUserRepository(query *query.Queries) domain.UserRepository {
	return &userRepository{query}
}

func (u *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := u.query.FindUser(ctx, username)
	if err != nil {
		logrus.Error(err)
	}

	return &user, nil
}

func (u *userRepository) Store(ctx context.Context, user *model.User) error {
	params := query.CreateUserParams{
		Username: user.Username,
		Password: user.Password,
	}

	_, err := u.query.CreateUser(ctx, params)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}
