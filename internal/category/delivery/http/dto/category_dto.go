package dto

import "github.com/google/uuid"

type CreateCategoryReq struct {
	Name string `json:"name" validate:"required"`
}

type UpdateCategoryReq struct {
	Name string `json:"name" validate:"required"`
}

type CategoryRes struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
