CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'viewer',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku         TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    unit        TEXT NOT NULL,
    category    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT UNIQUE NOT NULL,
    address     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE lots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_number      TEXT UNIQUE NOT NULL,
    product_id      UUID NOT NULL REFERENCES products(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    quantity         NUMERIC(15,4) NOT NULL CHECK (quantity >= 0),
    unit_cost        NUMERIC(15,4) NOT NULL CHECK (unit_cost >= 0),
    total_cost       NUMERIC(15,4) GENERATED ALWAYS AS (quantity * unit_cost) STORED,
    received_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expiry_date      DATE,
    status           TEXT NOT NULL DEFAULT 'available',
    supplier         TEXT,
    reference_doc    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lots_product ON lots(product_id);
CREATE INDEX idx_lots_warehouse ON lots(warehouse_id);
CREATE INDEX idx_lots_status ON lots(status);
CREATE INDEX idx_lots_received ON lots(received_at);

CREATE TABLE lot_movements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id          UUID NOT NULL REFERENCES lots(id),
    movement_type   TEXT NOT NULL,
    quantity         NUMERIC(15,4) NOT NULL,
    reference_doc    TEXT,
    notes            TEXT,
    performed_by     UUID REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_movements_lot ON lot_movements(lot_id);
CREATE INDEX idx_movements_type ON lot_movements(movement_type);
CREATE INDEX idx_movements_created ON lot_movements(created_at);

CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name  TEXT NOT NULL,
    record_id   UUID NOT NULL,
    action      TEXT NOT NULL,
    old_data    JSONB,
    new_data    JSONB,
    performed_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_record ON audit_log(table_name, record_id);
