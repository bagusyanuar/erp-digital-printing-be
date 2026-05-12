package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userUsecase struct {
	userRepo domain.UserRepository
	logger   *zap.Logger
}

func NewUserUsecase(userRepo domain.UserRepository, logger *zap.Logger) domain.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		logger:   logger,
	}
}

func (u *userUsecase) Create(ctx context.Context, user *domain.User) error {
	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.userRepo.FindByID(ctx, id)
}

func (u *userUsecase) FindAll(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.FindAll(ctx)
}

func (u *userUsecase) Update(ctx context.Context, user *domain.User) error {
	return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}
