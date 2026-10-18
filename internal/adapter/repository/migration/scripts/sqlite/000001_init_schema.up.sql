CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    unit_price REAL NOT NULL,
    min_threshold INTEGER NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(tenant_id, sku)
);

CREATE INDEX IF NOT EXISTS idx_items_tenant_id ON items(tenant_id);

CREATE TABLE IF NOT EXISTS stock_levels (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    warehouse_id TEXT NOT NULL DEFAULT 'default',
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL,
    UNIQUE(tenant_id, item_id, warehouse_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_levels_tenant ON stock_levels(tenant_id, item_id, warehouse_id);

CREATE TABLE IF NOT EXISTS stock_movements (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    warehouse_id TEXT NOT NULL DEFAULT 'default',
    movement_type TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    balance_after INTEGER NOT NULL,
    total_cogs REAL NOT NULL DEFAULT 0.0,
    reference TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_tenant ON stock_movements(tenant_id, item_id, warehouse_id);

CREATE TABLE IF NOT EXISTS stock_cost_layers (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    warehouse_id TEXT NOT NULL DEFAULT 'default',
    quantity INTEGER NOT NULL,
    remaining_quantity INTEGER NOT NULL,
    unit_cost REAL NOT NULL,
    created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cost_layers_lookup ON stock_cost_layers(tenant_id, item_id, warehouse_id, remaining_quantity);

CREATE TABLE IF NOT EXISTS tenant_settings (
    tenant_id TEXT PRIMARY KEY,
    valuation_method TEXT NOT NULL DEFAULT 'FIFO',
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS suppliers (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    contact_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TEXT NOT NULL,
    UNIQUE(tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_suppliers_tenant ON suppliers(tenant_id);

CREATE TABLE IF NOT EXISTS purchase_orders (
    id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    po_number TEXT NOT NULL,
    supplier_id TEXT NOT NULL,
    warehouse_id TEXT NOT NULL DEFAULT 'default',
    status TEXT NOT NULL DEFAULT 'DRAFT',
    notes TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(tenant_id, po_number)
);

CREATE INDEX IF NOT EXISTS idx_po_tenant ON purchase_orders(tenant_id, po_number);

CREATE TABLE IF NOT EXISTS purchase_order_items (
    id TEXT PRIMARY KEY,
    po_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    item_id TEXT NOT NULL,
    sku TEXT NOT NULL,
    ordered_quantity INTEGER NOT NULL,
    received_quantity INTEGER NOT NULL DEFAULT 0,
    unit_cost REAL NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_po_items_po ON purchase_order_items(po_id);
