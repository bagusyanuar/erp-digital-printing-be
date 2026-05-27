package dto

import "github.com/google/uuid"

type CreateResellerReq struct {
	CustomerLevelID *uuid.UUID `json:"customer_level_id"`
	Name            string     `json:"name" validate:"required"`
	Email           *string    `json:"email" validate:"omitempty,email"`
	Phone           string     `json:"phone"`
	Address         string     `json:"address"`
	CreditLimit     float64    `json:"credit_limit"`
}

type UpdateResellerReq struct {
	CustomerLevelID *uuid.UUID `json:"customer_level_id"`
	Name            string     `json:"name" validate:"required"`
	Email           *string    `json:"email" validate:"omitempty,email"`
	Phone           string     `json:"phone"`
	Address         string     `json:"address"`
	CreditLimit     float64    `json:"credit_limit"`
}

type ResellerRes struct {
	ID              uuid.UUID  `json:"id"`
	CustomerLevelID *uuid.UUID `json:"customer_level_id"`
	Name            string     `json:"name"`
	Email           *string    `json:"email"`
	Phone           string     `json:"phone"`
	Address         string     `json:"address"`
	CreditLimit     float64    `json:"credit_limit"`
	CreatedAt       string     `json:"created_at"`
	UpdatedAt       string     `json:"updated_at"`
}
