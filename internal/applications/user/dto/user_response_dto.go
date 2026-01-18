package dto

import (
	"time"

	"github.com/uptrace/bun"
)

type UserGetResponse struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt bun.NullTime `json:"updated_at"`
}
