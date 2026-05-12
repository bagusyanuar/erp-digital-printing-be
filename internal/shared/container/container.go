package container

import (
	authHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	userHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler *userHttp.UserHandler
	AuthHandler *authHttp.AuthHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Container {
	return &Container{
		UserHandler: newUserHandler(db, logger),
		AuthHandler: newAuthHandler(db, cfg, logger),
	}
}
