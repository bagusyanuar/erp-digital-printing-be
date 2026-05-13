package repository

import (
	"context"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/rbac/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type rbacRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) domain.RBACRepository {
	return &rbacRepository{db: db}
}

func (r *rbacRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *rbacRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id).Error
	return &role, err
}

func (r *rbacRepository) FindRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "name = ?", name).Error
	return &role, err
}

func (r *rbacRepository) FindAllRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *rbacRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Role{}, "id = ?", id).Error
}

func (r *rbacRepository) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *rbacRepository) FindAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	return perms, err
}

func (r *rbacRepository) FindPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var perm domain.Permission
	err := r.db.WithContext(ctx).First(&perm, "id = ?", id).Error
	return &perm, err
}

func (r *rbacRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Permission{}, "id = ?", id).Error
}

func (r *rbacRepository) AssignPermissionToRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("role_permissions").Create(map[string]any{
		"role_id":       roleID,
		"permission_id": permID,
	}).Error
}

func (r *rbacRepository) RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("role_permissions").
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(nil).Error
}

func (r *rbacRepository) AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("user_roles").Create(map[string]any{
		"user_id": userID,
		"role_id": roleID,
	}).Error
}

func (r *rbacRepository) RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Table("user_roles").
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(nil).Error
}
