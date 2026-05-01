-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    versions BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_by BIGINT NOT NULL DEFAULT 0,
    updated_by BIGINT NOT NULL DEFAULT 0,
    deleted_by BIGINT DEFAULT NULL,

    order_number VARCHAR(50) NOT NULL UNIQUE,

    user_id BIGINT NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'pending',

    subtotal_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    refunded_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    refundable_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,

    payment_method VARCHAR(50) NULL,
    payment_status VARCHAR(50) NULL,
    payment_provider VARCHAR(50) NULL,
    transaction_id VARCHAR(255) NULL UNIQUE,
    payment_reference VARCHAR(255) NULL,
    paid_at TIMESTAMPTZ NULL,

    shipping_method VARCHAR(50) NULL,
    shipping_provider VARCHAR(50) NULL,
    tracking_number VARCHAR(255) NULL,
    shipping_status VARCHAR(50) NULL,
    shipped_at TIMESTAMPTZ NULL,
    delivered_at TIMESTAMPTZ NULL,
    estimated_delivery TIMESTAMPTZ NULL,

    shipping_address_id BIGINT NULL,
    shipping_name VARCHAR(255) NULL,
    shipping_phone VARCHAR(50) NULL,
    shipping_address_line TEXT NULL,
    shipping_city VARCHAR(100) NULL,
    shipping_province VARCHAR(100) NULL,
    shipping_postal_code VARCHAR(20) NULL,
    shipping_country VARCHAR(2) DEFAULT 'ID',

    notes TEXT NULL,
    internal_notes TEXT NULL,
    cancellation_reason TEXT NULL,
    refund_reason TEXT NULL,
    cancelled_at TIMESTAMPTZ NULL,
    refunded_at TIMESTAMPTZ NULL,

    source VARCHAR(50) NULL,
    campaign_code VARCHAR(100) NULL,
    referral_code VARCHAR(100) NULL,
    utm_source VARCHAR(100) NULL,
    utm_medium VARCHAR(100) NULL,
    utm_campaign VARCHAR(100) NULL
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_payment_status ON orders (payment_status);
CREATE INDEX idx_orders_shipping_status ON orders (shipping_status);
CREATE INDEX idx_orders_created_at ON orders (created_at);
CREATE INDEX idx_orders_order_number ON orders (order_number);
CREATE INDEX idx_orders_transaction_id ON orders (transaction_id);
CREATE INDEX idx_orders_paid_at ON orders (paid_at);
CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    versions BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    created_by BIGINT NOT NULL DEFAULT 0,
    updated_by BIGINT NOT NULL DEFAULT 0,
    deleted_by BIGINT DEFAULT NULL,

    order_id BIGINT NOT NULL,

    product_id BIGINT NOT NULL,
    product_sku VARCHAR(100) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    variant_id BIGINT NULL,
    variant_name VARCHAR(255) NULL,

    unit_price DECIMAL(15,2) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    subtotal_amount DECIMAL(15,2) NOT NULL,
    discount_amount DECIMAL(15,2) DEFAULT 0.00,
    tax_amount DECIMAL(15,2) DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL,

    fulfillment_status VARCHAR(50) DEFAULT 'pending',

    refunded_quantity INT DEFAULT 0,
    refunded_amount DECIMAL(15,2) DEFAULT 0.00,

    notes TEXT NULL,
    product_snapshot JSONB NULL,

    CONSTRAINT fk_order_items_order_id
        FOREIGN KEY (order_id)
        REFERENCES orders (id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_items_product_id ON order_items (product_id);
CREATE INDEX idx_order_items_product_sku ON order_items (product_sku);
CREATE INDEX idx_order_items_fulfillment_status ON order_items (fulfillment_status);
CREATE INDEX idx_order_items_deleted_at ON order_items (deleted_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
