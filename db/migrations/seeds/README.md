# Database Seeds Directory

## Overview

This directory contains **repeatable, idempotent** SQL files that populate your database with reference data, default configurations, and test data.

## Key Concepts

### Seeds vs Migrations

| Aspect | Migrations | Seeds |
|--------|-----------|-------|
| **Purpose** | Schema changes | Data population |
| **Versioned** | Yes (tracked by goose) | No (repeatable) |
| **Timing** | Run once, in order | Run any time, multiple times |
| **Idempotent** | Should be | Must be |
| **Rollback** | Has down migration | No rollback (safe to re-run) |

### Idempotency

Seeds **must be idempotent** - safe to run multiple times without creating duplicates or errors.

**Techniques:**
- `INSERT IGNORE` - Skip if record exists
- `INSERT ... ON DUPLICATE KEY UPDATE` - Insert or update
- `REPLACE INTO` - Delete and re-insert (use cautiously)
- Check existence before insert with `WHERE NOT EXISTS`

## File Naming Convention

```
NN_descriptive_name.sql

Where:
  NN = Two-digit number (00-99) for execution order
  descriptive_name = Kebab-case description
```

**Examples:**
```
00_base_roles.sql           ✅ Runs first (core dependencies)
01_default_users.sql        ✅ Runs second (depends on roles)
02_test_data.sql           ✅ Runs third (depends on users)
03_product_categories.sql  ✅ Fourth
```

## Execution Order

Seeds run in **alphabetical order** by filename. Use the number prefix to control execution:

```
00-09: Core system data (roles, permissions, system users)
10-19: Reference data (categories, tags, statuses)
20-29: Configuration (settings, feature flags)
30-39: Sample/demo data (for development)
40-49: Test data (for QA/testing)
50-99: Reserved for custom seeds
```

## Seed File Structure

### Required Header

Every seed file should start with:

```sql
-- Seed: [Short Title]
-- Description: [What this seed does]
-- Environment: [all|local|dev|staging] (which environments to run in)
-- Idempotent: [yes|no] (can it be run multiple times safely?)
-- Dependencies: [List required tables/other seeds]
-- Author: [Team/Person name]
-- Created: [YYYY-MM-DD]
```

### Example Template

```sql
-- Seed: Product Categories
-- Description: Initialize product categorization system
-- Environment: all
-- Idempotent: yes (uses INSERT IGNORE)
-- Dependencies: Requires categories table
-- Author: Engineering Team
-- Created: 2025-01-15

-- =============================================================================
-- MAIN DATA INSERTION
-- =============================================================================

INSERT IGNORE INTO categories (id, name, slug, parent_id, created_at) VALUES
(1, 'Electronics', 'electronics', NULL, NOW()),
(2, 'Smartphones', 'smartphones', 1, NOW()),
(3, 'Laptops', 'laptops', 1, NOW());

-- =============================================================================
-- VERIFICATION (optional - commented out)
-- =============================================================================
-- SELECT * FROM categories ORDER BY parent_id, id;
```

## Best Practices

### 1. Make Seeds Truly Idempotent

**❌ Bad - Creates duplicates:**
```sql
INSERT INTO roles (name) VALUES ('Admin');
```

**✅ Good - Uses INSERT IGNORE:**
```sql
INSERT IGNORE INTO roles (id, name) VALUES (1, 'Admin');
```

**✅ Better - Uses ON DUPLICATE KEY UPDATE:**
```sql
INSERT INTO roles (id, name, description) VALUES 
(1, 'Admin', 'Administrator role')
ON DUPLICATE KEY UPDATE
    description = VALUES(description),
    updated_at = NOW();
```

### 2. Use Explicit IDs for Reference Data

**❌ Bad - Auto-increment IDs:**
```sql
INSERT INTO roles (name) VALUES ('Admin');
-- ID could be 1, 5, 17... unpredictable!
```

**✅ Good - Explicit IDs:**
```sql
INSERT IGNORE INTO roles (id, name) VALUES 
(1, 'Admin'),
(2, 'User'),
(3, 'Manager');
-- IDs are always consistent
```

### 3. Document Dependencies

```sql
-- Dependencies: 
--   - roles table must exist
--   - Run after 00_base_roles.sql
--   - Run before 02_test_data.sql
```

### 4. Use Sections for Organization

```sql
-- =============================================================================
-- SECTION 1: Core Roles
-- =============================================================================
INSERT IGNORE INTO roles ...;

-- =============================================================================
-- SECTION 2: Permissions
-- =============================================================================
INSERT IGNORE INTO permissions ...;

-- =============================================================================
-- SECTION 3: Role-Permission Mappings
-- =============================================================================
INSERT IGNORE INTO role_permissions ...;
```

### 5. Include Verification Queries

```sql
-- =============================================================================
-- VERIFICATION QUERIES (commented out)
-- =============================================================================
-- Run these to verify data was inserted correctly:

-- Count total records
-- SELECT COUNT(*) FROM categories;

-- Check hierarchy
-- SELECT c.name, p.name AS parent 
-- FROM categories c 
-- LEFT JOIN categories p ON c.parent_id = p.id;
```

### 6. Handle Environment-Specific Data

