package usecase

import (
	"context"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/sirupsen/logrus"

	model "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type userUsercase struct {
	userRepository domain.UserRepository
}

func NewUserUsecase(repository domain.UserRepository) domain.UserUsecase {
	return &userUsercase{repository}
}

func (u *userUsercase) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := u.userRepository.GetByUsername(ctx, username)
	if err != nil {
		logrus.Error(err)
	}

	return user, nil
}

func (u *userUsercase) Store(ctx context.Context, user *model.User) error {
	err := u.userRepository.Store(ctx, user)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}
