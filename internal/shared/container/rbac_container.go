package container

import (
	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/usecase"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/casbin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newRBACHandler(db *gorm.DB, csb *casbin.CasbinHelper, logger *zap.Logger) *http.RBACHandler {
	repo := repository.NewRBACRepository(db)
	uc := usecase.NewRBACUsecase(repo, csb, logger)
	return http.NewRBACHandler(uc)
}
