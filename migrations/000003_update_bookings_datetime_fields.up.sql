ALTER TABLE bookings
    DROP CONSTRAINT IF EXISTS bookings_date_range_check;

DROP INDEX IF EXISTS idx_bookings_start_date;
DROP INDEX IF EXISTS idx_bookings_end_date;

ALTER TABLE bookings
    ADD COLUMN pickup_at TIMESTAMPTZ,
    ADD COLUMN return_at TIMESTAMPTZ,
    ADD COLUMN billing_days INTEGER,
    ADD COLUMN estimated_total NUMERIC(10, 2);

UPDATE bookings b
SET
    pickup_at = b.start_date::timestamptz,
    return_at = b.end_date::timestamptz + INTERVAL '1 day',
    billing_days = GREATEST(
        1,
        CEIL(
            EXTRACT(EPOCH FROM ((b.end_date::timestamptz + INTERVAL '1 day') - b.start_date::timestamptz)) / 86400
        )::integer
    ),
    estimated_total = c.price_per_day * GREATEST(
        1,
        CEIL(
            EXTRACT(EPOCH FROM ((b.end_date::timestamptz + INTERVAL '1 day') - b.start_date::timestamptz)) / 86400
        )::integer
    )
FROM cars c
WHERE c.id = b.car_id;

ALTER TABLE bookings
    ALTER COLUMN pickup_at SET NOT NULL,
    ALTER COLUMN return_at SET NOT NULL,
    ALTER COLUMN billing_days SET NOT NULL,
    ALTER COLUMN estimated_total SET NOT NULL,
    DROP COLUMN start_date,
    DROP COLUMN end_date,
    ADD CONSTRAINT bookings_datetime_range_check CHECK (return_at > pickup_at),
    ADD CONSTRAINT bookings_billing_days_check CHECK (billing_days >= 1),
    ADD CONSTRAINT bookings_estimated_total_check CHECK (estimated_total >= 0);
