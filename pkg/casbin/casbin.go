package casbin

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type CasbinHelper struct {
	Enforcer *casbin.Enforcer
}

func NewCasbinHelper(db *gorm.DB, modelPath string) (*CasbinHelper, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin gorm adapter: %w", err)
	}

	enforcer, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize casbin enforcer: %w", err)
	}

	// Load policies from DB
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load casbin policy: %w", err)
	}

	return &CasbinHelper{
		Enforcer: enforcer,
	}, nil
}

func (h *CasbinHelper) CheckPermission(sub string, obj string, act string) (bool, error) {
	return h.Enforcer.Enforce(sub, obj, act)
}

func (h *CasbinHelper) AddPolicy(sub string, obj string, act string) (bool, error) {
	return h.Enforcer.AddPolicy(sub, obj, act)
}

func (h *CasbinHelper) AddRoleForUser(user string, role string) (bool, error) {
	return h.Enforcer.AddGroupingPolicy(user, role)
}
