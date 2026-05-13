package usecase

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/casbin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type rbacUsecase struct {
	repo   domain.RBACRepository
	casbin *casbin.CasbinHelper
	logger *zap.Logger
}

func NewRBACUsecase(repo domain.RBACRepository, csb *casbin.CasbinHelper, logger *zap.Logger) domain.RBACUsecase {
	return &rbacUsecase{
		repo:   repo,
		casbin: csb,
		logger: logger,
	}
}

func (u *rbacUsecase) CreateRole(ctx context.Context, role *domain.Role) error {
	return u.repo.CreateRole(ctx, role)
}

func (u *rbacUsecase) FindAllRoles(ctx context.Context) ([]domain.Role, error) {
	return u.repo.FindAllRoles(ctx)
}

func (u *rbacUsecase) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := u.repo.FindRoleByID(ctx, id)
	if err != nil {
		return err
	}

	if err := u.repo.DeleteRole(ctx, id); err != nil {
		return err
	}

	// Sync Casbin: Remove all policies (p) for this role
	if _, err := u.casbin.Enforcer.RemoveFilteredPolicy(0, role.Name); err != nil {
		u.logger.Error("failed to remove casbin policy", zap.String("role", role.Name), zap.Error(err))
	}

	// Sync Casbin: Remove all grouping (g) for this role
	if _, err := u.casbin.Enforcer.RemoveFilteredGroupingPolicy(1, role.Name); err != nil {
		u.logger.Error("failed to remove casbin grouping", zap.String("role", role.Name), zap.Error(err))
	}

	return nil
}

func (u *rbacUsecase) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	role, err := u.repo.FindRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err := u.repo.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return err
	}

	// Sync Casbin: Add grouping policy (g)
	_, err = u.casbin.AddRoleForUser(userID.String(), role.Name)
	return err
}

func (u *rbacUsecase) AssignPermissionToRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	role, err := u.repo.FindRoleByID(ctx, roleID)
	if err != nil {
		return err
	}

	perm, err := u.repo.FindPermissionByID(ctx, permID)
	if err != nil {
		return err
	}

	if err := u.repo.AssignPermissionToRole(ctx, roleID, permID); err != nil {
		return err
	}

	// Sync Casbin: Add policy (p)
	_, err = u.casbin.AddPolicy(role.Name, perm.Resource, perm.Action)
	return err
}

func (u *rbacUsecase) SyncAll(ctx context.Context) error {
	roles, err := u.repo.FindAllRoles(ctx)
	if err != nil {
		return err
	}

	u.casbin.Enforcer.ClearPolicy()

	// Add Super Admin Wildcard
	u.casbin.AddPolicy("administrator", "*", "*")

	for _, r := range roles {
		if r.Name == "administrator" {
			continue // Skip, already handled by wildcard
		}
		for _, p := range r.Permissions {
			u.casbin.AddPolicy(r.Name, p.Resource, p.Action)
		}
	}

	return u.casbin.Enforcer.SavePolicy()
}
