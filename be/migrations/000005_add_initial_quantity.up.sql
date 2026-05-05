ALTER TABLE lots ADD COLUMN initial_quantity NUMERIC(15,4);
UPDATE lots SET initial_quantity = quantity + COALESCE((SELECT SUM(sa.qty) FROM sale_allocations sa WHERE sa.lot_id = lots.id), 0);
ALTER TABLE lots ALTER COLUMN initial_quantity SET NOT NULL;
ALTER TABLE lots ALTER COLUMN initial_quantity SET DEFAULT 0;
