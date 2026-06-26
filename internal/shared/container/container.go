package container

import (
	authHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http"
	categoryHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/category/delivery/http"
	cfHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/cashflow/delivery/http"
	orderHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/order/delivery/http"
	productHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/product/delivery/http"
	rbacHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/delivery/http"
	resellerHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	userHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http"
	expenseHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/expense/delivery/http"
	supplierHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/supplier/delivery/http"
	capitalHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/capital/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/casbin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler      *userHttp.UserHandler
	AuthHandler      *authHttp.AuthHandler
	RBACHandler      *rbacHttp.RBACHandler
	ResellerHandler  *resellerHttp.ResellerHandler
	CategoryHandler   *categoryHttp.CategoryHandler
	AttributeHandler  *productHttp.AttributeHandler
	ProductHandler    *productHttp.ProductHandler
	OrderHandler      *orderHttp.OrderHandler
	CashFlowHandler  *cfHttp.CashFlowHandler
	ExpenseHandler   *expenseHttp.ExpenseHandler
	SupplierHandler  *supplierHttp.SupplierHandler
	CapitalHandler   *capitalHttp.CapitalHandler
	JWTUtil           jwt.JWTUtil
	Casbin            *casbin.CasbinHelper
}

func NewContainer(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Container {
	jwtUtil := jwt.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.Issuer)
	
	csb, err := casbin.NewCasbinHelper(db, cfg.Casbin.ModelPath)
	if err != nil {
		logger.Fatal("failed to initialize casbin", zap.Error(err))
	}

	return &Container{
		UserHandler:      newUserHandler(db, logger),
		AuthHandler:      newAuthHandler(db, cfg, logger, jwtUtil),
		RBACHandler:      newRBACHandler(db, csb, logger),
		ResellerHandler:  newResellerHandler(db, logger),
		CategoryHandler:   newCategoryHandler(db, logger),
		AttributeHandler:  newAttributeHandler(db, logger),
		ProductHandler:    newProductHandler(db, logger),
		OrderHandler:      newOrderHandler(db, logger),
		CashFlowHandler:  newCashFlowHandler(db, logger),
		ExpenseHandler:   newExpenseHandler(db, logger),
		SupplierHandler:  newSupplierHandler(db, logger),
		CapitalHandler:   newCapitalHandler(db, logger),
		JWTUtil:           jwtUtil,
		Casbin:            csb,
	}
}