```sql
-- Sample data for local/dev only
-- Comment out or remove for production

-- Local development users (DO NOT use in production)
INSERT IGNORE INTO users (id, email, password) VALUES
(999, 'test@local.dev', '$2a$10$hashed...');
```

### 7. Use Transactions for Complex Seeds

```sql
START TRANSACTION;

INSERT IGNORE INTO categories ...;
INSERT IGNORE INTO category_attributes ...;
INSERT IGNORE INTO category_filters ...;

COMMIT;
```

## Environment-Specific Seeds

### All Environments
```sql
-- Seed: Base Roles
-- Environment: all
-- These should run in production too

INSERT IGNORE INTO roles (id, name) VALUES
(1, 'Admin'),
(2, 'User');
```

### Development Only
```sql
-- Seed: Test Users
-- Environment: local, dev
-- WARNING: Contains test data - DO NOT run in production

INSERT IGNORE INTO users (email, password) VALUES
('test@example.com', 'hashed_test_password');
```

## Common Patterns

### Pattern 1: Hierarchical Data

```sql
-- Parent categories first
INSERT IGNORE INTO categories (id, name, parent_id) VALUES
(1, 'Electronics', NULL),
(2, 'Clothing', NULL);

-- Child categories second
INSERT IGNORE INTO categories (id, name, parent_id) VALUES
(10, 'Phones', 1),
(11, 'Laptops', 1),
(20, 'Shirts', 2),
(21, 'Pants', 2);
```

### Pattern 2: Many-to-Many Relationships

```sql
-- First, seed both sides
INSERT IGNORE INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob');
INSERT IGNORE INTO roles (id, name) VALUES (1, 'Admin'), (2, 'User');

-- Then, seed the junction table
INSERT IGNORE INTO user_roles (user_id, role_id) VALUES
(1, 1), -- Alice is Admin
(2, 2); -- Bob is User
```

### Pattern 3: Configuration Data

```sql
-- Use ON DUPLICATE KEY UPDATE for settings
INSERT INTO settings (`key`, value, description) VALUES
('site_name', 'MyApp', 'Application name')
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();
```

### Pattern 4: Sample Data with Flags

```sql
-- Mark sample data so it can be removed later
INSERT IGNORE INTO products (name, is_sample_data) VALUES
('Demo Product 1', TRUE),
('Demo Product 2', TRUE);

-- To remove sample data in production:
-- DELETE FROM products WHERE is_sample_data = TRUE;
```

## Running Seeds

### Run All Seeds
```bash
make seed-run
```

### Run Specific Seed
```bash
make seed-file name=00_base_roles.sql
```

### List Available Seeds
```bash
make seed-list
```

## Troubleshooting

### Problem: Duplicate Entry Error

**Cause:** Seed is not idempotent

**Solution:** Use `INSERT IGNORE` or check for existence:
```sql
-- Before
INSERT INTO roles (id, name) VALUES (1, 'Admin');

-- After
INSERT IGNORE INTO roles (id, name) VALUES (1, 'Admin');
```

### Problem: Foreign Key Constraint Error

**Cause:** Running seeds out of order

**Solution:** 
1. Ensure proper numbering (dependencies run first)
2. Check that referenced tables exist
3. Use `SET FOREIGN_KEY_CHECKS=0` temporarily (use cautiously)

### Problem: Seeds Take Too Long

**Cause:** Large data sets

**Solution:**
- Batch inserts (insert 1000 rows at a time)
- Disable indexes temporarily
- Use transactions
- Consider splitting into multiple files

## Production Considerations

### Before Production

1. **Review all seeds** - Ensure no test data
2. **Remove debug data** - Delete entries with `is_sample_data = TRUE`
3. **Verify passwords** - Change default admin passwords
4. **Check environment flags** - Ensure only production seeds run
5. **Test rollback** - Verify you can remove seed data if needed

### In Production

1. **Backup first** - Always backup before seeding
2. **Test in staging** - Run seeds in staging environment first
3. **Monitor execution** - Watch for errors and performance
4. **Verify results** - Run verification queries
5. **Document changes** - Log what was seeded and when

## Security Notes

### Default Passwords

**❌ Never use these in production:**
```sql
INSERT INTO users (email, password) VALUES
('admin@example.com', 'admin123'); -- DANGER!
```

**✅ Use strong hashed passwords and document them:**
```sql
-- Default password: ChangeMeImmediately!
-- IMPORTANT: Change this after first login
INSERT IGNORE INTO users (email, password) VALUES
('admin@example.com', '$2a$10$properly.hashed.bcrypt.password');
```

### Sensitive Data

- Never commit API keys or secrets to seeds
- Use environment variables for sensitive config
- Document which settings need manual updates
- Audit log all seed executions in production

## Example Seeds

This directory includes these example seeds:

1. **00_base_roles.sql** - System roles
2. **01_default_users.sql** - Default admin users
3. **02_test_data.sql** - Sample users for development
4. **03_product_categories_and_tags.sql** - Product categorization
5. **04_rbac_roles_permissions.sql** - Complete RBAC system
6. **05_application_settings.sql** - App configuration

## Getting Help

- Check seed file headers for documentation
- Review example seeds for patterns
- Run with `--verbose` flag for detailed output
- Ask team lead before modifying core seeds

---

**Remember:** Seeds should be safe to run multiple times. If it's not idempotent, it shouldn't be a seed!
