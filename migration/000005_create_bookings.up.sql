CREATE TABLE IF NOT EXISTS bookings (
  id BIGSERIAL PRIMARY KEY,
  booking_code VARCHAR(40) NOT NULL UNIQUE,
  guest_id BIGINT NOT NULL REFERENCES guests(id),
  guest_name VARCHAR(255) NOT NULL,
  guest_email VARCHAR(255) NOT NULL,
  guest_phone VARCHAR(20) NOT NULL,
  guest_address TEXT NOT NULL,
  event_id BIGINT NOT NULL REFERENCES events(id),
  event_slug VARCHAR(160) NOT NULL,
  event_name VARCHAR(200) NOT NULL,
  event_venue_name VARCHAR(200) NOT NULL,
  event_venue_address TEXT NOT NULL,
  event_starts_at TIMESTAMPTZ NOT NULL,
  event_ends_at TIMESTAMPTZ,
  status VARCHAR(50) NOT NULL DEFAULT 'pending_payment',
  total_ticket INTEGER NOT NULL,
  total_price BIGINT NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  paid_at TIMESTAMPTZ,
  cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bookings_guest_id ON bookings (guest_id);
CREATE INDEX IF NOT EXISTS idx_bookings_event_id ON bookings (event_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings (status);
