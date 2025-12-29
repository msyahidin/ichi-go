-- Seed: Product Categories and Tags
-- Description: Initialize product categorization system with base data
-- Environment: all (required for application to function)
-- Idempotent: yes (safe to run multiple times)
-- Dependencies: Requires products, categories, and tags tables to exist
-- Author: Engineering Team
-- Created: 2025-01-15

-- =============================================================================
-- CATEGORIES
-- =============================================================================
-- Insert base product categories
-- Using INSERT IGNORE to make this idempotent (safe to run multiple times)

INSERT IGNORE INTO categories (id, name, slug, description, parent_id, sort_order, is_active, created_at) VALUES
-- Top-level categories
(1, 'Baby Care', 'baby-care', 'Products for baby care and hygiene', NULL, 1, 1, NOW()),
(2, 'Feeding', 'feeding', 'Baby feeding essentials', NULL, 2, 1, NOW()),
(3, 'Toys & Activities', 'toys-activities', 'Educational and entertainment products', NULL, 3, 1, NOW()),
(4, 'Health & Safety', 'health-safety', 'Health monitoring and safety equipment', NULL, 4, 1, NOW()),
(5, 'Nursery', 'nursery', 'Furniture and nursery essentials', NULL, 5, 1, NOW()),

-- Sub-categories for Baby Care
(101, 'Diapers & Wipes', 'diapers-wipes', 'Disposable and cloth diapers, wet wipes', 1, 1, 1, NOW()),
(102, 'Bath & Skin Care', 'bath-skin-care', 'Baby soap, shampoo, lotion, and oils', 1, 2, 1, NOW()),
(103, 'Hair Care', 'hair-care', 'Baby hair care products', 1, 3, 1, NOW()),

-- Sub-categories for Feeding
(201, 'Bottles & Nipples', 'bottles-nipples', 'Baby bottles, nipples, and accessories', 2, 1, 1, NOW()),
(202, 'Breast Pumps', 'breast-pumps', 'Manual and electric breast pumps', 2, 2, 1, NOW()),
(203, 'Formula & Supplements', 'formula-supplements', 'Baby formula and nutritional supplements', 2, 3, 1, NOW()),
(204, 'Solid Foods', 'solid-foods', 'Baby food and snacks', 2, 4, 1, NOW()),
(205, 'Feeding Accessories', 'feeding-accessories', 'Bibs, bowls, spoons, high chairs', 2, 5, 1, NOW()),

-- Sub-categories for Toys & Activities
(301, 'Developmental Toys', 'developmental-toys', 'Toys for motor skills and cognitive development', 3, 1, 1, NOW()),
(302, 'Books', 'books', 'Picture books and story books', 3, 2, 1, NOW()),
(303, 'Music & Sounds', 'music-sounds', 'Musical toys and sound books', 3, 3, 1, NOW()),

-- Sub-categories for Health & Safety
(401, 'Thermometers', 'thermometers', 'Digital and infrared thermometers', 4, 1, 1, NOW()),
(402, 'Safety Gates & Guards', 'safety-gates-guards', 'Baby gates and corner guards', 4, 2, 1, NOW()),
(403, 'Baby Monitors', 'baby-monitors', 'Audio and video baby monitors', 4, 3, 1, NOW()),

-- Sub-categories for Nursery
(501, 'Cribs & Bedding', 'cribs-bedding', 'Baby cribs, mattresses, and bedding', 5, 1, 1, NOW()),
(502, 'Storage & Organization', 'storage-organization', 'Closet organizers and storage boxes', 5, 2, 1, NOW());

-- =============================================================================
-- TAGS
-- =============================================================================
-- Insert product tags for filtering and search
-- Tags help with product discovery and filtering

INSERT IGNORE INTO tags (id, name, slug, description, tag_type, created_at) VALUES
-- Age-based tags
(1, '0-3 months', '0-3-months', 'Suitable for newborns to 3 months', 'age_range', NOW()),
(2, '3-6 months', '3-6-months', 'Suitable for babies 3-6 months', 'age_range', NOW()),
(3, '6-12 months', '6-12-months', 'Suitable for babies 6-12 months', 'age_range', NOW()),
(4, '12+ months', '12-plus-months', 'Suitable for toddlers 12 months and up', 'age_range', NOW()),

-- Feature-based tags
(10, 'Organic', 'organic', 'Made with organic materials', 'feature', NOW()),
(11, 'Hypoallergenic', 'hypoallergenic', 'Hypoallergenic formula', 'feature', NOW()),
(12, 'BPA-Free', 'bpa-free', 'Free from BPA chemicals', 'feature', NOW()),
(13, 'Eco-Friendly', 'eco-friendly', 'Environmentally friendly product', 'feature', NOW()),
(14, 'Dermatologist Tested', 'dermatologist-tested', 'Tested by dermatologists', 'feature', NOW()),

-- Brand-based tags (popular brands)
(20, 'Premium', 'premium', 'Premium quality products', 'quality', NOW()),
(21, 'Budget-Friendly', 'budget-friendly', 'Affordable options', 'quality', NOW()),
(22, 'Best Seller', 'best-seller', 'Top selling products', 'popularity', NOW()),

-- Special needs tags
(30, 'Sensitive Skin', 'sensitive-skin', 'For babies with sensitive skin', 'special_need', NOW()),
(31, 'Lactose-Free', 'lactose-free', 'Lactose-free formula', 'special_need', NOW()),
(32, 'Allergy-Friendly', 'allergy-friendly', 'Suitable for babies with allergies', 'special_need', NOW());

