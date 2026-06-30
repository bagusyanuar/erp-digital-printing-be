package dto

import "github.com/google/uuid"

type CreateUserReq struct {
	Username string      `json:"username" validate:"required,min=3,max=50"`
	Password string      `json:"password" validate:"required,min=6"`
	RoleIDs  []uuid.UUID `json:"role_ids" validate:"required,dive,uuid"`
}

type UpdateUserReq struct {
	Username string      `json:"username" validate:"required,min=3,max=50"`
	Password string      `json:"password" validate:"omitempty,min=6"`
	RoleIDs  []uuid.UUID `json:"role_ids" validate:"omitempty,dive,uuid"`
}

type UserRes struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

