CREATE TABLE IF NOT EXISTS waiting_rooms (
  id BIGSERIAL PRIMARY KEY,
  event_id BIGINT NOT NULL REFERENCES events(id),
  event_name VARCHAR(200) NOT NULL,
  ticket_category_id BIGINT NOT NULL REFERENCES ticket_categories(id),
  queue_token VARCHAR(80) NOT NULL UNIQUE,
  checkout_token VARCHAR(80) UNIQUE,
  booking_id BIGINT REFERENCES bookings(id),
  status VARCHAR(50) NOT NULL DEFAULT 'waiting',
  failed_reason TEXT,
  expired_at TIMESTAMPTZ,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_waiting_rooms_event_name ON waiting_rooms (event_name);
CREATE INDEX IF NOT EXISTS idx_waiting_rooms_ticket_category_id ON waiting_rooms (ticket_category_id);
CREATE INDEX IF NOT EXISTS idx_waiting_rooms_status ON waiting_rooms (status);
CREATE INDEX IF NOT EXISTS idx_waiting_rooms_queue_token ON waiting_rooms (queue_token);
CREATE INDEX IF NOT EXISTS idx_waiting_rooms_checkout_token ON waiting_rooms (checkout_token);
