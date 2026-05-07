ALTER TABLE lots
    DROP COLUMN IF EXISTS paid_amount,
    DROP COLUMN IF EXISTS payment_status;

ALTER TABLE sales
    DROP COLUMN IF EXISTS paid_amount;
