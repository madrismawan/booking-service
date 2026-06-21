CREATE TABLE IF NOT EXISTS ticket_categories (
  id BIGSERIAL PRIMARY KEY,
  event_id BIGINT NOT NULL REFERENCES events(id),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  price BIGINT NOT NULL,
  sale_starts_at TIMESTAMPTZ,
  sale_ends_at TIMESTAMPTZ,
  max_per_booking INTEGER NOT NULL DEFAULT 4,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_categories_event_id ON ticket_categories (event_id);
