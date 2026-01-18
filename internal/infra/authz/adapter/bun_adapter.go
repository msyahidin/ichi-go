package adapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"

	"github.com/uptrace/bun"
)

// CasbinRule represents a Casbin policy/grouping rule stored in database
type CasbinRule struct {
	bun.BaseModel `bun:"table:casbin_rule,alias:cr"`

	ID    int64  `bun:"id,pk,autoincrement"`
	Ptype string `bun:"ptype,notnull"` // p (policy) or g (grouping)
	V0    string `bun:"v0"`            // subject (user or role)
	V1    string `bun:"v1"`            // domain (tenant_id)
	V2    string `bun:"v2"`            // object (resource)
	V3    string `bun:"v3"`            // action
	V4    string `bun:"v4"`            // reserved
	V5    string `bun:"v5"`            // reserved
}

// BunAdapter is a Casbin adapter for Bun ORM with tenant filtering support
type BunAdapter struct {
	db     *bun.DB
	ctx    context.Context
	filter *Filter
}

// Filter defines filtering criteria for loading policies
type Filter struct {
	TenantID string   // Specific tenant to load
	Ptypes   []string // Policy types to load (e.g., ["p", "g"])
}

// NewBunAdapter creates a new Bun adapter for Casbin
func NewBunAdapter(db *bun.DB) (*BunAdapter, error) {
	if db == nil {
		return nil, errors.New("database connection is nil")
	}

	adapter := &BunAdapter{
		db:  db,
		ctx: context.Background(),
	}

	return adapter, nil
}

// LoadPolicy loads all policies from database into Casbin model
func (a *BunAdapter) LoadPolicy(model model.Model) error {
	return a.loadPolicyWithFilter(model, nil)
}

// LoadFilteredPolicy loads filtered policies based on Filter criteria
func (a *BunAdapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	if filter == nil {
		return a.LoadPolicy(model)
	}

	filterValue, ok := filter.(*Filter)
	if !ok {
		return errors.New("invalid filter type: must be *adapter.Filter")
	}

	a.filter = filterValue
	return a.loadPolicyWithFilter(model, filterValue)
}

// IsFiltered returns true if the adapter is currently using filtered loading
func (a *BunAdapter) IsFiltered() bool {
	return a.filter != nil
}

// loadPolicyWithFilter implements the actual policy loading logic
func (a *BunAdapter) loadPolicyWithFilter(model model.Model, filter *Filter) error {
	var rules []CasbinRule

	query := a.db.NewSelect().Model(&rules).Order("id ASC")

	// Apply tenant filter
	if filter != nil && filter.TenantID != "" {
		// Load policies for specific tenant + wildcard policies
		query = query.Where("v1 = ? OR v1 = ?", filter.TenantID, "*")
	}

	// Apply ptype filter
	if filter != nil && len(filter.Ptypes) > 0 {
		query = query.Where("ptype IN (?)", bun.In(filter.Ptypes))
	}

	if err := query.Scan(a.ctx); err != nil {
		return fmt.Errorf("failed to load policies: %w", err)
	}

	// Load rules into Casbin model
	for _, rule := range rules {
		if err := loadPolicyLine(&rule, model); err != nil {
			return fmt.Errorf("failed to load policy line: %w", err)
		}
	}

	return nil
}

// SavePolicy saves all policies from Casbin model to database
func (a *BunAdapter) SavePolicy(model model.Model) error {
	// Clear existing policies
	if _, err := a.db.NewDelete().Model((*CasbinRule)(nil)).Where("1=1").Exec(a.ctx); err != nil {
		return fmt.Errorf("failed to clear existing policies: %w", err)
	}

	// Save all policies from model
	var rules []CasbinRule

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			rules = append(rules, savePolicyLine(ptype, rule))
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			rules = append(rules, savePolicyLine(ptype, rule))
		}
	}

	if len(rules) == 0 {
		return nil
	}

	// Batch insert
	if _, err := a.db.NewInsert().Model(&rules).Exec(a.ctx); err != nil {
		return fmt.Errorf("failed to save policies: %w", err)
	}

	return nil
}

// AddPolicy adds a single policy rule to database
func (a *BunAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)

	if _, err := a.db.NewInsert().Model(&line).Exec(a.ctx); err != nil {
		return fmt.Errorf("failed to add policy: %w", err)
	}

	return nil
}

