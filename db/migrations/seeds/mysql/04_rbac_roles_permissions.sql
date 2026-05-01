-- Seed: Role-Based Access Control (RBAC) System
-- Description: Initialize roles, permissions, and their relationships
-- Environment: all (required for authentication and authorization)
-- Idempotent: yes (uses INSERT IGNORE and INSERT ... ON DUPLICATE KEY UPDATE)
-- Dependencies: Requires roles, permissions, and role_permissions tables
-- Author: Engineering Team
-- Created: 2025-01-15

-- =============================================================================
-- ROLES
-- =============================================================================
-- Core system roles that define user access levels
-- id 1-100 are reserved for system roles (do not modify)
-- id 101+ can be used for custom roles

INSERT IGNORE INTO roles (id, name, slug, description, level, is_system_role, created_at) VALUES
-- System Administrator (highest privilege)
(1, 'Super Admin', 'super-admin', 
 'Full system access including user management, system configuration, and all modules', 
 100, 1, NOW()),

-- Administrative Roles
(2, 'Admin', 'admin', 
 'Administrative access to most features, cannot modify system settings', 
 90, 1, NOW()),

(3, 'Manager', 'manager', 
 'Can manage content, users within their department, and view reports', 
 70, 1, NOW()),

-- Operational Roles
(4, 'Content Editor', 'content-editor', 
 'Can create and edit content, manage products and categories', 
 50, 1, NOW()),

(5, 'Customer Service', 'customer-service', 
 'Can view and manage customer orders, respond to inquiries', 
 40, 1, NOW()),

(6, 'Inventory Manager', 'inventory-manager', 
 'Can manage product inventory, update stock levels', 
 40, 1, NOW()),

-- User Roles
(7, 'Premium User', 'premium-user', 
 'Paid subscription user with access to premium features', 
 30, 1, NOW()),

(8, 'User', 'user', 
 'Standard registered user with basic access', 
 20, 1, NOW()),

(9, 'Guest', 'guest', 
 'Limited access for non-registered users', 
 10, 1, NOW());

-- =============================================================================
-- PERMISSIONS
-- =============================================================================
-- Granular permissions organized by module
-- Format: module.action or module.resource.action

INSERT IGNORE INTO permissions (id, name, slug, description, module, created_at) VALUES
-- User Management Permissions (1-20)
(1, 'View Users', 'users.view', 'Can view user list and profiles', 'users', NOW()),
(2, 'Create Users', 'users.create', 'Can create new user accounts', 'users', NOW()),
(3, 'Edit Users', 'users.edit', 'Can modify user information', 'users', NOW()),
(4, 'Delete Users', 'users.delete', 'Can delete user accounts', 'users', NOW()),
(5, 'Manage Roles', 'users.roles.manage', 'Can assign roles to users', 'users', NOW()),

-- Product Management Permissions (21-40)
(21, 'View Products', 'products.view', 'Can view product catalog', 'products', NOW()),
(22, 'Create Products', 'products.create', 'Can add new products', 'products', NOW()),
(23, 'Edit Products', 'products.edit', 'Can modify product information', 'products', NOW()),
(24, 'Delete Products', 'products.delete', 'Can remove products', 'products', NOW()),
(25, 'Publish Products', 'products.publish', 'Can publish/unpublish products', 'products', NOW()),
(26, 'Manage Categories', 'products.categories.manage', 'Can create/edit product categories', 'products', NOW()),
(27, 'Manage Inventory', 'products.inventory.manage', 'Can update stock levels', 'products', NOW()),

-- Order Management Permissions (41-60)
(41, 'View Orders', 'orders.view', 'Can view order list and details', 'orders', NOW()),
(42, 'Create Orders', 'orders.create', 'Can create new orders (for customers)', 'orders', NOW()),
(43, 'Edit Orders', 'orders.edit', 'Can modify order information', 'orders', NOW()),
(44, 'Cancel Orders', 'orders.cancel', 'Can cancel orders', 'orders', NOW()),
(45, 'Process Orders', 'orders.process', 'Can update order status (processing, shipping, etc.)', 'orders', NOW()),
(46, 'Refund Orders', 'orders.refund', 'Can issue refunds', 'orders', NOW()),

-- Content Management Permissions (61-80)
(61, 'View Content', 'content.view', 'Can view content pages', 'content', NOW()),
(62, 'Create Content', 'content.create', 'Can create new content pages', 'content', NOW()),
(63, 'Edit Content', 'content.edit', 'Can modify content pages', 'content', NOW()),
(64, 'Delete Content', 'content.delete', 'Can remove content pages', 'content', NOW()),
(65, 'Publish Content', 'content.publish', 'Can publish/unpublish content', 'content', NOW()),

