CREATE TABLE items (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    sku VARCHAR2(64) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    description CLOB,
    unit_price NUMBER(15, 2) NOT NULL,
    min_threshold NUMBER(10) DEFAULT 0 NOT NULL,
    is_active NUMBER(1) DEFAULT 1 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT uq_items_tenant_sku UNIQUE(tenant_id, sku)
);

CREATE INDEX idx_items_tenant_id ON items(tenant_id);

CREATE TABLE stock_levels (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    item_id VARCHAR2(64) NOT NULL,
    warehouse_id VARCHAR2(64) DEFAULT 'default' NOT NULL,
    quantity NUMBER(10) DEFAULT 0 NOT NULL,
    reserved_quantity NUMBER(10) DEFAULT 0 NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT uq_stock_levels UNIQUE(tenant_id, item_id, warehouse_id)
);

CREATE INDEX idx_stock_levels_tenant ON stock_levels(tenant_id, item_id, warehouse_id);

CREATE TABLE stock_movements (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    item_id VARCHAR2(64) NOT NULL,
    warehouse_id VARCHAR2(64) DEFAULT 'default' NOT NULL,
    movement_type VARCHAR2(32) NOT NULL,
    quantity NUMBER(10) NOT NULL,
    balance_after NUMBER(10) NOT NULL,
    total_cogs NUMBER(15, 2) DEFAULT 0.0 NOT NULL,
    reference VARCHAR2(255),
    notes CLOB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_stock_movements_tenant ON stock_movements(tenant_id, item_id, warehouse_id);

CREATE TABLE stock_cost_layers (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    item_id VARCHAR2(64) NOT NULL,
    warehouse_id VARCHAR2(64) DEFAULT 'default' NOT NULL,
    quantity NUMBER(10) NOT NULL,
    remaining_quantity NUMBER(10) NOT NULL,
    unit_cost NUMBER(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_cost_layers_lookup ON stock_cost_layers(tenant_id, item_id, warehouse_id, remaining_quantity);

CREATE TABLE tenant_settings (
    tenant_id VARCHAR2(64) PRIMARY KEY,
    valuation_method VARCHAR2(32) DEFAULT 'FIFO' NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE suppliers (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    name VARCHAR2(255) NOT NULL,
    contact_name VARCHAR2(255),
    email VARCHAR2(255) NOT NULL,
    phone VARCHAR2(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT uq_suppliers_tenant_name UNIQUE(tenant_id, name)
);

CREATE INDEX idx_suppliers_tenant ON suppliers(tenant_id);

CREATE TABLE purchase_orders (
    id VARCHAR2(64) PRIMARY KEY,
    tenant_id VARCHAR2(64) NOT NULL,
    po_number VARCHAR2(64) NOT NULL,
    supplier_id VARCHAR2(64) NOT NULL,
    warehouse_id VARCHAR2(64) DEFAULT 'default' NOT NULL,
    status VARCHAR2(32) DEFAULT 'DRAFT' NOT NULL,
    notes CLOB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT uq_po_tenant_po UNIQUE(tenant_id, po_number)
);

CREATE INDEX idx_po_tenant ON purchase_orders(tenant_id, po_number);

CREATE TABLE purchase_order_items (
    id VARCHAR2(64) PRIMARY KEY,
    po_id VARCHAR2(64) NOT NULL,
    tenant_id VARCHAR2(64) NOT NULL,
    item_id VARCHAR2(64) NOT NULL,
    sku VARCHAR2(64) NOT NULL,
    ordered_quantity NUMBER(10) NOT NULL,
    received_quantity NUMBER(10) DEFAULT 0 NOT NULL,
    unit_cost NUMBER(15, 2) NOT NULL
);

CREATE INDEX idx_po_items_po ON purchase_order_items(po_id);
