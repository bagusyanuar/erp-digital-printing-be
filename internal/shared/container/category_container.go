package container

import (
	categoryHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/category/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/repository"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/category/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newCategoryHandler(db *gorm.DB, logger *zap.Logger) *categoryHttp.CategoryHandler {
	categoryRepo := repository.NewCategoryRepository(db)
	categoryUsecase := usecase.NewCategoryUsecase(categoryRepo, logger)
	return categoryHttp.NewCategoryHandler(categoryUsecase)
}