-- Analytics & Reports Permissions (81-100)
(81, 'View Reports', 'reports.view', 'Can view analytics and reports', 'reports', NOW()),
(82, 'Export Reports', 'reports.export', 'Can export reports to CSV/Excel', 'reports', NOW()),
(83, 'View Dashboard', 'dashboard.view', 'Can access admin dashboard', 'reports', NOW()),

-- System Configuration Permissions (101-120)
(101, 'System Settings', 'system.settings', 'Can modify system configuration', 'system', NOW()),
(102, 'System Logs', 'system.logs', 'Can view system logs', 'system', NOW()),
(103, 'Database Management', 'system.database', 'Can perform database operations', 'system', NOW()),

-- Customer Support Permissions (121-140)
(121, 'View Tickets', 'support.tickets.view', 'Can view support tickets', 'support', NOW()),
(122, 'Respond Tickets', 'support.tickets.respond', 'Can respond to support tickets', 'support', NOW()),
(123, 'Close Tickets', 'support.tickets.close', 'Can close support tickets', 'support', NOW());

-- =============================================================================
-- ROLE-PERMISSION RELATIONSHIPS
-- =============================================================================
-- Assign permissions to roles based on their access level

-- Super Admin - Full Access (all permissions)
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT 1, id, NOW() FROM permissions;

-- Admin - Most permissions (excluding system-level operations)
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT 2, id, NOW() FROM permissions 
WHERE id NOT IN (101, 102, 103); -- Exclude system configuration

-- Manager - Management permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- User management (view only)
(3, 1, NOW()),
-- Product management (full)
(3, 21, NOW()), (3, 22, NOW()), (3, 23, NOW()), (3, 25, NOW()), (3, 26, NOW()),
-- Order management (view and process)
(3, 41, NOW()), (3, 45, NOW()),
-- Content management (full)
(3, 61, NOW()), (3, 62, NOW()), (3, 63, NOW()), (3, 65, NOW()),
-- Reports
(3, 81, NOW()), (3, 82, NOW()), (3, 83, NOW()),
-- Support
(3, 121, NOW()), (3, 122, NOW()), (3, 123, NOW());

-- Content Editor - Content and product permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- Product management
(4, 21, NOW()), (4, 22, NOW()), (4, 23, NOW()), (4, 26, NOW()),
-- Content management
(4, 61, NOW()), (4, 62, NOW()), (4, 63, NOW()), (4, 65, NOW()),
-- Basic dashboard access
(4, 83, NOW());

-- Customer Service - Order and support permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- View users (to help customers)
(5, 1, NOW()),
-- View products (to assist with inquiries)
(5, 21, NOW()),
-- Order management (view and process, but not cancel or refund)
(5, 41, NOW()), (5, 45, NOW()),
-- Support tickets (full access)
(5, 121, NOW()), (5, 122, NOW()), (5, 123, NOW()),
-- Basic dashboard
(5, 83, NOW());

-- Inventory Manager - Inventory-specific permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- Product inventory management
(6, 21, NOW()), (6, 27, NOW()),
-- View orders (to check what's being ordered)
(6, 41, NOW()),
-- Basic dashboard
(6, 83, NOW());

-- Premium User - Enhanced user permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- Can create orders
(7, 42, NOW()),
-- Can view their own content
(7, 61, NOW());

-- User - Basic user permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- Can create orders
(8, 42, NOW());

-- Guest - Very limited (view-only)
INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at) VALUES
-- Can view products
(9, 21, NOW());

-- =============================================================================
-- DEFAULT ADMIN USERS
-- =============================================================================
-- Create default admin accounts for initial system access
-- IMPORTANT: Change these passwords immediately after first login!

-- Note: These are bcrypt hashes for the following passwords:
-- super_admin_password = Admin123!@#
-- admin_password = Admin123!

INSERT IGNORE INTO users (id, name, email, password, email_verified_at, created_at) VALUES
(1, 'System Administrator', 'superadmin@ichigo.local', 
 '$2a$10$YourBcryptHashHere', -- Change this!
 NOW(), NOW()),

(2, 'Administrator', 'admin@ichigo.local', 
 '$2a$10$YourBcryptHashHere', -- Change this!
 NOW(), NOW());

