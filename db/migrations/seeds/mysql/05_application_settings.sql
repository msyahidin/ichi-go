-- Seed: Application Configuration Settings
-- Description: Initialize system-wide configuration and feature flags
-- Environment: all (customize values per environment)
-- Idempotent: yes (uses INSERT ... ON DUPLICATE KEY UPDATE)
-- Dependencies: Requires settings table
-- Author: Engineering Team
-- Created: 2025-01-15

-- =============================================================================
-- SETTINGS TABLE STRUCTURE (for reference)
-- =============================================================================
-- Expected table structure:
-- CREATE TABLE settings (
--     id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
--     category VARCHAR(50) NOT NULL,
--     key VARCHAR(100) NOT NULL,
--     value TEXT,
--     type ENUM('string', 'number', 'boolean', 'json', 'text') DEFAULT 'string',
--     description TEXT,
--     is_public BOOLEAN DEFAULT 0,
--     is_editable BOOLEAN DEFAULT 1,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
--     UNIQUE KEY unique_category_key (category, `key`)
-- );

-- =============================================================================
-- APPLICATION SETTINGS
-- =============================================================================

-- General Application Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('app', 'name', 'Ichigo Baby Care', 'string', 'Application name displayed in UI', 1, 1),
('app', 'version', '1.0.0', 'string', 'Current application version', 1, 0),
('app', 'maintenance_mode', 'false', 'boolean', 'Enable maintenance mode', 0, 1),
('app', 'timezone', 'Asia/Jakarta', 'string', 'Default application timezone', 0, 1),
('app', 'locale', 'id', 'string', 'Default application locale (id/en)', 1, 1),
('app', 'date_format', 'Y-m-d H:i:s', 'string', 'Default date format', 0, 1),
('app', 'items_per_page', '20', 'number', 'Default pagination size', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    description = VALUES(description),
    updated_at = NOW();

-- Email Configuration
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('email', 'from_name', 'Ichigo Baby Care', 'string', 'Sender name for emails', 0, 1),
('email', 'from_address', 'noreply@ichigo.app', 'string', 'Sender email address', 0, 1),
('email', 'support_email', 'support@ichigo.app', 'string', 'Support email address', 1, 1),
('email', 'enabled', 'true', 'boolean', 'Enable email sending', 0, 1),
('email', 'queue_emails', 'true', 'boolean', 'Queue emails instead of immediate send', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- SMS/Notification Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('notifications', 'sms_enabled', 'true', 'boolean', 'Enable SMS notifications', 0, 1),
('notifications', 'push_enabled', 'true', 'boolean', 'Enable push notifications', 0, 1),
('notifications', 'email_notifications', 'true', 'boolean', 'Enable email notifications', 0, 1),
('notifications', 'order_status_updates', 'true', 'boolean', 'Notify users of order status changes', 0, 1),
('notifications', 'promotional_emails', 'true', 'boolean', 'Send promotional emails', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- E-commerce Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('ecommerce', 'currency', 'IDR', 'string', 'Default currency code', 1, 1),
('ecommerce', 'currency_symbol', 'Rp', 'string', 'Currency symbol', 1, 1),
('ecommerce', 'min_order_amount', '50000', 'number', 'Minimum order amount in IDR', 1, 1),
('ecommerce', 'free_shipping_threshold', '500000', 'number', 'Free shipping above this amount', 1, 1),
('ecommerce', 'tax_rate', '11', 'number', 'Tax rate percentage (VAT/PPN)', 0, 1),
('ecommerce', 'stock_threshold', '10', 'number', 'Low stock warning threshold', 0, 1),
('ecommerce', 'enable_reviews', 'true', 'boolean', 'Allow product reviews', 1, 1),
('ecommerce', 'review_moderation', 'true', 'boolean', 'Reviews require approval', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Payment Gateway Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('payment', 'enabled_methods', '["bank_transfer", "e_wallet", "credit_card"]', 'json', 'Enabled payment methods', 0, 1),
('payment', 'bank_transfer_enabled', 'true', 'boolean', 'Enable bank transfer', 1, 1),
('payment', 'e_wallet_enabled', 'true', 'boolean', 'Enable e-wallet payments', 1, 1),
('payment', 'credit_card_enabled', 'true', 'boolean', 'Enable credit card payments', 1, 1),
('payment', 'cod_enabled', 'false', 'boolean', 'Enable cash on delivery', 1, 1),
('payment', 'auto_cancel_hours', '24', 'number', 'Auto-cancel unpaid orders after hours', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Shipping Configuration
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('shipping', 'providers', '["JNE", "J&T", "SiCepat", "Ninja Express"]', 'json', 'Available shipping providers', 1, 1),
('shipping', 'default_weight', '1000', 'number', 'Default product weight in grams', 0, 1),
('shipping', 'processing_days', '1-2', 'string', 'Order processing time', 1, 1),
('shipping', 'same_day_delivery', 'false', 'boolean', 'Offer same-day delivery', 1, 1),
('shipping', 'international_shipping', 'false', 'boolean', 'Enable international shipping', 1, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Security Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('security', 'session_lifetime', '120', 'number', 'Session lifetime in minutes', 0, 1),
('security', 'password_min_length', '8', 'number', 'Minimum password length', 0, 1),
('security', 'password_require_special', 'true', 'boolean', 'Require special characters in password', 0, 1),
('security', 'max_login_attempts', '5', 'number', 'Max login attempts before lockout', 0, 1),
('security', 'lockout_duration', '15', 'number', 'Account lockout duration in minutes', 0, 1),
('security', 'two_factor_enabled', 'false', 'boolean', 'Enable two-factor authentication', 0, 1),
('security', 'ip_whitelist', '[]', 'json', 'Whitelisted IP addresses for admin access', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Feature Flags
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('features', 'user_registration', 'true', 'boolean', 'Allow new user registration', 1, 1),
('features', 'social_login', 'true', 'boolean', 'Enable social media login', 1, 1),
('features', 'wishlist', 'true', 'boolean', 'Enable product wishlist feature', 1, 1),
('features', 'compare_products', 'true', 'boolean', 'Enable product comparison', 1, 1),
('features', 'loyalty_points', 'false', 'boolean', 'Enable loyalty points system', 1, 1),
('features', 'referral_program', 'false', 'boolean', 'Enable referral program', 1, 1),
('features', 'live_chat', 'true', 'boolean', 'Enable live chat support', 1, 1),
('features', 'blog', 'true', 'boolean', 'Enable blog/articles section', 1, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- SEO Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('seo', 'site_title', 'Ichigo - Premium Baby Care Products', 'string', 'Default page title', 1, 1),
('seo', 'meta_description', 'Your trusted source for premium baby care products in Indonesia', 'text', 'Default meta description', 1, 1),
('seo', 'meta_keywords', 'baby care, baby products, Indonesia, e-commerce', 'text', 'Default meta keywords', 1, 1),
('seo', 'og_image', '/images/og-default.jpg', 'string', 'Default Open Graph image', 1, 1),
('seo', 'enable_sitemap', 'true', 'boolean', 'Generate XML sitemap', 0, 1),
('seo', 'robots_txt_enabled', 'true', 'boolean', 'Enable robots.txt', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Social Media Links
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('social', 'facebook_url', 'https://facebook.com/ichigo', 'string', 'Facebook page URL', 1, 1),
('social', 'instagram_url', 'https://instagram.com/ichigo', 'string', 'Instagram profile URL', 1, 1),
('social', 'twitter_url', 'https://twitter.com/ichigo', 'string', 'Twitter profile URL', 1, 1),
('social', 'youtube_url', 'https://youtube.com/ichigo', 'string', 'YouTube channel URL', 1, 1),
('social', 'whatsapp_number', '+628123456789', 'string', 'WhatsApp customer service number', 1, 1),
('social', 'tiktok_url', '', 'string', 'TikTok profile URL', 1, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Analytics & Tracking
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('analytics', 'google_analytics_id', '', 'string', 'Google Analytics tracking ID', 0, 1),
('analytics', 'facebook_pixel_id', '', 'string', 'Facebook Pixel ID', 0, 1),
('analytics', 'enable_tracking', 'true', 'boolean', 'Enable analytics tracking', 0, 1),
('analytics', 'track_conversions', 'true', 'boolean', 'Track e-commerce conversions', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Cache & Performance
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('cache', 'enabled', 'true', 'boolean', 'Enable application caching', 0, 1),
('cache', 'driver', 'redis', 'string', 'Cache driver (redis/file)', 0, 1),
('cache', 'ttl', '3600', 'number', 'Default cache TTL in seconds', 0, 1),
('cache', 'product_cache_ttl', '7200', 'number', 'Product data cache TTL', 0, 1),
('cache', 'category_cache_ttl', '86400', 'number', 'Category data cache TTL', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Image & Media Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('media', 'max_upload_size', '5', 'number', 'Maximum upload size in MB', 0, 1),
('media', 'allowed_extensions', '["jpg", "jpeg", "png", "gif", "webp"]', 'json', 'Allowed image extensions', 0, 1),
('media', 'image_quality', '85', 'number', 'Image compression quality (1-100)', 0, 1),
('media', 'generate_thumbnails', 'true', 'boolean', 'Auto-generate image thumbnails', 0, 1),
('media', 'thumbnail_sizes', '{"small": [150, 150], "medium": [300, 300], "large": [800, 800]}', 'json', 'Thumbnail dimensions', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- Customer Support Settings
INSERT INTO settings (category, `key`, value, type, description, is_public, is_editable) VALUES
('support', 'business_hours', '{"mon-fri": "09:00-18:00", "sat": "09:00-15:00", "sun": "closed"}', 'json', 'Customer service hours', 1, 1),
('support', 'response_time', '24 hours', 'string', 'Expected response time', 1, 1),
('support', 'auto_reply_enabled', 'true', 'boolean', 'Send automatic acknowledgment', 0, 1),
('support', 'ticket_categories', '["Product Inquiry", "Order Issue", "Payment", "Shipping", "Returns", "Technical", "Other"]', 'json', 'Support ticket categories', 0, 1)
ON DUPLICATE KEY UPDATE 
    value = VALUES(value),
    updated_at = NOW();

-- =============================================================================
-- ENVIRONMENT-SPECIFIC OVERRIDES
-- =============================================================================
-- Uncomment and modify these for different environments

-- PRODUCTION OVERRIDES
-- UPDATE settings SET value = 'true' WHERE category = 'cache' AND `key` = 'enabled';
-- UPDATE settings SET value = 'false' WHERE category = 'app' AND `key` = 'maintenance_mode';
-- UPDATE settings SET value = 'false' WHERE category = 'features' AND `key` = 'loyalty_points';

-- STAGING OVERRIDES
-- UPDATE settings SET value = 'staging@ichigo.app' WHERE category = 'email' AND `key` = 'from_address';
-- UPDATE settings SET value = 'false' WHERE category = 'payment' AND `key` = 'credit_card_enabled';

-- LOCAL DEVELOPMENT OVERRIDES
-- UPDATE settings SET value = 'true' WHERE category = 'app' AND `key` = 'debug';
-- UPDATE settings SET value = 'false' WHERE category = 'cache' AND `key` = 'enabled';
-- UPDATE settings SET value = 'local@ichigo.app' WHERE category = 'email' AND `key` = 'from_address';

-- =============================================================================
-- HELPER QUERIES (commented out - for reference)
-- =============================================================================

-- Get all public settings (safe to expose in frontend)
-- SELECT category, `key`, value, type, description
-- FROM settings
-- WHERE is_public = 1
-- ORDER BY category, `key`;

-- Get settings by category
-- SELECT `key`, value, type, description
-- FROM settings
-- WHERE category = 'ecommerce'
-- ORDER BY `key`;

-- Get editable settings for admin panel
-- SELECT category, `key`, value, type, description
-- FROM settings
-- WHERE is_editable = 1
-- ORDER BY category, `key`;

-- Get feature flags
-- SELECT `key`, value, description
-- FROM settings
-- WHERE category = 'features'
-- ORDER BY `key`;

-- =============================================================================
-- USAGE IN APPLICATION
-- =============================================================================
-- To use these settings in your Go application, create a settings service:
--
-- func GetSetting(category, key string) (string, error) {
--     var setting Setting
--     err := db.Where("category = ? AND `key` = ?", category, key).First(&setting).Error
--     return setting.Value, err
-- }
--
-- func GetPublicSettings() ([]Setting, error) {
--     var settings []Setting
--     err := db.Where("is_public = ?", true).Find(&settings).Error
--     return settings, err
-- }
--
-- func UpdateSetting(category, key, value string) error {
--     return db.Model(&Setting{}).
--         Where("category = ? AND `key` = ? AND is_editable = ?", category, key, true).
--         Update("value", value).Error
-- }

-- =============================================================================
-- NOTES
-- =============================================================================
-- 1. Use ON DUPLICATE KEY UPDATE to make this seed idempotent
-- 2. Public settings can be cached and exposed to frontend
-- 3. Non-editable settings prevent accidental changes via admin panel
-- 4. JSON values should be properly escaped when updating
-- 5. Consider caching frequently accessed settings
-- 6. Audit log all setting changes in production
-- 7. Back up settings before major updates

-- =============================================================================
-- END OF SETTINGS SEED FILE
-- =============================================================================
