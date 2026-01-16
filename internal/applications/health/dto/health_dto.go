package dto

import (
	"time"

	"ichi-go/internal/applications/health/repository"
)


type HealthRequest struct {
	// TODO: Add your fields here
}


type HealthResponse struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// TODO: Add your fields here
	// Name        string    `json:"name"`
	// Description string    `json:"description,omitempty"`
}

func HealthResponseFromModel(m *repository.HealthModel) *HealthResponse {
	if m == nil {
		return nil
	}
	return &HealthResponse{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt.Time,
		// TODO: Map your fields
		// Name:        m.Name,
		// Description: m.Description,
	}
}

