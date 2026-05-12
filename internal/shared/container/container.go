package container

import (
	userHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler *userHttp.UserHandler
}

func NewContainer(db *gorm.DB, logger *zap.Logger) *Container {
	return &Container{
		UserHandler: newUserHandler(db, logger),
	}
}
