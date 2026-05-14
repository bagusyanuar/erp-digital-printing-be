package request

import (
	"fmt"
	"strings"
)

// PaginationParam handles query parameters for pagination and sorting
type PaginationParam struct {
	Page      int    `json:"page" query:"page"`
	Limit     int    `json:"limit" query:"limit"`
	SortBy    string `json:"sort_by" query:"sort_by"`
	SortOrder string `json:"sort_order" query:"sort_order"`
}

// GetPage returns the current page with a default value of 1 if not specified
func (p PaginationParam) GetPage() int {
	if p.Page < 1 {
		return 1
	}
	return p.Page
}

// GetLimit returns the limit with a default value of 10 if not specified
func (p PaginationParam) GetLimit() int {
	if p.Limit < 1 {
		return 10
	}
	return p.Limit
}

// GetOffset calculates the database offset based on page and limit
func (p PaginationParam) GetOffset() int {
	page := p.GetPage()
	return (page - 1) * p.GetLimit()
}

// GetSort returns the GORM compatible sort string (e.g., "id desc")
func (p PaginationParam) GetSort() string {
	if p.SortBy == "" {
		return "created_at desc"
	}

	order := strings.ToLower(p.SortOrder)
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	return fmt.Sprintf("%s %s", p.SortBy, order)
}

// FilterParam holds generic filtering criteria
type FilterParam map[string]any
