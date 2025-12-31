-- +goose Up
-- +goose StatementBegin

-- Create orders table
CREATE TABLE IF NOT EXISTS `orders` (
    -- Primary Key & Versioning (from CoreModel)
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `versions` BIGINT NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_by` BIGINT NOT NULL DEFAULT 0,
    `updated_by` BIGINT NOT NULL DEFAULT 0,
    `deleted_by` BIGINT NULL DEFAULT NULL,

    -- Order Identification
    `order_number` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Human-readable order number',

    -- Customer Information
    `user_id` BIGINT NOT NULL COMMENT 'Foreign key to users table',
    `user_name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL,
    `phone` VARCHAR(50) NULL,

    -- Order Status
    `status` VARCHAR(50) NOT NULL DEFAULT 'pending' COMMENT 'Order lifecycle status',

    -- Financial Information
    `subtotal_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `tax_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `shipping_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `discount_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `total_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `currency` VARCHAR(3) NOT NULL DEFAULT 'IDR',
    `refunded_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    `refundable_amount` DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    -- Payment Information
    `payment_method` VARCHAR(50) NULL,
    `payment_status` VARCHAR(50) NULL,
    `payment_provider` VARCHAR(50) NULL,
    `transaction_id` VARCHAR(255) NULL UNIQUE,
    `payment_reference` VARCHAR(255) NULL,
    `paid_at` TIMESTAMP NULL,

    -- Shipping Information
    `shipping_method` VARCHAR(50) NULL,
    `shipping_provider` VARCHAR(50) NULL,
    `tracking_number` VARCHAR(255) NULL,
    `shipping_status` VARCHAR(50) NULL,
    `shipped_at` TIMESTAMP NULL,
    `delivered_at` TIMESTAMP NULL,
    `estimated_delivery` TIMESTAMP NULL,

    -- Shipping Address
    `shipping_address_id` BIGINT NULL,
    `shipping_name` VARCHAR(255) NULL,
    `shipping_phone` VARCHAR(50) NULL,
    `shipping_address_line` TEXT NULL,
    `shipping_city` VARCHAR(100) NULL,
    `shipping_province` VARCHAR(100) NULL,
    `shipping_postal_code` VARCHAR(20) NULL,
    `shipping_country` VARCHAR(2) DEFAULT 'ID',

    -- Additional Information
    `notes` TEXT NULL COMMENT 'Customer notes',
    `internal_notes` TEXT NULL COMMENT 'Staff/admin notes',
    `cancellation_reason` TEXT NULL,
    `refund_reason` TEXT NULL,
    `cancelled_at` TIMESTAMP NULL,
    `refunded_at` TIMESTAMP NULL,

    -- Marketing & Analytics
    `source` VARCHAR(50) NULL,
    `campaign_code` VARCHAR(100) NULL,
    `referral_code` VARCHAR(100) NULL,
    `utm_source` VARCHAR(100) NULL,
    `utm_medium` VARCHAR(100) NULL,
    `utm_campaign` VARCHAR(100) NULL,

    -- Indexes
    INDEX `idx_orders_user_id` (`user_id`),
    INDEX `idx_orders_status` (`status`),
    INDEX `idx_orders_payment_status` (`payment_status`),
    INDEX `idx_orders_shipping_status` (`shipping_status`),
    INDEX `idx_orders_created_at` (`created_at`),
    INDEX `idx_orders_order_number` (`order_number`),
    INDEX `idx_orders_transaction_id` (`transaction_id`),
    INDEX `idx_orders_paid_at` (`paid_at`),
    INDEX `idx_orders_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Customer orders table';

-- Create order_items table
CREATE TABLE IF NOT EXISTS `order_items` (
    -- Primary Key & Versioning (from CoreModel)
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `versions` BIGINT NOT NULL DEFAULT 0,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `created_by` BIGINT NOT NULL DEFAULT 0,
    `updated_by` BIGINT NOT NULL DEFAULT 0,
    `deleted_by` BIGINT NULL DEFAULT NULL,

    -- Relationships
    `order_id` BIGINT NOT NULL COMMENT 'Foreign key to orders table',

    -- Product Information
    `product_id` BIGINT NOT NULL COMMENT 'Reference to products table',
    `product_sku` VARCHAR(100) NOT NULL,
    `product_name` VARCHAR(255) NOT NULL,
    `variant_id` BIGINT NULL,
    `variant_name` VARCHAR(255) NULL,

    -- Pricing
    `unit_price` DECIMAL(15,2) NOT NULL,
    `quantity` INT NOT NULL DEFAULT 1,
    `subtotal_amount` DECIMAL(15,2) NOT NULL,
    `discount_amount` DECIMAL(15,2) DEFAULT 0.00,
    `tax_amount` DECIMAL(15,2) DEFAULT 0.00,
    `total_amount` DECIMAL(15,2) NOT NULL,

    -- Fulfillment Status
    `fulfillment_status` VARCHAR(50) DEFAULT 'pending',

    -- Refund Information
    `refunded_quantity` INT DEFAULT 0,
    `refunded_amount` DECIMAL(15,2) DEFAULT 0.00,

    -- Additional Information
    `notes` TEXT NULL,
    `product_snapshot` JSON NULL COMMENT 'Product details snapshot at purchase time',

    -- Foreign Key Constraint
    CONSTRAINT `fk_order_items_order_id`
        FOREIGN KEY (`order_id`)
        REFERENCES `orders` (`id`)
        ON DELETE CASCADE
        ON UPDATE CASCADE,

    -- Indexes
    INDEX `idx_order_items_order_id` (`order_id`),
    INDEX `idx_order_items_product_id` (`product_id`),
    INDEX `idx_order_items_product_sku` (`product_sku`),
    INDEX `idx_order_items_fulfillment_status` (`fulfillment_status`),
    INDEX `idx_order_items_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Order line items table';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS `order_items`;
DROP TABLE IF EXISTS `orders`;

-- +goose StatementEnd