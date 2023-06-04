package usecase

import (
	"context"

	"github.com/Jimmyweng006/Jimmy-Chat/server/domain"
	"github.com/sirupsen/logrus"

	db "github.com/Jimmyweng006/Jimmy-Chat/db/sqlc"
)

type userUsecase struct {
	userRepository domain.UserRepository
}

func NewUserUsecase(repository domain.UserRepository) domain.UserUsecase {
	return &userUsecase{repository}
}

func (u *userUsecase) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	user, err := u.userRepository.GetByUsername(ctx, username)
	if err != nil {
		logrus.Error(err)
	}

	return user, nil
}

func (u *userUsecase) Store(ctx context.Context, user *db.User) error {
	err := u.userRepository.Store(ctx, user)
	if err != nil {
		logrus.Error(err)
	}

	return nil
}
