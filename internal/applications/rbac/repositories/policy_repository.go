package repositories

import (
	"context"
	"fmt"

	"ichi-go/internal/applications/rbac/models"

	"github.com/uptrace/bun"
)

// PolicyRepository handles Casbin rule database operations
type PolicyRepository struct {
	db *bun.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *bun.DB) *PolicyRepository {
	return &PolicyRepository{
		db: db,
	}
}

// FindAll retrieves all Casbin rules
func (r *PolicyRepository) FindAll(ctx context.Context) ([]models.CasbinRule, error) {
	var rules []models.CasbinRule

	err := r.db.NewSelect().
		Model(&rules).
		Order("id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find all policies: %w", err)
	}

	return rules, nil
}

// FindByTenant retrieves all policies for a specific tenant
func (r *PolicyRepository) FindByTenant(ctx context.Context, tenantID string) ([]models.CasbinRule, error) {
	var rules []models.CasbinRule

	err := r.db.NewSelect().
		Model(&rules).
		Where("v1 = ? OR v1 = ?", tenantID, "*").
		Order("id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to find policies by tenant: %w", err)
	}

	return rules, nil
}

// FindPolicies retrieves all policy rules (ptype = 'p')
func (r *PolicyRepository) FindPolicies(ctx context.Context, tenantID string) ([]models.CasbinRule, error) {
	var rules []models.CasbinRule

	query := r.db.NewSelect().
		Model(&rules).
		Where("ptype = ?", "p")

	if tenantID != "" {
		query = query.Where("v1 = ? OR v1 = ?", tenantID, "*")
	}

	err := query.Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find policies: %w", err)
	}

	return rules, nil
}

// FindGroupings retrieves all grouping rules (ptype = 'g' or 'g2')
func (r *PolicyRepository) FindGroupings(ctx context.Context, tenantID string) ([]models.CasbinRule, error) {
	var rules []models.CasbinRule

	query := r.db.NewSelect().
		Model(&rules).
		Where("ptype IN (?)", bun.In([]string{"g", "g2"}))

	if tenantID != "" {
		query = query.Where("v1 = ? OR v2 = ?", tenantID, tenantID)
	}

	err := query.Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find groupings: %w", err)
	}

	return rules, nil
}

// FindBySubject retrieves all rules for a specific subject (user or role)
func (r *PolicyRepository) FindBySubject(ctx context.Context, subject string, tenantID string) ([]models.CasbinRule, error) {
	var rules []models.CasbinRule

	query := r.db.NewSelect().
		Model(&rules).
		Where("v0 = ?", subject)

	if tenantID != "" {
		query = query.Where("v1 = ? OR v1 = ?", tenantID, "*")
	}

	err := query.Order("id ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find rules by subject: %w", err)
	}

	return rules, nil
}

// CountByTenant counts policies for a specific tenant
func (r *PolicyRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.CasbinRule)(nil)).
		Where("v1 = ? OR v1 = ?", tenantID, "*").
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count policies by tenant: %w", err)
	}

	return count, nil
}

// CountAll counts all policies in the system
func (r *PolicyRepository) CountAll(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.CasbinRule)(nil)).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count all policies: %w", err)
	}

	return count, nil
}

// DeleteBySubject deletes all rules for a subject in a tenant
func (r *PolicyRepository) DeleteBySubject(ctx context.Context, subject string, tenantID string) error {
	query := r.db.NewDelete().
		Model((*models.CasbinRule)(nil)).
		Where("v0 = ?", subject)

	if tenantID != "" {
		query = query.Where("v1 = ?", tenantID)
	}

	_, err := query.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete rules by subject: %w", err)
	}

	return nil
}

// DeleteByTenant deletes all policies for a tenant
func (r *PolicyRepository) DeleteByTenant(ctx context.Context, tenantID string) error {
	_, err := r.db.NewDelete().
		Model((*models.CasbinRule)(nil)).
		Where("v1 = ?", tenantID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete policies by tenant: %w", err)
	}

	return nil
}
