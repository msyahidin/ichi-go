package repository

import (
	"context"

	"github.com/uptrace/bun"

	"ichi-go/pkg/db/model"
	dbRepository "ichi-go/pkg/db/repository"
)

type HealthModel struct {
	model.CoreModel
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
	*dbRepository.BaseRepository[HealthModel]
}

func NewHealthRepository(db *bun.DB) HealthRepository {
	base := dbRepository.NewRepository[HealthModel](db, &HealthModel{})
	return &RepositoryImpl{BaseRepository: base}
}

func (r *RepositoryImpl) Get(ctx context.Context) error {
	// TODO: Implement database logic
	return nil
}
