package container

import (
	authHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/auth/delivery/http"
	categoryHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/category/delivery/http"
	rbacHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/delivery/http"
	productHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/product/delivery/http"
	resellerHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/reseller/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	userHttp "github.com/bagusyanuar/erp-digital-printing-be/internal/user/delivery/http"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/casbin"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler     *userHttp.UserHandler
	AuthHandler     *authHttp.AuthHandler
	RBACHandler     *rbacHttp.RBACHandler
	ResellerHandler *resellerHttp.ResellerHandler
	CategoryHandler  *categoryHttp.CategoryHandler
	AttributeHandler *productHttp.AttributeHandler
	JWTUtil          jwt.JWTUtil
	Casbin           *casbin.CasbinHelper
}

func NewContainer(db *gorm.DB, cfg *config.Config, logger *zap.Logger) *Container {
	jwtUtil := jwt.NewJWTUtil(cfg.JWT.Secret, cfg.JWT.Issuer)
	
	csb, err := casbin.NewCasbinHelper(db, cfg.Casbin.ModelPath)
	if err != nil {
		logger.Fatal("failed to initialize casbin", zap.Error(err))
	}

	return &Container{
		UserHandler:     newUserHandler(db, logger),
		AuthHandler:     newAuthHandler(db, cfg, logger, jwtUtil),
		RBACHandler:     newRBACHandler(db, csb, logger),
		ResellerHandler: newResellerHandler(db, logger),
		CategoryHandler:  newCategoryHandler(db, logger),
		AttributeHandler: newAttributeHandler(db, logger),
		JWTUtil:          jwtUtil,
		Casbin:           csb,
	}
}
