-- Seed: RBAC Casbin-Based Authorization System
-- Description: Initialize Casbin-based RBAC with permissions, roles, and policies
-- Environment: all (required for authorization)
-- Idempotent: yes (uses INSERT ... ON CONFLICT DO NOTHING)
-- Dependencies: Requires casbin_rule, rbac_roles, rbac_permissions tables

-- =============================================================================
-- RBAC PERMISSIONS (Master List)
-- =============================================================================

INSERT INTO rbac_permissions (id, name, slug, description, module, resource, action, created_at) VALUES
-- User Management Permissions (1-20)
(1, 'View Users', 'users.view', 'Can view user list and profiles', 'users', 'users', 'view', NOW()),
(2, 'Create Users', 'users.create', 'Can create new user accounts', 'users', 'users', 'create', NOW()),
(3, 'Edit Users', 'users.edit', 'Can modify user information', 'users', 'users', 'edit', NOW()),
(4, 'Delete Users', 'users.delete', 'Can delete user accounts', 'users', 'users', 'delete', NOW()),
(5, 'Manage User Roles', 'users.roles.manage', 'Can assign roles to users', 'users', 'user_roles', 'manage', NOW()),

-- Product Management Permissions (21-40)
(21, 'View Products', 'products.view', 'Can view product catalog', 'products', 'products', 'view', NOW()),
(22, 'Create Products', 'products.create', 'Can add new products', 'products', 'products', 'create', NOW()),
(23, 'Edit Products', 'products.edit', 'Can modify product information', 'products', 'products', 'edit', NOW()),
(24, 'Delete Products', 'products.delete', 'Can remove products', 'products', 'products', 'delete', NOW()),
(25, 'Publish Products', 'products.publish', 'Can publish/unpublish products', 'products', 'products', 'publish', NOW()),
(26, 'Manage Categories', 'products.categories.manage', 'Can create/edit product categories', 'products', 'categories', 'manage', NOW()),
(27, 'Manage Inventory', 'products.inventory.manage', 'Can update stock levels', 'products', 'inventory', 'manage', NOW()),

-- Order Management Permissions (41-60)
(41, 'View Orders', 'orders.view', 'Can view order list and details', 'orders', 'orders', 'view', NOW()),
(42, 'Create Orders', 'orders.create', 'Can create new orders', 'orders', 'orders', 'create', NOW()),
(43, 'Edit Orders', 'orders.edit', 'Can modify order information', 'orders', 'orders', 'edit', NOW()),
(44, 'Cancel Orders', 'orders.cancel', 'Can cancel orders', 'orders', 'orders', 'cancel', NOW()),
(45, 'Process Orders', 'orders.process', 'Can update order status', 'orders', 'orders', 'process', NOW()),
(46, 'Refund Orders', 'orders.refund', 'Can issue refunds', 'orders', 'orders', 'refund', NOW()),

-- Content Management Permissions (61-80)
(61, 'View Content', 'content.view', 'Can view content pages', 'content', 'content', 'view', NOW()),
(62, 'Create Content', 'content.create', 'Can create new content pages', 'content', 'content', 'create', NOW()),
(63, 'Edit Content', 'content.edit', 'Can modify content pages', 'content', 'content', 'edit', NOW()),
(64, 'Delete Content', 'content.delete', 'Can remove content pages', 'content', 'content', 'delete', NOW()),
(65, 'Publish Content', 'content.publish', 'Can publish/unpublish content', 'content', 'content', 'publish', NOW()),

-- Analytics & Reports Permissions (81-100)
(81, 'View Reports', 'reports.view', 'Can view analytics and reports', 'reports', 'reports', 'view', NOW()),
(82, 'Export Reports', 'reports.export', 'Can export reports to CSV/Excel', 'reports', 'reports', 'export', NOW()),
(83, 'View Dashboard', 'dashboard.view', 'Can access admin dashboard', 'reports', 'dashboard', 'view', NOW()),

-- System Configuration Permissions (101-120)
(101, 'System Settings', 'system.settings', 'Can modify system configuration', 'system', 'settings', 'edit', NOW()),
(102, 'System Logs', 'system.logs', 'Can view system logs', 'system', 'logs', 'view', NOW()),
(103, 'Database Management', 'system.database', 'Can perform database operations', 'system', 'database', 'manage', NOW()),

-- Customer Support Permissions (121-140)
(121, 'View Tickets', 'support.tickets.view', 'Can view support tickets', 'support', 'tickets', 'view', NOW()),
(122, 'Respond Tickets', 'support.tickets.respond', 'Can respond to support tickets', 'support', 'tickets', 'respond', NOW()),
(123, 'Close Tickets', 'support.tickets.close', 'Can close support tickets', 'support', 'tickets', 'close', NOW()),

