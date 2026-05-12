package container

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/user/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newUserHandler(db *gorm.DB, logger *zap.Logger) *http.UserHandler {
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, logger)
	return http.NewUserHandler(userUsecase)
}
