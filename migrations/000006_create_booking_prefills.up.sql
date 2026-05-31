CREATE TABLE booking_prefills (
    id BIGSERIAL PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    name TEXT,
    email TEXT,
    phone TEXT,
    pickup_at TEXT,
    return_at TEXT,
    message TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_prefills_expires_at
    ON booking_prefills (expires_at);
