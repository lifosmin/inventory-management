ALTER TABLE lots
    ADD COLUMN paid_amount    NUMERIC(15,4) NOT NULL DEFAULT 0,
    ADD COLUMN payment_status TEXT          NOT NULL DEFAULT 'unpaid';

ALTER TABLE sales
    ADD COLUMN paid_amount NUMERIC(15,4) NOT NULL DEFAULT 0;
