package bun

import (
	"context"
	"errors"
	"fmt"
	"github.com/uptrace/bun"
	"ichi-go/pkg/logger"
	"time"
)

type BaseRepository[T any] struct {
	db    *bun.DB
	model *T
}

func NewRepository[T any](db *bun.DB, model *T) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:    db,
		model: model,
	}
}

func (r *BaseRepository[T]) DB() *bun.DB {
	return r.db
}

func (r *BaseRepository[T]) Query(scopes ...QueryScope) *bun.SelectQuery {
	query := r.db.NewSelect().Model(r.model)

	// Apply all scopes
	for _, scope := range scopes {
		query = scope(query)
	}

	return query
}

// Find - find by ID
func (r *BaseRepository[T]) Find(ctx context.Context, id int64) (*T, error) {
	model := new(T)
	err := r.db.NewSelect().
		Model(model).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return model, nil
}

// FindBy - find by custom field
func (r *BaseRepository[T]) FindBy(ctx context.Context, field string, value interface{}) (*T, error) {
	model := new(T)
	err := r.db.NewSelect().
		Model(model).
		Where("? = ?", bun.Ident(field), value).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return model, nil
}

// All - get all with scopes
func (r *BaseRepository[T]) All(ctx context.Context, scopes ...QueryScope) ([]*T, error) {
	var models []*T
	query := r.Query(scopes...)
	err := query.Scan(ctx, &models)
	return models, err
}

func (r *BaseRepository[T]) PaginateWithCount(ctx context.Context, page, perPage int, scopes ...QueryScope) ([]*T, int, error) {
	var models []*T

	// Build base query with scopes (excluding pagination)
	baseQuery := r.Query(scopes...)

	// Get total count
	count, err := baseQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	paginatedQuery := r.Query(append(scopes, Paginate(page, perPage))...)
	err = paginatedQuery.Scan(ctx, &models)

	return models, count, err
}

// Create - create new record
func (r *BaseRepository[T]) Create(ctx context.Context, model *T) (*T, error) {
	res, err := r.DB().NewInsert().
		Model(&model).
		Exec(ctx)
	if err != nil {
		logger.Errorf("Error user repo with  %+v, err: %+v", model, err)
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("Error getting rows affected when creating with data: %+v, err: %+v", model, err)
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("no rows inserted")
	}

	logger.Debugf("Data created with result: %+v", model)
	return model, nil
}

// Update - update existing record
func (r *BaseRepository[T]) Update(ctx context.Context, model *T) (*T, error) {
	res, err := r.db.NewUpdate().
		Model(model).
		OmitZero().
		WherePK().
		Exec(ctx)
	if err != nil {
		logger.Errorf("Error user repo with  %+v, err: %+v", model, err)
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Errorf("Error getting rows affected when updating with data: %+v, err: %+v", model, err)
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, errors.New("no rows updated")
	}

	logger.Debugf("Data updated with result: %+v", model)
	return model, nil
}

// SoftDelete - soft delete record
func (r *BaseRepository[T]) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model(r.model).
		Set("deleted_at = ?", time.Now()).
		//Set("deleted_by", 0).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Restore - restore soft deleted record
func (r *BaseRepository[T]) Restore(ctx context.Context, id int64) error {
	_, err := r.db.NewUpdate().
		Model(r.model).
		Set("deleted_at = NULL").
		//Set("updated_by", 0).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

// Count - count with scopes
func (r *BaseRepository[T]) Count(ctx context.Context, scopes ...QueryScope) (int, error) {
	query := r.Query(scopes...)
	return query.Count(ctx)
}

// Exists - check if record exists
func (r *BaseRepository[T]) Exists(ctx context.Context, scopes ...QueryScope) (bool, error) {
	count, err := r.Count(ctx, scopes...)
	return count > 0, err
}