-- Assign roles to admin users
INSERT IGNORE INTO user_roles (user_id, role_id, assigned_at, assigned_by) VALUES
(1, 1, NOW(), 1), -- System Administrator gets Super Admin role
(2, 2, NOW(), 1); -- Administrator gets Admin role

-- =============================================================================
-- PERMISSION GROUPS (Optional - for UI organization)
-- =============================================================================
-- Group permissions for easier management in admin UI

INSERT IGNORE INTO permission_groups (id, name, slug, description, sort_order, created_at) VALUES
(1, 'User Management', 'user-management', 'Permissions related to user accounts', 1, NOW()),
(2, 'Product Management', 'product-management', 'Permissions related to products and inventory', 2, NOW()),
(3, 'Order Management', 'order-management', 'Permissions related to order processing', 3, NOW()),
(4, 'Content Management', 'content-management', 'Permissions related to content pages', 4, NOW()),
(5, 'Analytics & Reports', 'analytics-reports', 'Permissions related to reports and analytics', 5, NOW()),
(6, 'System Administration', 'system-admin', 'System-level permissions', 6, NOW()),
(7, 'Customer Support', 'customer-support', 'Permissions related to customer support', 7, NOW());

-- Link permissions to their groups
INSERT IGNORE INTO permission_group_items (group_id, permission_id) VALUES
-- User Management group
(1, 1), (1, 2), (1, 3), (1, 4), (1, 5),
-- Product Management group
(2, 21), (2, 22), (2, 23), (2, 24), (2, 25), (2, 26), (2, 27),
-- Order Management group
(3, 41), (3, 42), (3, 43), (3, 44), (3, 45), (3, 46),
-- Content Management group
(4, 61), (4, 62), (4, 63), (4, 64), (4, 65),
-- Analytics & Reports group
(5, 81), (5, 82), (5, 83),
-- System Administration group
(6, 101), (6, 102), (6, 103),
-- Customer Support group
(7, 121), (7, 122), (7, 123);

-- =============================================================================
-- VERIFICATION QUERIES (commented out - for reference)
-- =============================================================================

-- View all roles with their permission counts
-- SELECT 
--     r.id,
--     r.name AS role,
--     r.level,
--     COUNT(rp.permission_id) AS permission_count
-- FROM roles r
-- LEFT JOIN role_permissions rp ON r.id = rp.role_id
-- GROUP BY r.id, r.name, r.level
-- ORDER BY r.level DESC;

-- View permissions for a specific role (e.g., Manager - role_id = 3)
-- SELECT 
--     r.name AS role,
--     p.module,
--     p.name AS permission,
--     p.slug
-- FROM roles r
-- JOIN role_permissions rp ON r.id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE r.id = 3
-- ORDER BY p.module, p.name;

-- View all permissions grouped by module
-- SELECT 
--     module,
--     GROUP_CONCAT(name ORDER BY name SEPARATOR ', ') AS permissions
-- FROM permissions
-- GROUP BY module
-- ORDER BY module;

-- Check which roles have access to a specific permission (e.g., 'products.delete')
-- SELECT 
--     r.name AS role,
--     r.level,
--     p.name AS permission
-- FROM roles r
-- JOIN role_permissions rp ON r.id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE p.slug = 'products.delete'
-- ORDER BY r.level DESC;

-- =============================================================================
-- IMPORTANT SECURITY NOTES
-- =============================================================================
-- 1. CHANGE DEFAULT ADMIN PASSWORDS IMMEDIATELY AFTER FIRST LOGIN
-- 2. Never share super admin credentials
-- 3. Use environment-specific credentials (don't use these in production)
-- 4. Regularly audit user permissions and remove unused accounts
-- 5. Consider implementing 2FA for admin accounts
-- 6. Log all permission changes for audit trails
-- 7. Review and update permissions as your application evolves

-- =============================================================================
-- USAGE GUIDELINES
-- =============================================================================
-- How to check if a user has a specific permission in your application:
--
-- SELECT COUNT(*) > 0 AS has_permission
-- FROM user_roles ur
-- JOIN role_permissions rp ON ur.role_id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE ur.user_id = ? AND p.slug = ?;
--
-- How to get all permissions for a user:
--
-- SELECT DISTINCT p.slug, p.name, p.module
-- FROM user_roles ur
-- JOIN role_permissions rp ON ur.role_id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE ur.user_id = ?
-- ORDER BY p.module, p.name;

-- =============================================================================
-- END OF RBAC SEED FILE
-- =============================================================================
