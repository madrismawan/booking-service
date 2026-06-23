CREATE TABLE IF NOT EXISTS ticket_stocks (
  id BIGSERIAL PRIMARY KEY,
  ticket_category_id BIGINT NOT NULL UNIQUE REFERENCES ticket_categories(id),
  total_quantity INTEGER NOT NULL,
  available_quantity INTEGER NOT NULL,
  reserved_quantity INTEGER NOT NULL DEFAULT 0,
  sold_quantity INTEGER NOT NULL DEFAULT 0,
  version BIGINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT ticket_stocks_quantity_check CHECK (
    total_quantity >= 0
    AND available_quantity >= 0
    AND reserved_quantity >= 0
    AND sold_quantity >= 0
    AND available_quantity + reserved_quantity + sold_quantity = total_quantity
  )
);
