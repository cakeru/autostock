# Database Schema

## Overview

AutoStock uses PostgreSQL as its primary database. The schema is designed to be:
- **Multi-branch ready**: All tenant-specific tables include `branch_id`
- **Audit-friendly**: Timestamps and soft deletes where appropriate
- **Normalized**: Proper foreign key relationships
- **Extensible**: JSONB fields for future flexibility

## Entity Relationship Diagram

```
┌─────────────┐
│   branches  │
└──────┬──────┘
       │ 1:N
       ▼
┌─────────────┐       ┌─────────────┐
│    users    │       │  settings   │
└──────┬──────┘       └─────────────┘
       │
       │ (creates)
       ▼
┌─────────────┐       ┌─────────────┐       ┌─────────────┐
│  customers  │◄──────│  vehicles   │       │   products  │
└──────┬──────┘       └─────────────┘       └──────┬──────┘
       │                                            │
       │                                            │
       ▼                                            ▼
┌─────────────┐                            ┌─────────────┐
│service_jobs │◄───────────────────────────│service_items│
└──────┬──────┘                            └─────────────┘
       │
       │ 1:1
       ▼
┌─────────────┐       ┌─────────────┐
│  invoices   │◄──────│invoice_items│
└──────┬──────┘       └─────────────┘
       │
       ▼
┌─────────────┐
│ audit_logs  │
└─────────────┘
```

## Table Definitions

### Core Tables

#### branches
Stores garage/branch information. Currently single-branch, but ready for multi-branch expansion.

```sql
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

-- Indexes
CREATE INDEX idx_branches_is_active ON branches(is_active);
```

#### users
System users (admin and staff).

```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'staff')),
    permissions JSONB DEFAULT '[]', -- Array of permission strings
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_branch_id ON users(branch_id);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_is_active ON users(is_active);
```

**Permissions Array Examples:**
```json
// Admin - all permissions
["inventory:view", "inventory:create", "inventory:update", "inventory:delete",
 "customer:view", "customer:create", "customer:update", "customer:delete",
 "service:view", "service:create", "service:update", "service:delete",
 "invoice:view", "invoice:create", "invoice:void",
 "user:view", "user:create", "user:update", "user:delete",
 "settings:view", "settings:update",
 "report:view"]

// Staff - limited permissions
["inventory:view", "inventory:create", "inventory:update",
 "customer:view", "customer:create", "customer:update",
 "service:view", "service:create", "service:update",
 "invoice:view", "invoice:create"]
```

#### settings
Key-value store for branch-specific and global settings.

```sql
CREATE TABLE settings (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT REFERENCES branches(id) ON DELETE CASCADE, -- NULL for global settings
    key VARCHAR(255) NOT NULL,
    value TEXT,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(branch_id, key)
);

-- Indexes
CREATE INDEX idx_settings_branch_id ON settings(branch_id);
CREATE INDEX idx_settings_key ON settings(key);
```

**Common Settings Keys:**
- `exchange_rate_usd_khr` - Default USD to KHR exchange rate (e.g., "4050")
- `tax_rate_percent` - Tax rate percentage (e.g., "10")
- `tax_enabled` - Whether tax calculation is enabled ("true"/"false")
- `invoice_prefix` - Invoice number prefix (e.g., "INV")
- `invoice_year_format` - Year format in invoice number (e.g., "2006")
- `telegram_enabled` - Whether Telegram bot is enabled
- `telegram_bot_token` - Telegram bot API token
- `telegram_chat_id` - Telegram chat/group ID for notifications
- `low_stock_threshold` - Default low stock alert threshold

### Inventory Tables

#### products
All inventory items (tires, parts, labor, consumables).

```sql
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    
    -- Basic info
    type VARCHAR(50) NOT NULL CHECK (type IN ('tire', 'part', 'labor', 'consumable')),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    
    -- Pricing
    buy_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    sell_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    
    -- Stock
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    min_stock_alert INTEGER DEFAULT 5, -- Alert when stock falls below this
    unit VARCHAR(50) DEFAULT 'piece', -- piece, liter, hour, etc.
    
    -- Tire-specific fields (nullable for non-tire products)
    tire_size VARCHAR(50), -- e.g., "205/55R16"
    tire_brand VARCHAR(100), -- e.g., "Michelin", "Bridgestone"
    tire_model VARCHAR(100), -- e.g., "Primacy 4"
    tire_pattern VARCHAR(100), -- Tread pattern
    dot_code VARCHAR(50), -- Manufacturing date code
    load_index VARCHAR(10), -- e.g., "91"
    speed_rating VARCHAR(5), -- e.g., "V", "H", "W"
    tire_type VARCHAR(50), -- passenger, truck, suv, motorcycle
    
    -- Location
    location VARCHAR(100), -- Shelf/rack location in warehouse
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_products_branch_id ON products(branch_id);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_tire_size ON products(tire_size) WHERE type = 'tire';
CREATE INDEX idx_products_tire_brand ON products(tire_brand) WHERE type = 'tire';
CREATE INDEX idx_products_stock_quantity ON products(stock_quantity);
CREATE INDEX idx_products_is_active ON products(is_active);

-- Composite index for common queries
CREATE INDEX idx_products_branch_type_active ON products(branch_id, type, is_active);
```

