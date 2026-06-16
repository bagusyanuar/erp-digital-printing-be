package dto

import "github.com/google/uuid"

type CreateSupplierReq struct {
	Name        string `json:"name" validate:"required"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
}

type UpdateSupplierReq struct {
	Name        string `json:"name" validate:"required"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Address     string `json:"address"`
}

type SupplierRes struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	ContactName string    `json:"contact_name"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}
