-- Revert default to 'not_paid' and (optionally) revert rows that were set to 'unpaid'
ALTER TABLE sales ALTER COLUMN payment_status SET DEFAULT 'not_paid';

-- WARNING: the following will set any 'unpaid' values back to 'not_paid'.
-- Only run if you are sure this won't overwrite legitimate 'unpaid' semantics.
UPDATE sales
SET payment_status = 'not_paid'
WHERE payment_status = 'unpaid';
