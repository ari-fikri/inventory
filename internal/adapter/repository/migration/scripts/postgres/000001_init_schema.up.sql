CREATE TABLE IF NOT EXISTS items (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    sku VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    unit_price NUMERIC(15, 2) NOT NULL,
    min_threshold INTEGER NOT NULL DEFAULT 0,
    is_active SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_items_tenant_sku UNIQUE(tenant_id, sku)
);

CREATE INDEX IF NOT EXISTS idx_items_tenant_id ON items(tenant_id);

CREATE TABLE IF NOT EXISTS stock_levels (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    warehouse_id VARCHAR(64) NOT NULL DEFAULT 'default',
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_stock_levels UNIQUE(tenant_id, item_id, warehouse_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_levels_tenant ON stock_levels(tenant_id, item_id, warehouse_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    warehouse_id VARCHAR(64) NOT NULL DEFAULT 'default',
    movement_type VARCHAR(32) NOT NULL,
    quantity INTEGER NOT NULL,
    balance_after INTEGER NOT NULL,
    total_cogs NUMERIC(15, 2) NOT NULL DEFAULT 0.0,
    reference VARCHAR(255) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_tenant ON stock_movements(tenant_id, item_id, warehouse_id);

CREATE TABLE IF NOT EXISTS stock_cost_layers (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    warehouse_id VARCHAR(64) NOT NULL DEFAULT 'default',
    quantity INTEGER NOT NULL,
    remaining_quantity INTEGER NOT NULL,
    unit_cost NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cost_layers_lookup ON stock_cost_layers(tenant_id, item_id, warehouse_id, remaining_quantity);

CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id VARCHAR(64) PRIMARY KEY,
    valuation_method VARCHAR(32) NOT NULL DEFAULT 'FIFO',
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS suppliers (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_suppliers_tenant_name UNIQUE(tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_suppliers_tenant ON suppliers(tenant_id);

CREATE TABLE IF NOT EXISTS purchase_orders (
    id VARCHAR(64) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    po_number VARCHAR(64) NOT NULL,
    supplier_id VARCHAR(64) NOT NULL,
    warehouse_id VARCHAR(64) NOT NULL DEFAULT 'default',
    status VARCHAR(32) NOT NULL DEFAULT 'DRAFT',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_purchase_orders_tenant_po UNIQUE(tenant_id, po_number)
);

CREATE INDEX IF NOT EXISTS idx_po_tenant ON purchase_orders(tenant_id, po_number);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id VARCHAR(64) PRIMARY KEY,
    po_id VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    item_id VARCHAR(64) NOT NULL,
    sku VARCHAR(64) NOT NULL,
    ordered_quantity INTEGER NOT NULL,
    received_quantity INTEGER NOT NULL DEFAULT 0,
    unit_cost NUMERIC(15, 2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_po_items_po ON purchase_order_items(po_id);
