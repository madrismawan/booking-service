CREATE TABLE IF NOT EXISTS payment_transactions (
  id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT NOT NULL REFERENCES bookings(id),
  transaction_code VARCHAR(80),
  provider VARCHAR(80) NOT NULL,
  provider_transaction_id VARCHAR(120),
  provider_event_id VARCHAR(120),
  payment_method VARCHAR(80),
  status VARCHAR(50) NOT NULL,
  amount BIGINT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  paid_at TIMESTAMPTZ,
  expired_at TIMESTAMPTZ,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_booking_id ON payment_transactions (booking_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_provider_event_id
  ON payment_transactions (provider, provider_event_id)
  WHERE provider_event_id IS NOT NULL;
