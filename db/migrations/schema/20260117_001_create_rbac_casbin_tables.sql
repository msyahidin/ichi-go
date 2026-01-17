-- +goose Up
-- =============================================================================
-- RBAC System - Casbin-Based Universal Authorization
-- =============================================================================
-- Description: Create comprehensive RBAC tables for Casbin-based authorization
-- Features: Multi-tenant support, Platform permissions, Audit logging (SOC2/GDPR)
-- Author: Engineering Team
-- Created: 2026-01-17
-- =============================================================================

-- -----------------------------------------------------------------------------
-- Table 1: casbin_rule (Core Policies)
-- -----------------------------------------------------------------------------
-- Stores all Casbin policies (permissions) and groupings (role assignments)
-- Supports tenant isolation via v1 (domain) field

CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL COMMENT 'Policy type: p (policy), g (grouping), g2 (platform role)',
    v0 VARCHAR(100) COMMENT 'Subject (user_id or role)',
    v1 VARCHAR(100) COMMENT 'Domain (tenant_id or * for global)',
    v2 VARCHAR(100) COMMENT 'Object (resource)',
    v3 VARCHAR(100) COMMENT 'Action',
    v4 VARCHAR(100) COMMENT 'Reserved for future use',
    v5 VARCHAR(100) COMMENT 'Reserved for future use (e.g., expires_at)',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Performance indexes
    INDEX idx_ptype (ptype),
    INDEX idx_v1_domain (v1),
    INDEX idx_v0_subject (v0),
    INDEX idx_v2_object (v2),
    UNIQUE INDEX idx_casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Casbin policies and role assignments';

-- -----------------------------------------------------------------------------
-- Table 2: rbac_roles (Application Roles)
-- -----------------------------------------------------------------------------
-- Defines roles that can be assigned to users
-- Supports both global (tenant_id NULL) and tenant-specific roles

