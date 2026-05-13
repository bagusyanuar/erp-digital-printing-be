package dto

import "github.com/google/uuid"

type CreateRoleReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type AssignRoleReq struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

type AssignPermissionReq struct {
	RoleID uuid.UUID `json:"role_id" validate:"required"`
	PermID uuid.UUID `json:"permission_id" validate:"required"`
}
