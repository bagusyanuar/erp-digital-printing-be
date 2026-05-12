package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/domain"
	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

type authUsecase struct {
	userRepo   userDomain.UserRepository
	accessJWT  jwt.JWTUtil
	refreshJWT jwt.JWTUtil
	cfg        *config.Config
	logger     *zap.Logger
}

func NewAuthUsecase(
	userRepo userDomain.UserRepository,
	accessJWT jwt.JWTUtil,
	refreshJWT jwt.JWTUtil,
	cfg *config.Config,
	logger *zap.Logger,
) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		accessJWT:  accessJWT,
		refreshJWT: refreshJWT,
		cfg:        cfg,
		logger:     logger,
	}
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid username or password")
	}

	// Generate Access Token
	accessToken, err := u.accessJWT.GenerateToken(
		user.ID,
		user.Username,
		time.Duration(u.cfg.JWT.Expiration)*time.Minute,
	)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshToken, err := u.refreshJWT.GenerateToken(
		user.ID,
		user.Username,
		time.Duration(u.cfg.JWT.ExpirationRefresh)*time.Hour*24,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
