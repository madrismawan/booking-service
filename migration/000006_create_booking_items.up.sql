CREATE TABLE IF NOT EXISTS booking_items (
  id BIGSERIAL PRIMARY KEY,
  booking_id BIGINT NOT NULL REFERENCES bookings(id),
  ticket_category_id BIGINT NOT NULL REFERENCES ticket_categories(id),
  ticket_category_name VARCHAR(100) NOT NULL,
  ticket_category_description TEXT,
  quantity INTEGER NOT NULL,
  unit_price BIGINT NOT NULL,
  subtotal_price BIGINT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_booking_items_booking_id ON booking_items (booking_id);
CREATE INDEX IF NOT EXISTS idx_booking_items_ticket_category_id ON booking_items (ticket_category_id);
