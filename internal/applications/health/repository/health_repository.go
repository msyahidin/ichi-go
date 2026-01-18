package repository

import (
	"context"

	bunBase "ichi-go/internal/infra/database/bun"

	"github.com/uptrace/bun"
)

type HealthModel struct {
	bunBase.CoreModel
	bun.BaseModel `bun:"table:healths,alias:health"`

	// TODO: Add your fields here
	// Name        string    `bun:"name,notnull"`
	// Description string    `bun:"description"`
	// Price       float64   `bun:"price,notnull"`
}

type HealthRepository interface {
	Get(ctx context.Context) error
}

type RepositoryImpl struct {
	*bunBase.BaseRepository[HealthModel]
}

func NewHealthRepository(db *bun.DB) HealthRepository {
	base := bunBase.NewRepository[HealthModel](db, &HealthModel{})
	return &RepositoryImpl{BaseRepository: base}
}

func (r *RepositoryImpl) Get(ctx context.Context) error {
	// TODO: Implement database logic
	return nil
}
