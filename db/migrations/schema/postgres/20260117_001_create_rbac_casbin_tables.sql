-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ptype ON casbin_rule (ptype);
CREATE INDEX idx_v1_domain ON casbin_rule (v1);
CREATE INDEX idx_v0_subject ON casbin_rule (v0);
CREATE INDEX idx_v2_object ON casbin_rule (v2);
CREATE UNIQUE INDEX idx_casbin_rule ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

CREATE TABLE IF NOT EXISTS rbac_roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    description TEXT,
    tenant_id VARCHAR(100),
    level INT DEFAULT 0,
    is_system_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_slug ON rbac_roles (slug);
CREATE INDEX idx_tenant ON rbac_roles (tenant_id);
CREATE INDEX idx_deleted ON rbac_roles (deleted_at);
CREATE UNIQUE INDEX idx_slug_tenant ON rbac_roles (slug, tenant_id);

CREATE TABLE IF NOT EXISTS rbac_permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    module VARCHAR(50),
    resource VARCHAR(50),
    action VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_module ON rbac_permissions (module);
CREATE INDEX idx_resource_action ON rbac_permissions (resource, action);

CREATE TABLE IF NOT EXISTS rbac_user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    tenant_id VARCHAR(100) NOT NULL,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by BIGINT,
    expires_at TIMESTAMPTZ NULL,
    FOREIGN KEY (role_id) REFERENCES rbac_roles(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_tenant ON rbac_user_roles (user_id, tenant_id);
CREATE INDEX idx_role_tenant ON rbac_user_roles (role_id, tenant_id);
CREATE INDEX idx_expires ON rbac_user_roles (expires_at);
CREATE UNIQUE INDEX idx_user_role_tenant ON rbac_user_roles (user_id, role_id, tenant_id);

CREATE TABLE IF NOT EXISTS platform_permissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    permission VARCHAR(100) NOT NULL,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    granted_by BIGINT,
    expires_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_user_id ON platform_permissions (user_id);
CREATE INDEX idx_permission ON platform_permissions (permission);
CREATE INDEX idx_pp_expires ON platform_permissions (expires_at);
CREATE UNIQUE INDEX idx_user_permission ON platform_permissions (user_id, permission);

CREATE TABLE IF NOT EXISTS rbac_audit_log (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL,
    timestamp TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),

    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL CHECK (actor_type IN ('user', 'system', 'platform_admin')),
    actor_email_hash VARCHAR(64),

    action VARCHAR(50) NOT NULL CHECK (action IN (
        'policy_added',
        'policy_removed',
        'role_assigned',
        'role_revoked',
        'permission_checked',
        'permission_denied'
    )),
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),

    subject_id VARCHAR(255),
    subject_email_hash VARCHAR(64),
    tenant_id VARCHAR(255) NOT NULL,

    policy_before JSONB,
    policy_after JSONB,

    decision VARCHAR(10) CHECK (decision IN ('allow', 'deny')),
    decision_reason VARCHAR(500),

    reason TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(36),

    latency_ms INT
);

CREATE INDEX idx_audit_timestamp ON rbac_audit_log (timestamp);
CREATE INDEX idx_audit_actor ON rbac_audit_log (actor_id, timestamp);
CREATE INDEX idx_audit_subject ON rbac_audit_log (subject_id, timestamp);
CREATE INDEX idx_audit_tenant ON rbac_audit_log (tenant_id, timestamp);
CREATE INDEX idx_audit_action ON rbac_audit_log (action, timestamp);
CREATE INDEX idx_audit_request ON rbac_audit_log (request_id);
CREATE INDEX idx_audit_decision ON rbac_audit_log (decision, timestamp);

CREATE TABLE IF NOT EXISTS rbac_permission_groups (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_sort_order ON rbac_permission_groups (sort_order);

CREATE TABLE IF NOT EXISTS rbac_permission_group_items (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    FOREIGN KEY (group_id) REFERENCES rbac_permission_groups(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    UNIQUE (group_id, permission_id)
);

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    tenant_id VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by BIGINT,
    FOREIGN KEY (role_id) REFERENCES rbac_roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES rbac_permissions(id) ON DELETE CASCADE
);

CREATE INDEX idx_rp_role_tenant ON rbac_role_permissions (role_id, tenant_id);
CREATE INDEX idx_rp_permission ON rbac_role_permissions (permission_id);
CREATE UNIQUE INDEX idx_role_permission_tenant ON rbac_role_permissions (role_id, permission_id, tenant_id);

CREATE TABLE IF NOT EXISTS rbac_policy_cache_metadata (
    tenant_id VARCHAR(100) PRIMARY KEY,
    last_loaded_at TIMESTAMPTZ,
    policy_count INT,
    last_modified_at TIMESTAMPTZ
);

CREATE INDEX idx_last_modified ON rbac_policy_cache_metadata (last_modified_at);

INSERT INTO rbac_permission_groups (id, name, slug, description, sort_order, created_at) VALUES
(1, 'User Management', 'user-management', 'Permissions for managing users and authentication', 1, NOW()),
(2, 'Product Management', 'product-management', 'Permissions for managing products and inventory', 2, NOW()),
(3, 'Order Management', 'order-management', 'Permissions for order processing and fulfillment', 3, NOW()),
(4, 'Content Management', 'content-management', 'Permissions for CMS and content pages', 4, NOW()),
(5, 'Analytics & Reports', 'analytics-reports', 'Permissions for viewing reports and analytics', 5, NOW()),
(6, 'System Administration', 'system-admin', 'System-level administrative permissions', 6, NOW()),
(7, 'Customer Support', 'customer-support', 'Permissions for customer service operations', 7, NOW()),
(8, 'RBAC Management', 'rbac-management', 'Permissions for managing roles and permissions', 8, NOW())
ON CONFLICT (id) DO UPDATE SET updated_at = NOW();

SELECT setval(pg_get_serial_sequence('rbac_permission_groups', 'id'), COALESCE(MAX(id), 0), true) FROM rbac_permission_groups;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
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
-- +goose StatementEnd