**Tire Size Format:**
Standard format: `WIDTH/ASPECT_RATIOR_DIAMETER`
- Example: `205/55R16`
  - 205 = Width in mm
  - 55 = Aspect ratio (sidewall height as % of width)
  - R = Radial construction
  - 16 = Rim diameter in inches

### Customer Tables

#### customers
Customer profiles.

```sql
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    
    -- Contact info
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    
    -- Additional info
    notes TEXT,
    customer_since DATE,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_customers_branch_id ON customers(branch_id);
CREATE INDEX idx_customers_name ON customers(name);
CREATE INDEX idx_customers_phone ON customers(phone);
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_is_active ON customers(is_active);
```

#### vehicles
Customer vehicles.

```sql
CREATE TABLE vehicles (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    
    -- Vehicle info
    plate_number VARCHAR(50) NOT NULL,
    make VARCHAR(100), -- e.g., "Toyota", "Honda"
    model VARCHAR(100), -- e.g., "Camry", "Civic"
    year INTEGER CHECK (year >= 1900 AND year <= 2100),
    vin VARCHAR(17), -- Vehicle Identification Number
    color VARCHAR(50),
    
    -- Additional info
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(customer_id, plate_number)
);

-- Indexes
CREATE INDEX idx_vehicles_customer_id ON vehicles(customer_id);
CREATE INDEX idx_vehicles_plate_number ON vehicles(plate_number);
CREATE INDEX idx_vehicles_make_model ON vehicles(make, model);
```

### Service & Invoice Tables

#### service_jobs
Service/repair jobs.

```sql
CREATE TABLE service_jobs (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    
    -- References
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE SET NULL,
    invoice_id BIGINT, -- Set when job is invoiced
    
    -- Job info
    job_number VARCHAR(50) NOT NULL, -- e.g., "JOB-2026-0001"
    status VARCHAR(50) NOT NULL DEFAULT 'pending' 
        CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    priority VARCHAR(50) DEFAULT 'normal' 
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    
    -- Description
    description TEXT, -- Customer's complaint/request
    diagnosis TEXT, -- Mechanic's findings
    work_performed TEXT, -- What was actually done
    
    -- Timing
    estimated_hours DECIMAL(5, 2),
    actual_hours DECIMAL(5, 2),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Notes
    notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_service_jobs_branch_id ON service_jobs(branch_id);
CREATE INDEX idx_service_jobs_customer_id ON service_jobs(customer_id);
CREATE INDEX idx_service_jobs_vehicle_id ON service_jobs(vehicle_id);
CREATE INDEX idx_service_jobs_invoice_id ON service_jobs(invoice_id);
CREATE INDEX idx_service_jobs_status ON service_jobs(status);
CREATE INDEX idx_service_jobs_job_number ON service_jobs(job_number);
CREATE INDEX idx_service_jobs_created_at ON service_jobs(created_at);
```

#### service_job_items
Products/labor used in a service job.

```sql
CREATE TABLE service_job_items (
    id BIGSERIAL PRIMARY KEY,
    service_job_id BIGINT NOT NULL REFERENCES service_jobs(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    
    -- Item details
    description VARCHAR(255), -- Override product name if needed
    quantity DECIMAL(10, 2) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CHECK (quantity > 0),
    CHECK (unit_price >= 0),
    CHECK (total_price >= 0)
);

-- Indexes
CREATE INDEX idx_service_job_items_service_job_id ON service_job_items(service_job_id);
CREATE INDEX idx_service_job_items_product_id ON service_job_items(product_id);
```

#### invoices
Customer invoices.