-- RBAC Management Permissions (141-160)
(141, 'View Roles', 'rbac.roles.view', 'Can view roles and permissions', 'rbac', 'roles', 'view', NOW()),
(142, 'Manage Roles', 'rbac.roles.manage', 'Can create/edit/delete roles', 'rbac', 'roles', 'manage', NOW()),
(143, 'View Policies', 'rbac.policies.view', 'Can view Casbin policies', 'rbac', 'policies', 'view', NOW()),
(144, 'Manage Policies', 'rbac.policies.manage', 'Can create/edit/delete policies', 'rbac', 'policies', 'manage', NOW()),
(145, 'View Audit Log', 'rbac.audit.view', 'Can view RBAC audit log', 'rbac', 'audit', 'view', NOW()),
(146, 'Export Audit Log', 'rbac.audit.export', 'Can export audit logs', 'rbac', 'audit', 'export', NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- RBAC ROLES (Application Roles)
-- =============================================================================

INSERT INTO rbac_roles (id, name, slug, description, tenant_id, level, is_system_role, created_at) VALUES
(1, 'Super Admin', 'super-admin', 'Full system access including platform administration', NULL, 100, TRUE, NOW()),
(2, 'Admin', 'admin', 'Administrative access to most features', NULL, 90, TRUE, NOW()),
(3, 'Manager', 'manager', 'Can manage content, users, and view reports', NULL, 70, TRUE, NOW()),
(4, 'Content Editor', 'content-editor', 'Can create and edit content and products', NULL, 50, TRUE, NOW()),
(5, 'Customer Service', 'customer-service', 'Can view and manage customer orders and support', NULL, 40, TRUE, NOW()),
(6, 'Inventory Manager', 'inventory-manager', 'Can manage product inventory', NULL, 40, TRUE, NOW()),
(7, 'Viewer', 'viewer', 'Read-only access to most resources', NULL, 30, TRUE, NOW()),
(8, 'User', 'user', 'Standard registered user', NULL, 20, TRUE, NOW()),
(9, 'Guest', 'guest', 'Limited access for non-registered users', NULL, 10, TRUE, NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- ROLE-PERMISSION MAPPINGS
-- =============================================================================

-- Super Admin - All permissions
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at)
SELECT 1, id, NULL, NOW() FROM rbac_permissions
ON CONFLICT DO NOTHING;

-- Admin - Most permissions (excluding system configuration)
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at)
SELECT 2, id, NULL, NOW() FROM rbac_permissions
WHERE id NOT IN (101, 102, 103)
ON CONFLICT DO NOTHING;

-- Manager - Specific permissions
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(3, 1, NULL, NOW()),
(3, 21, NULL, NOW()),
(3, 22, NULL, NOW()),
(3, 23, NULL, NOW()),
(3, 25, NULL, NOW()),
(3, 26, NULL, NOW()),
(3, 41, NULL, NOW()),
(3, 45, NULL, NOW()),
(3, 61, NULL, NOW()),
(3, 62, NULL, NOW()),
(3, 63, NULL, NOW()),
(3, 65, NULL, NOW()),
(3, 81, NULL, NOW()),
(3, 82, NULL, NOW()),
(3, 83, NULL, NOW()),
(3, 121, NULL, NOW()),
(3, 122, NULL, NOW()),
(3, 123, NULL, NOW()),
(3, 141, NULL, NOW())
ON CONFLICT DO NOTHING;

-- Content Editor
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(4, 21, NULL, NOW()),
(4, 22, NULL, NOW()),
(4, 23, NULL, NOW()),
(4, 26, NULL, NOW()),
(4, 61, NULL, NOW()),
(4, 62, NULL, NOW()),
(4, 63, NULL, NOW()),
(4, 65, NULL, NOW()),
(4, 83, NULL, NOW())
ON CONFLICT DO NOTHING;

-- Customer Service
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(5, 1, NULL, NOW()),
(5, 21, NULL, NOW()),
(5, 41, NULL, NOW()),
(5, 45, NULL, NOW()),
(5, 83, NULL, NOW()),
(5, 121, NULL, NOW()),
(5, 122, NULL, NOW()),
(5, 123, NULL, NOW())
ON CONFLICT DO NOTHING;

-- Inventory Manager
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(6, 21, NULL, NOW()),
(6, 27, NULL, NOW()),
(6, 41, NULL, NOW()),
(6, 83, NULL, NOW())
ON CONFLICT DO NOTHING;

-- Viewer - Read-only access
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(7, 1, NULL, NOW()),
(7, 21, NULL, NOW()),
(7, 41, NULL, NOW()),
(7, 61, NULL, NOW()),
(7, 81, NULL, NOW()),
(7, 83, NULL, NOW()),
(7, 121, NULL, NOW()),
(7, 141, NULL, NOW()),
(7, 143, NULL, NOW())
ON CONFLICT DO NOTHING;

