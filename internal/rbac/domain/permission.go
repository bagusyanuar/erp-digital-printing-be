package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"unique;not null" json:"name"`
	Resource    string    `gorm:"not null" json:"resource"`
	Action      string    `gorm:"not null" json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
