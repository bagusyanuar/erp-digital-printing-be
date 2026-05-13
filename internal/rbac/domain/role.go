package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   *time.Time   `gorm:"index" json:"deleted_at,omitempty"`
}

type UserRole struct {
	UserID    uuid.UUID `gorm:"primaryKey" json:"user_id"`
	RoleID    uuid.UUID `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}