-- =============================================================================
-- CATEGORY-TAG RELATIONSHIPS
-- =============================================================================
-- Define which tags are commonly used with which categories
-- This helps with filtering and search functionality

INSERT IGNORE INTO category_tags (category_id, tag_id, created_at) VALUES
-- Baby Care categories commonly associated with skin-related tags
(102, 10, NOW()), -- Bath & Skin Care -> Organic
(102, 11, NOW()), -- Bath & Skin Care -> Hypoallergenic
(102, 14, NOW()), -- Bath & Skin Care -> Dermatologist Tested
(102, 30, NOW()), -- Bath & Skin Care -> Sensitive Skin

-- Feeding categories with safety tags
(201, 12, NOW()), -- Bottles & Nipples -> BPA-Free
(203, 31, NOW()), -- Formula & Supplements -> Lactose-Free
(203, 32, NOW()), -- Formula & Supplements -> Allergy-Friendly

-- All categories can have quality/popularity tags
(1, 22, NOW()),   -- Baby Care -> Best Seller
(2, 22, NOW()),   -- Feeding -> Best Seller
(3, 22, NOW()),   -- Toys & Activities -> Best Seller
(4, 22, NOW()),   -- Health & Safety -> Best Seller
(5, 22, NOW());   -- Nursery -> Best Seller

-- =============================================================================
-- SAMPLE PRODUCTS (for testing/demo purposes)
-- =============================================================================
-- Insert a few sample products to demonstrate the system
-- Mark these with a special flag so they can be easily removed in production

INSERT IGNORE INTO products (id, name, slug, description, category_id, sku, price, stock_quantity, is_sample_data, created_at) VALUES
(1, 'Organic Baby Soap', 'organic-baby-soap', 'Gentle organic soap for sensitive baby skin', 102, 'BBS-001', 12.99, 100, 1, NOW()),
(2, 'BPA-Free Baby Bottle 250ml', 'bpa-free-baby-bottle-250ml', 'Safe and durable baby bottle', 201, 'BB-250', 15.99, 150, 1, NOW()),
(3, 'Soft Musical Rattle', 'soft-musical-rattle', 'Colorful rattle with gentle sounds', 303, 'MR-001', 8.99, 200, 1, NOW()),
(4, 'Digital Baby Thermometer', 'digital-baby-thermometer', 'Fast and accurate temperature reading', 401, 'THERM-001', 24.99, 75, 1, NOW());

-- =============================================================================
-- PRODUCT-TAG RELATIONSHIPS
-- =============================================================================
-- Tag the sample products appropriately

INSERT IGNORE INTO product_tags (product_id, tag_id, created_at) VALUES
-- Organic Baby Soap tags
(1, 1, NOW()),  -- 0-3 months
(1, 2, NOW()),  -- 3-6 months
(1, 10, NOW()), -- Organic
(1, 11, NOW()), -- Hypoallergenic
(1, 14, NOW()), -- Dermatologist Tested
(1, 30, NOW()), -- Sensitive Skin

-- BPA-Free Baby Bottle tags
(2, 1, NOW()),  -- 0-3 months
(2, 2, NOW()),  -- 3-6 months
(2, 3, NOW()),  -- 6-12 months
(2, 12, NOW()), -- BPA-Free
(2, 22, NOW()), -- Best Seller

-- Musical Rattle tags
(3, 2, NOW()),  -- 3-6 months
(3, 3, NOW()),  -- 6-12 months
(3, 21, NOW()), -- Budget-Friendly

-- Digital Thermometer tags
(4, 1, NOW()),  -- 0-3 months (suitable for all ages)
(4, 2, NOW()),  -- 3-6 months
(4, 3, NOW()),  -- 6-12 months
(4, 4, NOW()),  -- 12+ months
(4, 20, NOW()); -- Premium

-- =============================================================================
-- VERIFICATION QUERIES (commented out - for reference only)
-- =============================================================================
-- Run these queries after seeding to verify data integrity

-- Check category hierarchy
-- SELECT 
--     c.id, 
--     c.name AS category, 
--     p.name AS parent_category, 
--     c.sort_order
-- FROM categories c
-- LEFT JOIN categories p ON c.parent_id = p.id
-- ORDER BY COALESCE(c.parent_id, c.id), c.sort_order;

-- Check products with their categories and tags
-- SELECT 
--     p.name AS product,
--     c.name AS category,
--     GROUP_CONCAT(t.name SEPARATOR ', ') AS tags
-- FROM products p
-- JOIN categories c ON p.category_id = c.id
-- LEFT JOIN product_tags pt ON p.id = pt.product_id
-- LEFT JOIN tags t ON pt.tag_id = t.id
-- WHERE p.is_sample_data = 1
-- GROUP BY p.id, p.name, c.name;

-- Count items by type
-- SELECT 
--     'Categories' AS item_type, COUNT(*) AS count FROM categories
-- UNION ALL
-- SELECT 'Tags', COUNT(*) FROM tags
-- UNION ALL
-- SELECT 'Products (sample)', COUNT(*) FROM products WHERE is_sample_data = 1;

-- =============================================================================
-- NOTES FOR PRODUCTION
-- =============================================================================
-- 1. Remove sample products before production deployment:
--    DELETE FROM product_tags WHERE product_id IN (SELECT id FROM products WHERE is_sample_data = 1);
--    DELETE FROM products WHERE is_sample_data = 1;
--
-- 2. Categories and tags should remain as they are core reference data
--
-- 3. Update sort_order values if you need to reorder categories
--
-- 4. This seed is safe to run multiple times due to INSERT IGNORE
--
-- 5. For production, you may want to add more categories specific to your business

-- =============================================================================
-- END OF SEED FILE
-- =============================================================================
