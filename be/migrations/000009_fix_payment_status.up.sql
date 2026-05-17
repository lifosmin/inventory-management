-- Set default to 'unpaid' and update existing rows that were created with the old default
ALTER TABLE sales ALTER COLUMN payment_status SET DEFAULT 'unpaid';

UPDATE sales
SET payment_status = 'unpaid'
WHERE payment_status = 'not_paid' OR payment_status IS NULL;
