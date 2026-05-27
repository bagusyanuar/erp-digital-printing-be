package dto

import "github.com/google/uuid"

type CreateAttributeReq struct {
	Name      string `json:"name" validate:"required"`
	ValueType string `json:"value_type" validate:"required,oneof=text number boolean options"`
}

type UpdateAttributeReq struct {
	Name      string `json:"name" validate:"required"`
	ValueType string `json:"value_type" validate:"required,oneof=text number boolean options"`
}

type AttributeRes struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	ValueType string    `json:"value_type"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
