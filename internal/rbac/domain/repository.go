package domain

import (
	"context"

	"github.com/google/uuid"
)

type RBACRepository interface {
	// Roles
	CreateRole(ctx context.Context, role *Role) error
	FindRoleByID(ctx context.Context, id uuid.UUID) (*Role, error)
	FindRoleByName(ctx context.Context, name string) (*Role, error)
	FindAllRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// Permissions
	CreatePermission(ctx context.Context, perm *Permission) error
	FindAllPermissions(ctx context.Context) ([]Permission, error)
	FindPermissionByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error

	// Role Permissions
	AssignPermissionToRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error

	// User Roles
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
}

type RBACUsecase interface {
	CreateRole(ctx context.Context, role *Role) error
	FindAllRoles(ctx context.Context) ([]Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	AssignPermissionToRole(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	SyncAll(ctx context.Context) error
}
