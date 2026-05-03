ALTER TABLE lots ADD COLUMN shipment_status TEXT NOT NULL DEFAULT 'in_progress';

CREATE TABLE sales (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id      UUID NOT NULL REFERENCES products(id),
    warehouse_id    UUID NOT NULL REFERENCES warehouses(id),
    buyer_name      TEXT NOT NULL,
    qty             NUMERIC(15,4) NOT NULL CHECK (qty > 0),
    sell_price      NUMERIC(15,4) NOT NULL CHECK (sell_price >= 0),
    payment_status  TEXT NOT NULL DEFAULT 'not_paid',
    shipment_status TEXT NOT NULL DEFAULT 'in_progress',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sales_product ON sales(product_id);
CREATE INDEX idx_sales_warehouse ON sales(warehouse_id);

CREATE TABLE sale_allocations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id     UUID NOT NULL REFERENCES sales(id),
    lot_id      UUID NOT NULL REFERENCES lots(id),
    qty         NUMERIC(15,4) NOT NULL CHECK (qty > 0),
    unit_cost   NUMERIC(15,4) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sale_alloc_sale ON sale_allocations(sale_id);
CREATE INDEX idx_sale_alloc_lot ON sale_allocations(lot_id);
