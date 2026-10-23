package dto

import "github.com/google/uuid"

type CreateCategoryReq struct {
	Name          string     `json:"name" validate:"required"`
	SPKCategoryID *uuid.UUID `json:"spk_category_id,omitempty"`
}

type UpdateCategoryReq struct {
	Name          string     `json:"name" validate:"required"`
	SPKCategoryID *uuid.UUID `json:"spk_category_id,omitempty"`
}

type AssignSPKCategoryReq struct {
	SPKCategoryID *uuid.UUID `json:"spk_category_id"`
}

type CategoryRes struct {
	ID            uuid.UUID    `json:"id"`
	Name          string       `json:"name"`
	Sequence      int          `json:"sequence"`
	SPKCategoryID *uuid.UUID   `json:"spk_category_id,omitempty"`
	SPKCategory   *CategoryRes `json:"spk_category,omitempty"`
	CreatedAt     string       `json:"created_at"`
	UpdatedAt     string       `json:"updated_at"`
}
