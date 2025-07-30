CREATE TABLE IF NOT EXISTS payments (
    correlation_id UUID PRIMARY KEY,
    amount DECIMAL NOT NULL,
    processed_by SMALLINT NOT NULL,
    requested_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_payments_requested_at_processed_by ON payments(requested_at, processed_by);