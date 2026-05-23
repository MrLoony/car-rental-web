ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS bookings_datetime_range_check,
    DROP CONSTRAINT IF EXISTS bookings_billing_days_check,
    DROP CONSTRAINT IF EXISTS bookings_estimated_total_check;

ALTER TABLE bookings
    ADD COLUMN start_date DATE,
    ADD COLUMN end_date DATE;

UPDATE bookings
SET
    start_date = pickup_at::date,
    end_date = return_at::date;

ALTER TABLE bookings
    ALTER COLUMN start_date SET NOT NULL,
    ALTER COLUMN end_date SET NOT NULL,
    DROP COLUMN pickup_at,
    DROP COLUMN return_at,
    DROP COLUMN billing_days,
    DROP COLUMN estimated_total,
    ADD CONSTRAINT bookings_date_range_check CHECK (end_date >= start_date);

CREATE INDEX idx_bookings_start_date ON bookings (start_date);
CREATE INDEX idx_bookings_end_date ON bookings (end_date);
