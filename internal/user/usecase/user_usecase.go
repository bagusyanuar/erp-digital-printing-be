package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Error("Failed to hash password during user creation", zap.Error(err))
		return err
	}
	user.Password = string(hashedPassword)
	return u.userRepo.Create(ctx, user)
}

func (u *userUsecase) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return u.userRepo.FindByID(ctx, id)
}

func (u *userUsecase) FindAll(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.FindAll(ctx)
}

func (u *userUsecase) Update(ctx context.Context, user *domain.User) error {
	// If a new password is provided (not hashed yet, or we detect it's not a bcrypt hash), hash it.
	// bcrypt hashes usually start with $2a$, $2b$, or $2y$ and have a length of 60.
	if user.Password != "" && !isHashed(user.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			u.logger.Error("Failed to hash password during user update", zap.Error(err))
			return err
		}
		user.Password = string(hashedPassword)
	}
	return u.userRepo.Update(ctx, user)
}

func (u *userUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func isHashed(password string) bool {
	return len(password) == 60 && (password[:4] == "$2a$" || password[:4] == "$2b$" || password[:4] == "$2y$")
}