CREATE TABLE IF NOT EXISTS rbac_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT 'Human-readable role name',
    slug VARCHAR(100) NOT NULL COMMENT 'URL-safe identifier (e.g., admin, viewer)',
    description TEXT COMMENT 'Role description',
    tenant_id VARCHAR(100) COMMENT 'NULL for global roles, specific for tenant-custom roles',
    level INT DEFAULT 0 COMMENT 'Hierarchy level for role comparison',
    is_system_role BOOLEAN DEFAULT 0 COMMENT 'Protected system roles cannot be deleted',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT 'Soft delete timestamp',

    INDEX idx_slug (slug),
    INDEX idx_tenant (tenant_id),
    INDEX idx_deleted (deleted_at),
    UNIQUE INDEX idx_slug_tenant (slug, tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Application roles (admin, viewer, etc.)';

-- -----------------------------------------------------------------------------
-- Table 3: rbac_permissions (Permission Definitions)
-- -----------------------------------------------------------------------------
-- Master list of all available permissions in the system

CREATE TABLE IF NOT EXISTS rbac_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT 'Human-readable permission name',
    slug VARCHAR(100) NOT NULL UNIQUE COMMENT 'Unique identifier (e.g., products.view)',
    description TEXT COMMENT 'Permission description',
    module VARCHAR(50) COMMENT 'Module/domain name (e.g., products, orders)',
    resource VARCHAR(50) COMMENT 'Resource type (e.g., products, orders)',
    action VARCHAR(50) COMMENT 'Action (e.g., view, create, delete)',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_module (module),
    INDEX idx_resource_action (resource, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Permission definitions';

-- -----------------------------------------------------------------------------
-- Table 4: rbac_user_roles (User-Role Assignments)
-- -----------------------------------------------------------------------------
-- Assigns roles to users within a specific tenant context

CREATE TABLE IF NOT EXISTS rbac_user_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT 'FK to users table',
    role_id BIGINT NOT NULL COMMENT 'FK to rbac_roles table',
    tenant_id VARCHAR(100) NOT NULL COMMENT 'Tenant scope for this assignment',

    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by BIGINT COMMENT 'User ID who performed the assignment',
    expires_at TIMESTAMP NULL COMMENT 'For time-bound roles (future feature)',

    FOREIGN KEY (role_id) REFERENCES rbac_roles(id) ON DELETE CASCADE,
    INDEX idx_user_tenant (user_id, tenant_id),
    INDEX idx_role_tenant (role_id, tenant_id),
    INDEX idx_expires (expires_at),
    UNIQUE INDEX idx_user_role_tenant (user_id, role_id, tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User-role assignments (tenant-scoped)';

-- -----------------------------------------------------------------------------
-- Table 5: platform_permissions (Layer 1 - Global Platform Permissions)
-- -----------------------------------------------------------------------------
-- Platform-level permissions that bypass tenant isolation (e.g., platform.admin)

CREATE TABLE IF NOT EXISTS platform_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL COMMENT 'FK to users table',
    permission VARCHAR(100) NOT NULL COMMENT 'Platform permission (e.g., platform.admin)',

    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    granted_by BIGINT COMMENT 'User ID who granted permission',
    expires_at TIMESTAMP NULL COMMENT 'For time-bound permissions',

    INDEX idx_user_id (user_id),
    INDEX idx_permission (permission),
    INDEX idx_expires (expires_at),
    UNIQUE INDEX idx_user_permission (user_id, permission)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Platform-level global permissions';

-- -----------------------------------------------------------------------------
-- Table 6: rbac_audit_log (SOC2/GDPR Compliance)
-- -----------------------------------------------------------------------------
-- Immutable audit trail for all RBAC operations and decisions

CREATE TABLE IF NOT EXISTS rbac_audit_log (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL COMMENT 'UUID for event tracking',
    timestamp TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT 'High-precision timestamp',

    -- Actor context
    actor_id VARCHAR(255) NOT NULL COMMENT 'User/system performing the action',
    actor_type ENUM('user', 'system', 'platform_admin') NOT NULL,
    actor_email_hash VARCHAR(64) COMMENT 'SHA-256 hash for GDPR anonymization',

    -- Action details
    action ENUM(
        'policy_added',
        'policy_removed',
        'role_assigned',
        'role_revoked',
        'permission_checked',
        'permission_denied'
    ) NOT NULL,
    resource_type VARCHAR(100) COMMENT 'Type of resource (role, permission, policy)',
    resource_id VARCHAR(255) COMMENT 'Resource identifier',

    -- Target context
    subject_id VARCHAR(255) COMMENT 'User affected by the action',
    subject_email_hash VARCHAR(64) COMMENT 'SHA-256 hash for GDPR',
    tenant_id VARCHAR(255) NOT NULL COMMENT 'Tenant context',

    -- Policy details (JSON)
    policy_before JSON COMMENT 'State before change',
    policy_after JSON COMMENT 'State after change',

    -- Authorization decision
    decision ENUM('allow', 'deny') NULL COMMENT 'For permission_checked events',
    decision_reason VARCHAR(500) COMMENT 'Why permission was allowed/denied',

    -- Metadata
    reason TEXT COMMENT 'Justification for the action',
    ip_address VARCHAR(45) COMMENT 'Client IP address',
    user_agent TEXT COMMENT 'Client user agent',
    request_id VARCHAR(36) COMMENT 'Correlation ID',

    -- Performance metrics
    latency_ms INT COMMENT 'Time taken for operation',

    INDEX idx_timestamp (timestamp),
    INDEX idx_actor (actor_id, timestamp),
    INDEX idx_subject (subject_id, timestamp),
    INDEX idx_tenant (tenant_id, timestamp),
    INDEX idx_action (action, timestamp),
    INDEX idx_request (request_id),
    INDEX idx_decision (decision, timestamp)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Immutable audit log for RBAC events';

-- -----------------------------------------------------------------------------
-- Table 7: rbac_permission_groups (UI Organization)
-- -----------------------------------------------------------------------------
-- Groups permissions for easier management in admin UI

CREATE TABLE IF NOT EXISTS rbac_permission_groups (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT 'Group name (e.g., User Management)',
    slug VARCHAR(100) NOT NULL UNIQUE COMMENT 'URL-safe identifier',
    description TEXT,
    sort_order INT DEFAULT 0 COMMENT 'Display order in UI',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Permission groups for UI organization';

-- -----------------------------------------------------------------------------
-- Table 8: rbac_permission_group_items (Group Membership)
-- -----------------------------------------------------------------------------
-- Maps permissions to their groups

CREATE TABLE IF NOT EXISTS rbac_permission_group_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    group_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,

    FOREIGN KEY (group_id) REFERENCES rbac_permission_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_group_permission (group_id, permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Permission-to-group mappings';

-- -----------------------------------------------------------------------------
-- Table 9: rbac_role_permissions (Role-Permission Mapping)
-- -----------------------------------------------------------------------------
-- Denormalized mapping for quick role permission lookups (supplementary to casbin_rule)

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    tenant_id VARCHAR(100) COMMENT 'NULL for global, specific for tenant overrides',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT COMMENT 'User who created the mapping',

    FOREIGN KEY (role_id) REFERENCES rbac_roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    INDEX idx_role_tenant (role_id, tenant_id),
    INDEX idx_permission (permission_id),
    UNIQUE INDEX idx_role_permission_tenant (role_id, permission_id, tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Role-permission mappings (denormalized)';

-- -----------------------------------------------------------------------------
-- Table 10: rbac_policy_cache_metadata (Cache Management)
-- -----------------------------------------------------------------------------
-- Tracks when policies were last loaded for cache invalidation

CREATE TABLE IF NOT EXISTS rbac_policy_cache_metadata (
    tenant_id VARCHAR(100) PRIMARY KEY,
    last_loaded_at TIMESTAMP COMMENT 'When policies were last loaded into enforcer',
    policy_count INT COMMENT 'Number of policies for this tenant',
    last_modified_at TIMESTAMP COMMENT 'Last policy change timestamp',

    INDEX idx_last_modified (last_modified_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Policy cache metadata for invalidation';

-- =============================================================================
-- Initial Data
-- =============================================================================

-- Insert default permission groups
INSERT INTO rbac_permission_groups (id, name, slug, description, sort_order, created_at) VALUES
(1, 'User Management', 'user-management', 'Permissions for managing users and authentication', 1, NOW()),
(2, 'Product Management', 'product-management', 'Permissions for managing products and inventory', 2, NOW()),
(3, 'Order Management', 'order-management', 'Permissions for order processing and fulfillment', 3, NOW()),
(4, 'Content Management', 'content-management', 'Permissions for CMS and content pages', 4, NOW()),
(5, 'Analytics & Reports', 'analytics-reports', 'Permissions for viewing reports and analytics', 5, NOW()),
(6, 'System Administration', 'system-admin', 'System-level administrative permissions', 6, NOW()),
(7, 'Customer Support', 'customer-support', 'Permissions for customer service operations', 7, NOW()),
(8, 'RBAC Management', 'rbac-management', 'Permissions for managing roles and permissions', 8, NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- +goose Down

DROP TABLE IF EXISTS rbac_policy_cache_metadata;
DROP TABLE IF EXISTS rbac_role_permissions;
DROP TABLE IF EXISTS rbac_permission_group_items;
DROP TABLE IF EXISTS rbac_permission_groups;
DROP TABLE IF EXISTS rbac_audit_log;
DROP TABLE IF EXISTS platform_permissions;
DROP TABLE IF EXISTS rbac_user_roles;
DROP TABLE IF EXISTS rbac_permissions;
DROP TABLE IF EXISTS rbac_roles;
DROP TABLE IF EXISTS casbin_rule;