// RemovePolicy removes a single policy rule from database
func (a *BunAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	line := savePolicyLine(ptype, rule)

	query := a.db.NewDelete().Model(&line).Where("ptype = ?", line.Ptype)

	if line.V0 != "" {
		query = query.Where("v0 = ?", line.V0)
	}
	if line.V1 != "" {
		query = query.Where("v1 = ?", line.V1)
	}
	if line.V2 != "" {
		query = query.Where("v2 = ?", line.V2)
	}
	if line.V3 != "" {
		query = query.Where("v3 = ?", line.V3)
	}
	if line.V4 != "" {
		query = query.Where("v4 = ?", line.V4)
	}
	if line.V5 != "" {
		query = query.Where("v5 = ?", line.V5)
	}

	if _, err := query.Exec(a.ctx); err != nil {
		return fmt.Errorf("failed to remove policy: %w", err)
	}

	return nil
}

// RemoveFilteredPolicy removes policies that match the filter
func (a *BunAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := a.db.NewDelete().Model((*CasbinRule)(nil)).Where("ptype = ?", ptype)

	if fieldIndex == -1 {
		// Remove all policies of this type
		if _, err := query.Exec(a.ctx); err != nil {
			return fmt.Errorf("failed to remove filtered policy: %w", err)
		}
		return nil
	}

	// Apply field filters
	for i, value := range fieldValues {
		if value == "" {
			continue
		}

		fieldName := fmt.Sprintf("v%d", fieldIndex+i)
		query = query.Where(fmt.Sprintf("%s = ?", fieldName), value)
	}

	if _, err := query.Exec(a.ctx); err != nil {
		return fmt.Errorf("failed to remove filtered policy: %w", err)
	}

	return nil
}

// UpdatePolicy updates a policy rule (atomic remove + add)
func (a *BunAdapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	return a.UpdatePolicies(sec, ptype, [][]string{oldRule}, [][]string{newRule})
}

// UpdatePolicies updates multiple policies atomically
func (a *BunAdapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	// Use transaction for atomicity
	return a.db.RunInTx(a.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Remove old policies
		for _, rule := range oldRules {
			if err := a.RemovePolicy(sec, ptype, rule); err != nil {
				return err
			}
		}

		// Add new policies
		for _, rule := range newRules {
			if err := a.AddPolicy(sec, ptype, rule); err != nil {
				return err
			}
		}

		return nil
	})
}

// UpdateFilteredPolicies updates policies matching filter
func (a *BunAdapter) UpdateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) error {
	// Remove matching policies then add new ones
	return a.db.RunInTx(a.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if err := a.RemoveFilteredPolicy(sec, ptype, fieldIndex, fieldValues...); err != nil {
			return err
		}

		for _, rule := range newRules {
			if err := a.AddPolicy(sec, ptype, rule); err != nil {
				return err
			}
		}

		return nil
	})
}

// loadPolicyLine loads a single policy line into Casbin model
func loadPolicyLine(line *CasbinRule, model model.Model) error {
	var p = []string{line.Ptype}

	if line.V0 != "" {
		p = append(p, line.V0)
	}
	if line.V1 != "" {
		p = append(p, line.V1)
	}
	if line.V2 != "" {
		p = append(p, line.V2)
	}
	if line.V3 != "" {
		p = append(p, line.V3)
	}
	if line.V4 != "" {
		p = append(p, line.V4)
	}
	if line.V5 != "" {
		p = append(p, line.V5)
	}

	// Add to model (LoadPolicyArray expects the rule without ptype)
	return persist.LoadPolicyArray(p[1:], model)
}

// savePolicyLine converts a policy rule to CasbinRule struct
func savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{Ptype: ptype}

	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return line
}

// AddPolicies adds multiple policies in a single transaction
func (a *BunAdapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.db.RunInTx(a.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, rule := range rules {
			if err := a.AddPolicy(sec, ptype, rule); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemovePolicies removes multiple policies in a single transaction
func (a *BunAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.db.RunInTx(a.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, rule := range rules {
			if err := a.RemovePolicy(sec, ptype, rule); err != nil {
				return err
			}
		}
		return nil
	})
}

// CountPolicies returns the total number of policies in database
func (a *BunAdapter) CountPolicies() (int, error) {
	count, err := a.db.NewSelect().Model((*CasbinRule)(nil)).Count(a.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count policies: %w", err)
	}

	return count, nil
}

// CountPoliciesByTenant returns policy count for a specific tenant
func (a *BunAdapter) CountPoliciesByTenant(tenantID string) (int, error) {
	count, err := a.db.NewSelect().
		Model((*CasbinRule)(nil)).
		Where("v1 = ? OR v1 = ?", tenantID, "*").
		Count(a.ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count policies for tenant: %w", err)
	}

	return count, nil
}
