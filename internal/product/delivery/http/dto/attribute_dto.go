package dto

import "github.com/google/uuid"

type CreateAttributeReq struct {
	Name      string   `json:"name" validate:"required"`
	ValueType string   `json:"value_type" validate:"required,oneof=text number boolean options"`
	Options   []string `json:"options"`
}

type UpdateAttributeReq struct {
	Name      string   `json:"name" validate:"required"`
	ValueType string   `json:"value_type" validate:"required,oneof=text number boolean options"`
	Options   []string `json:"options"`
}

type AttributeOptionRes struct {
	ID    uuid.UUID `json:"id"`
	Value string    `json:"value"`
}

type AttributeRes struct {
	ID        uuid.UUID            `json:"id"`
	Name      string               `json:"name"`
	Code      string               `json:"code"`
	ValueType string               `json:"value_type"`
	Options   []AttributeOptionRes `json:"options,omitempty"`
	CreatedAt string               `json:"created_at"`
	UpdatedAt string               `json:"updated_at"`
}