```sql
CREATE TABLE invoices (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    
    -- Invoice identification
    invoice_number VARCHAR(50) NOT NULL UNIQUE, -- e.g., "INV-2026-0001"
    
    -- References
    customer_id BIGINT REFERENCES customers(id) ON DELETE SET NULL,
    vehicle_id BIGINT REFERENCES vehicles(id) ON DELETE SET NULL,
    service_job_id BIGINT REFERENCES service_jobs(id) ON DELETE SET NULL,
    
    -- Financial details
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5, 2) DEFAULT 0, -- Percentage, e.g., 10.00 for 10%
    tax_amount DECIMAL(10, 2) DEFAULT 0,
    discount DECIMAL(10, 2) DEFAULT 0,
    
    -- Currency
    total_usd DECIMAL(10, 2) NOT NULL DEFAULT 0,
    exchange_rate DECIMAL(10, 2) NOT NULL, -- USD to KHR rate used
    total_khr DECIMAL(10, 2) NOT NULL DEFAULT 0, -- Calculated: total_usd * exchange_rate
    
    -- Status
    payment_status VARCHAR(50) NOT NULL DEFAULT 'unpaid' 
        CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'refunded', 'voided')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' 
        CHECK (status IN ('draft', 'issued', 'paid', 'voided')),
    
    -- Payment info
    paid_amount DECIMAL(10, 2) DEFAULT 0,
    payment_method VARCHAR(50), -- cash, card, transfer, etc.
    payment_notes TEXT,
    
    -- Additional info
    notes TEXT,
    terms TEXT,
    
    -- Void info
    voided_at TIMESTAMP WITH TIME ZONE,
    void_reason TEXT,
    voided_by BIGINT REFERENCES users(id),
    
    -- Timestamps
    issued_at TIMESTAMP WITH TIME ZONE,
    due_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_invoices_branch_id ON invoices(branch_id);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_customer_id ON invoices(customer_id);
CREATE INDEX idx_invoices_vehicle_id ON invoices(vehicle_id);
CREATE INDEX idx_invoices_service_job_id ON invoices(service_job_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_payment_status ON invoices(payment_status);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);
CREATE INDEX idx_invoices_issued_at ON invoices(issued_at);
```

#### invoice_items
Line items in an invoice.

```sql
CREATE TABLE invoice_items (
    id BIGSERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE SET NULL, -- NULL for custom items
    
    -- Item details
    item_type VARCHAR(50) NOT NULL CHECK (item_type IN ('product', 'labor', 'custom')),
    description VARCHAR(500) NOT NULL,
    
    -- Quantity and pricing
    quantity DECIMAL(10, 2) NOT NULL DEFAULT 1,
    unit_price_usd DECIMAL(10, 2) NOT NULL,
    total_usd DECIMAL(10, 2) NOT NULL,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CHECK (quantity > 0),
    CHECK (unit_price_usd >= 0),
    CHECK (total_usd >= 0)
);

-- Indexes
CREATE INDEX idx_invoice_items_invoice_id ON invoice_items(invoice_id);
CREATE INDEX idx_invoice_items_product_id ON invoice_items(product_id);
```

### Audit & Activity Tables

#### audit_logs
Track all significant actions for audit trail.

```sql
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    branch_id BIGINT NOT NULL REFERENCES branches(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    
    -- Action details
    action VARCHAR(100) NOT NULL, -- create, update, delete, login, etc.
    entity_type VARCHAR(100) NOT NULL, -- user, product, customer, invoice, etc.
    entity_id BIGINT, -- ID of the affected entity
    
    -- Change details
    old_values JSONB, -- Previous values (for updates)
    new_values JSONB, -- New values
    changes JSONB, -- Diff of changes
    
    -- Context
    ip_address INET,
    user_agent TEXT,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_branch_id ON audit_logs(branch_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Composite index for common queries
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id, created_at DESC);
```

## Data Types & Conventions

### Primary Keys
- Use `BIGSERIAL` for all primary keys
- Never reuse or expose internal IDs in URLs (use UUIDs or hashed IDs if needed)

### Foreign Keys
- Always define foreign key constraints
- Use `ON DELETE CASCADE` for dependent records (e.g., invoice_items when invoice deleted)
- Use `ON DELETE SET NULL` for optional relationships
- Use `ON DELETE RESTRICT` to prevent deletion of referenced records

### Timestamps
- Use `TIMESTAMP WITH TIME ZONE` for all timestamps
- Store UTC, convert to local timezone on display
- Always include `created_at` and `updated_at`
- Use database triggers to auto-update `updated_at`

### Soft Deletes
- Use `is_active` boolean flag instead of hard deletes for important entities
- Hard delete only for truly temporary data (sessions, logs after retention period)

### JSONB Fields
- Use for flexible, schema-less data
- Examples: `permissions`, `old_values`, `new_values`
- Index specific JSONB paths if queried frequently

### Currency Storage
- Store all monetary values in USD as `DECIMAL(10, 2)`
- Store exchange rate with invoice for historical accuracy
- Calculate KHR amounts on-the-fly: `total_khr = total_usd * exchange_rate`

## Migrations

Database migrations are managed using **golang-migrate** (Go) or **Alembic** (Python alternative).

### Migration File Naming
```
migrations/
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_tire_fields.up.sql
├── 000002_add_tire_fields.down.sql
└── ...
```

