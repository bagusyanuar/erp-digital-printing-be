package container

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/usecase"
	userRepository "github.com/bagusyanuar/erp-digital-printing-be/internal/user/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newAuthHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *http.AuthHandler {
	// We reuse user repository for authentication
	userRepo := userRepository.NewUserRepository(db)
	
	// Create separate JWT utilities for access and refresh tokens
	accessJWT := jwt.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.Issuer)
	refreshJWT := jwt.NewJWTUtil(cfg.JWT.SecretRefresh, cfg.JWT.Issuer)

	authUsecase := usecase.NewAuthUsecase(userRepo, accessJWT, refreshJWT, cfg, logger)
	return http.NewAuthHandler(authUsecase)
}
