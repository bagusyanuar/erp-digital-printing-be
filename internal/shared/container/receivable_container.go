package container

import (
	receivableHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/delivery/http"
	receivableRepo "github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/repository"
	receivableUseCase "github.com/bagusyanuar/erp-digital-printing-be/internal/receivable/usecase"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func newReceivableHandler(db *gorm.DB, logger *zap.Logger) *receivableHttp.ReceivableHandler {
	repo := receivableRepo.NewReceivableRepository(db)
	usecase := receivableUseCase.NewReceivableUsecase(repo, logger)
	return receivableHttp.NewReceivableHandler(usecase)
}
