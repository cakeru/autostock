-- 000001_init_schema.up.sql
-- AutoStock initial schema: all tables, indexes, and seed data

BEGIN;

-- =============================================================================
-- Core Tables
-- =============================================================================

CREATE TABLE branches (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    logo_url VARCHAR(500),
    tax_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_branches_is_active ON branches(is_active);

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'staff')),
    permissions JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_users_branch_id ON users(branch_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);

CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(branch_id, key)
);
CREATE INDEX idx_settings_branch_id ON settings(branch_id);
CREATE INDEX idx_settings_key ON settings(key);

-- =============================================================================
-- Inventory Tables
-- =============================================================================

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('tire', 'part', 'labor', 'consumable')),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    buy_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    sell_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    min_stock_alert INTEGER DEFAULT 5,
    unit VARCHAR(50) DEFAULT 'piece',
    tire_size VARCHAR(50),
    tire_brand VARCHAR(100),
    tire_model VARCHAR(100),
    tire_pattern VARCHAR(100),
    dot_code VARCHAR(50),
    load_index VARCHAR(10),
    speed_rating VARCHAR(5),
    tire_type VARCHAR(50),
    location VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_products_branch_id ON products(branch_id);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_tire_size ON products(tire_size) WHERE type = 'tire';
CREATE INDEX idx_products_tire_brand ON products(tire_brand) WHERE type = 'tire';
CREATE INDEX idx_products_stock_quantity ON products(stock_quantity);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_branch_type_active ON products(branch_id, type, is_active);

-- =============================================================================
-- Customer Tables
-- =============================================================================

CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    notes TEXT,
    customer_since DATE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_customers_branch_id ON customers(branch_id);
CREATE INDEX idx_customers_name ON customers(name);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_is_active ON customers(is_active);

CREATE TABLE vehicles (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    plate_number VARCHAR(50) NOT NULL,
    make VARCHAR(100),
    model VARCHAR(100),
    year INTEGER CHECK (year >= 1900 AND year <= 2100),
    vin VARCHAR(17),
    color VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(customer_id, plate_number)
);
CREATE INDEX idx_vehicles_customer_id ON vehicles(customer_id);
CREATE INDEX idx_vehicles_plate_number ON vehicles(plate_number);
CREATE INDEX idx_vehicles_make_model ON vehicles(make, model);

-- =============================================================================
-- Service & Invoice Tables
-- =============================================================================

CREATE TABLE service_jobs (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE SET NULL,
    invoice_id BIGINT,
    job_number VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    priority VARCHAR(50) DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    description TEXT,
    diagnosis TEXT,
    work_performed TEXT,
    estimated_hours DECIMAL(5, 2),
    actual_hours DECIMAL(5, 2),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_service_jobs_branch_id ON service_jobs(branch_id);
CREATE INDEX idx_service_jobs_customer_id ON service_jobs(customer_id);
CREATE INDEX idx_service_jobs_vehicle_id ON service_jobs(vehicle_id);
CREATE INDEX idx_service_jobs_invoice_id ON service_jobs(invoice_id);
CREATE INDEX idx_service_jobs_status ON service_jobs(status);
CREATE INDEX idx_service_jobs_job_number ON service_jobs(job_number);
CREATE INDEX idx_service_jobs_created_at ON service_jobs(created_at);

CREATE TABLE service_job_items (
    id BIGSERIAL PRIMARY KEY,
    service_job_id BIGINT NOT NULL REFERENCES service_jobs(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    description VARCHAR(255),
    quantity DECIMAL(10, 2) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK (quantity > 0),
    CHECK (unit_price >= 0),
    CHECK (total_price >= 0)
);
CREATE INDEX idx_service_job_items_service_job_id ON service_job_items(service_job_id);
CREATE INDEX idx_service_job_items_product_id ON service_job_items(product_id);

CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    invoice_number VARCHAR(50) NOT NULL UNIQUE,
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5, 2) DEFAULT 0,
    tax_amount DECIMAL(10, 2) DEFAULT 0,
    discount DECIMAL(10, 2) DEFAULT 0,
    total_usd DECIMAL(10, 2) NOT NULL DEFAULT 0,
    exchange_rate DECIMAL(10, 2) NOT NULL,
    total_khr DECIMAL(10, 2) NOT NULL DEFAULT 0,
    payment_status VARCHAR(50) NOT NULL DEFAULT 'unpaid' CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'refunded', 'voided')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'issued', 'paid', 'voided')),
    paid_amount DECIMAL(10, 2) DEFAULT 0,
    payment_method VARCHAR(50),
    payment_notes TEXT,
    notes TEXT,
    terms TEXT,
    voided_at TIMESTAMP WITH TIME ZONE,
    void_reason TEXT,
    voided_by BIGINT REFERENCES users(id),
    issued_at TIMESTAMP WITH TIME ZONE,
    due_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_invoices_branch_id ON invoices(branch_id);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_vehicle_id ON invoices(vehicle_id);
CREATE INDEX idx_invoices_service_job_id ON invoices(service_job_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_payment_status ON invoices(payment_status);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);
CREATE INDEX idx_invoices_issued_at ON invoices(issued_at);

CREATE TABLE invoice_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('product', 'labor', 'custom')),
    description VARCHAR(500) NOT NULL,
    quantity DECIMAL(10, 2) NOT NULL DEFAULT 1,
    unit_price_usd DECIMAL(10, 2) NOT NULL,
    total_usd DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CHECK (quantity > 0),
    CHECK (unit_price_usd >= 0),
    CHECK (total_usd >= 0)
);
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);
CREATE INDEX idx_invoice_items_product_id ON invoice_items(product_id);

-- =============================================================================
-- Audit Tables
-- =============================================================================

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id BIGINT,
    old_values JSONB,
    new_values JSONB,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX idx_audit_logs_branch_id ON audit_logs(branch_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id, created_at DESC);

-- =============================================================================
-- Seed Data
-- =============================================================================

-- Default branch
INSERT INTO branches (name, address, phone) VALUES
('AutoStock Garage', 'Phnom Penh, Cambodia', '+855 12 345 678');

-- Admin user (password: admin123)
INSERT INTO users (branch_id, username, password_hash, full_name, role, permissions) VALUES
(1, 'admin',
 '$2a$12$NbStBtjAPLePbJsK4v9c/.GJUme2amTx48imqIx8FNg6kWf45QyyO',
 'Administrator', 'admin',
 '["inventory:view","inventory:create","inventory:update","inventory:delete","customer:view","customer:create","customer:update","customer:delete","service:view","service:create","service:update","service:delete","invoice:view","invoice:create","invoice:void","user:view","user:create","user:update","user:delete","settings:view","settings:update","report:view"]');

-- Default settings
INSERT INTO settings (branch_id, key, value, description) VALUES
(1, 'exchange_rate_usd_khr', '4050', 'Default USD to KHR exchange rate'),
(1, 'tax_rate_percent', '0', 'Tax rate percentage'),
(1, 'tax_enabled', 'false', 'Enable tax calculation'),
(1, 'invoice_prefix', 'INV', 'Invoice number prefix'),
(1, 'low_stock_threshold', '5', 'Default low stock alert threshold'),
(1, 'telegram_enabled', 'false', 'Telegram bot enabled');

COMMIT;