-- User - Basic permissions
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(8, 42, NULL, NOW())
ON CONFLICT DO NOTHING;

-- Guest - Very limited
INSERT INTO rbac_role_permissions (role_id, permission_id, tenant_id, created_at) VALUES
(9, 21, NULL, NOW())
ON CONFLICT DO NOTHING;

-- =============================================================================
-- CASBIN POLICIES
-- =============================================================================

-- Super Admin - Wildcard access
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'super-admin', '*', '*', '*')
ON CONFLICT DO NOTHING;

-- Admin - All resources in system tenant
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3)
SELECT 'p', 'admin', 'system', p.resource, p.action
FROM rbac_permissions p
WHERE p.id NOT IN (101, 102, 103)
ON CONFLICT DO NOTHING;

-- Manager
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'manager', 'system', 'users', 'view'),
('p', 'manager', 'system', 'products', 'view'),
('p', 'manager', 'system', 'products', 'create'),
('p', 'manager', 'system', 'products', 'edit'),
('p', 'manager', 'system', 'products', 'publish'),
('p', 'manager', 'system', 'categories', 'manage'),
('p', 'manager', 'system', 'orders', 'view'),
('p', 'manager', 'system', 'orders', 'process'),
('p', 'manager', 'system', 'content', 'view'),
('p', 'manager', 'system', 'content', 'create'),
('p', 'manager', 'system', 'content', 'edit'),
('p', 'manager', 'system', 'content', 'publish'),
('p', 'manager', 'system', 'reports', 'view'),
('p', 'manager', 'system', 'reports', 'export'),
('p', 'manager', 'system', 'dashboard', 'view'),
('p', 'manager', 'system', 'tickets', 'view'),
('p', 'manager', 'system', 'tickets', 'respond'),
('p', 'manager', 'system', 'tickets', 'close'),
('p', 'manager', 'system', 'roles', 'view')
ON CONFLICT DO NOTHING;

-- Content Editor
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'content-editor', 'system', 'products', 'view'),
('p', 'content-editor', 'system', 'products', 'create'),
('p', 'content-editor', 'system', 'products', 'edit'),
('p', 'content-editor', 'system', 'categories', 'manage'),
('p', 'content-editor', 'system', 'content', 'view'),
('p', 'content-editor', 'system', 'content', 'create'),
('p', 'content-editor', 'system', 'content', 'edit'),
('p', 'content-editor', 'system', 'content', 'publish'),
('p', 'content-editor', 'system', 'dashboard', 'view')
ON CONFLICT DO NOTHING;

-- Customer Service
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'customer-service', 'system', 'users', 'view'),
('p', 'customer-service', 'system', 'products', 'view'),
('p', 'customer-service', 'system', 'orders', 'view'),
('p', 'customer-service', 'system', 'orders', 'process'),
('p', 'customer-service', 'system', 'dashboard', 'view'),
('p', 'customer-service', 'system', 'tickets', 'view'),
('p', 'customer-service', 'system', 'tickets', 'respond'),
('p', 'customer-service', 'system', 'tickets', 'close')
ON CONFLICT DO NOTHING;

-- Inventory Manager
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'inventory-manager', 'system', 'products', 'view'),
('p', 'inventory-manager', 'system', 'inventory', 'manage'),
('p', 'inventory-manager', 'system', 'orders', 'view'),
('p', 'inventory-manager', 'system', 'dashboard', 'view')
ON CONFLICT DO NOTHING;

-- Viewer - Read-only wildcard
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'viewer', 'system', '*', 'view')
ON CONFLICT DO NOTHING;

-- User
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'user', 'system', 'orders', 'create')
ON CONFLICT DO NOTHING;

-- Guest
INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES
('p', 'guest', 'system', 'products', 'view')
ON CONFLICT DO NOTHING;

-- =============================================================================
-- PERMISSION GROUP MAPPINGS
-- =============================================================================

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(2, 21), (2, 22), (2, 23), (2, 24), (2, 25), (2, 26), (2, 27)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(3, 41), (3, 42), (3, 43), (3, 44), (3, 45), (3, 46)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(4, 61), (4, 62), (4, 63), (4, 64), (4, 65)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(5, 81), (5, 82), (5, 83)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(6, 101), (6, 102), (6, 103)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(7, 121), (7, 122), (7, 123)
ON CONFLICT DO NOTHING;

INSERT INTO rbac_permission_group_items (group_id, permission_id) VALUES
(8, 141), (8, 142), (8, 143), (8, 144), (8, 145), (8, 146)
ON CONFLICT DO NOTHING;
