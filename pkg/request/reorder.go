package request

import "github.com/google/uuid"

type ReorderItem struct {
	ID       uuid.UUID `json:"id" validate:"required"`
	Sequence int       `json:"sequence" validate:"min=0"`
}

type ReorderRequest struct {
	Items []ReorderItem `json:"items" validate:"required,min=1"`
}