### Running Migrations

```bash
# Using Go migration tool
migrate -path ./migrations -database "postgres://user:pass@localhost:5432/autostock?sslmode=disable" up

# Or using Docker
docker-compose exec backend migrate -path ./migrations -database "$DATABASE_URL" up
```

## Indexes Strategy

### When to Add Indexes
1. Foreign key columns (automatic in some databases)
2. Columns used in WHERE clauses frequently
3. Columns used in ORDER BY
4. Columns used in JOIN conditions
5. Composite indexes for multi-column queries

### Index Types
- **B-tree** (default): For equality and range queries
- **Hash**: For equality only (rarely needed)
- **GIN**: For JSONB, array, and full-text search
- **Partial indexes**: Index only a subset of rows (e.g., `WHERE is_active = true`)

### Monitoring
- Use `pg_stat_user_indexes` to monitor index usage
- Remove unused indexes to improve write performance
- Use `EXPLAIN ANALYZE` to verify query plans

## Backup Strategy

### Automated Backups
```bash
# Daily backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U autostock -h localhost autostock | gzip > "$BACKUP_DIR/autostock_$DATE.sql.gz"

# Keep last 30 days
find $BACKUP_DIR -name "autostock_*.sql.gz" -mtime +30 -delete
```

### Restore
```bash
# Restore from backup
gunzip -c autostock_20260704_120000.sql.gz | psql -U autostock -h localhost autostock
```

## Performance Optimization

### Connection Pooling
Use **pgxpool** in Go for connection pooling:
```go
poolConfig, err := pgxpool.ParseConfig(databaseURL)
poolConfig.MaxConns = 20
poolConfig.MinConns = 5
pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
```

### Query Optimization
1. Use `EXPLAIN ANALYZE` to identify slow queries
2. Avoid `SELECT *` - specify needed columns
3. Use `LIMIT` for large result sets
4. Batch inserts when possible
5. Use prepared statements for repeated queries

### Partitioning (Future)
For very large tables (e.g., `audit_logs`), consider partitioning by date:
```sql
CREATE TABLE audit_logs (
    -- columns
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_2026_01 PARTITION OF audit_logs
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');
```

## Data Integrity

### Constraints
- **NOT NULL**: Required fields
- **UNIQUE**: Prevent duplicates (e.g., username, email, invoice_number)
- **CHECK**: Validate data (e.g., `quantity > 0`, `role IN ('admin', 'staff')`)
- **FOREIGN KEY**: Maintain referential integrity

### Transactions
Use transactions for multi-step operations:
```go
tx, err := db.Begin()
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// Multiple operations
err = repo1.Create(ctx, tx, entity1)
err = repo2.Create(ctx, tx, entity2)

err = tx.Commit()
```

### Validation
- Validate at application layer before database
- Use database constraints as safety net
- Return clear error messages for validation failures

## Sample Data

### Initial Seed Data
```sql
-- Default branch
INSERT INTO branches (name, address, phone) VALUES 
('AutoStock Garage', 'Phnom Penh, Cambodia', '+855 12 345 678');

-- Admin user (password: admin123)
INSERT INTO users (branch_id, username, password_hash, full_name, role, permissions) VALUES 
(1, 'admin', '$2a$12$...', 'Administrator', 'admin', 
 '["inventory:view","inventory:create","inventory:update","inventory:delete","customer:view","customer:create","customer:update","customer:delete","service:view","service:create","service:update","service:delete","invoice:view","invoice:create","invoice:void","user:view","user:create","user:update","user:delete","settings:view","settings:update","report:view"]');

-- Default settings
INSERT INTO settings (branch_id, key, value, description) VALUES 
(1, 'exchange_rate_usd_khr', '4050', 'Default USD to KHR exchange rate'),
(1, 'tax_rate_percent', '0', 'Tax rate percentage'),
(1, 'tax_enabled', 'false', 'Enable tax calculation'),
(1, 'invoice_prefix', 'INV', 'Invoice number prefix'),
(1, 'low_stock_threshold', '5', 'Default low stock alert threshold');
```

## Future Schema Changes

### Phase 2 Additions
- **Multi-branch**: Already supported via `branch_id`
- **Tax module**: Add `tax_id` to branches, tax calculations to invoices
- **Loyalty program**: Add `loyalty_points` to customers, `loyalty_transactions` table
- **Payment integration**: Add `payment_gateway` table, `transaction_id` to invoices

### Phase 3 Additions
- **Consumables tracking**: Add `batch_number`, `expiry_date` to products
- **Advanced reporting**: Add `report_templates`, `scheduled_reports` tables
- **API integrations**: Add `api_keys`, `webhooks` tables
