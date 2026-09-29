package expensecategory

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID                 uuid.UUID `json:"id"`
	GroupID            uuid.UUID `json:"group_id"`
	Name               string    `json:"name"`
	IsActive           bool      `json:"is_active"`
	IncludeInDashboard bool      `json:"include_in_dashboard"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
