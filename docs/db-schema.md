# Database Schema - Ticket Booking System

Dokumen ini mengikuti migration aktif pada folder `migration/`.

---

### 01. **events**

**Purpose**: Menyimpan data konser atau acara.

| Column          | Data Type      | Constraint / Default                  | Description            |
| --------------- | -------------- | ------------------------------------- | ---------------------- |
| `id`            | `BIGSERIAL`    | Primary key                           | ID event.              |
| `slug`          | `VARCHAR(160)` | Not null, unique                      | Slug unik event.       |
| `name`          | `VARCHAR(200)` | Not null                              | Nama event.            |
| `description`   | `TEXT`         | Nullable                              | Deskripsi event.       |
| `venue_name`    | `VARCHAR(200)` | Not null                              | Nama venue.            |
| `venue_address` | `TEXT`         | Not null                              | Alamat venue.          |
| `starts_at`     | `TIMESTAMPTZ`  | Not null                              | Waktu mulai event.     |
| `ends_at`       | `TIMESTAMPTZ`  | Nullable                              | Waktu selesai event.   |
| `status`        | `VARCHAR(50)`  | Not null, default `draft`             | Status event.          |
| `created_at`    | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.     |
| `updated_at`    | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui. |
| `deleted_at`    | `TIMESTAMP`    | Nullable                              | Waktu soft delete.     |

Indexes:

- Unique `idx_events_slug` pada `slug`.
- `idx_events_starts_at` pada `starts_at`.
- `idx_events_status` pada `status`.

---

### 02. **ticket_categories**

**Purpose**: Menyimpan kategori tiket untuk setiap event.

| Column            | Data Type      | Constraint / Default                  | Description                        |
| ----------------- | -------------- | ------------------------------------- | ---------------------------------- |
| `id`              | `BIGSERIAL`    | Primary key                           | ID kategori tiket.                 |
| `event_id`        | `BIGINT`       | Not null, FK ke `events.id`           | Event pemilik kategori.            |
| `name`            | `VARCHAR(100)` | Not null                              | Nama kategori tiket.               |
| `description`     | `TEXT`         | Nullable                              | Deskripsi kategori.                |
| `price`           | `BIGINT`       | Not null                              | Harga satu tiket.                  |
| `sale_starts_at`  | `TIMESTAMPTZ`  | Nullable                              | Waktu mulai penjualan.             |
| `sale_ends_at`    | `TIMESTAMPTZ`  | Nullable                              | Waktu selesai penjualan.           |
| `max_per_booking` | `INTEGER`      | Not null, default `4`                 | Maksimal tiket dalam satu booking. |
| `created_at`      | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.                 |
| `updated_at`      | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui.             |
| `deleted_at`      | `TIMESTAMP`    | Nullable                              | Waktu soft delete.                 |

Indexes:

- `idx_ticket_categories_event_id` pada `event_id`.

---

### 03. **ticket_stocks**

**Purpose**: Menyimpan stok untuk setiap kategori tiket.

| Column               | Data Type   | Constraint / Default                  | Description                         |
| -------------------- | ----------- | ------------------------------------- | ----------------------------------- |
| `id`                 | `BIGSERIAL` | Primary key                           | ID stok.                            |
| `ticket_category_id` | `BIGINT`    | Not null, unique, FK                  | Referensi ke `ticket_categories.id` |
| `total_quantity`     | `INTEGER`   | Not null                              | Total stok tiket.                   |
| `available_quantity` | `INTEGER`   | Not null                              | Stok yang masih tersedia.           |
| `reserved_quantity`  | `INTEGER`   | Not null, default `0`                 | Stok yang sedang direservasi.       |
| `sold_quantity`      | `INTEGER`   | Not null, default `0`                 | Stok yang sudah terjual.            |
| `version`            | `BIGINT`    | Not null, default `1`                 | Versi perubahan stok.               |
| `created_at`         | `TIMESTAMP` | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.                  |
| `updated_at`         | `TIMESTAMP` | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui.              |

Check constraint `ticket_stocks_quantity_check` memastikan semua quantity tidak
negatif dan:

```text
available_quantity + reserved_quantity + sold_quantity = total_quantity
```

---

### 04. **guests**

**Purpose**: Menyimpan data pembeli tiket tanpa akun user.

| Column       | Data Type      | Constraint / Default                  | Description            |
| ------------ | -------------- | ------------------------------------- | ---------------------- |
| `id`         | `BIGSERIAL`    | Primary key                           | ID guest.              |
| `name`       | `VARCHAR(255)` | Not null                              | Nama pembeli.          |
| `email`      | `VARCHAR(255)` | Not null                              | Email pembeli.         |
| `phone`      | `VARCHAR(20)`  | Not null                              | Nomor telepon pembeli. |
| `address`    | `TEXT`         | Not null                              | Alamat pembeli.        |
| `created_at` | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.     |
| `updated_at` | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui. |
| `deleted_at` | `TIMESTAMP`    | Nullable                              | Waktu soft delete.     |

Indexes:

- `idx_guests_email` pada `email`.

---

### 05. **bookings**

**Purpose**: Menyimpan booking dan snapshot data guest serta event.

| Column                | Data Type      | Constraint / Default                  | Description                   |
| --------------------- | -------------- | ------------------------------------- | ----------------------------- |
| `id`                  | `BIGSERIAL`    | Primary key                           | ID booking.                   |
| `booking_code`        | `VARCHAR(40)`  | Not null, unique                      | Kode booking.                 |
| `guest_id`            | `BIGINT`       | Not null, FK ke `guests.id`           | Guest pemilik booking.        |
| `guest_name`          | `VARCHAR(255)` | Not null                              | Snapshot nama guest.          |
| `guest_email`         | `VARCHAR(255)` | Not null                              | Snapshot email guest.         |
| `guest_phone`         | `VARCHAR(20)`  | Not null                              | Snapshot nomor telepon guest. |
| `guest_address`       | `TEXT`         | Not null                              | Snapshot alamat guest.        |
| `event_id`            | `BIGINT`       | Not null, FK ke `events.id`           | Event yang dipesan.           |
| `event_slug`          | `VARCHAR(160)` | Not null                              | Snapshot slug event.          |
| `event_name`          | `VARCHAR(200)` | Not null                              | Snapshot nama event.          |
| `event_venue_name`    | `VARCHAR(200)` | Not null                              | Snapshot nama venue.          |
| `event_venue_address` | `TEXT`         | Not null                              | Snapshot alamat venue.        |
| `event_starts_at`     | `TIMESTAMPTZ`  | Not null                              | Snapshot waktu mulai event.   |
| `event_ends_at`       | `TIMESTAMPTZ`  | Nullable                              | Snapshot waktu selesai event. |
| `status`              | `VARCHAR(50)`  | Not null, default `pending_payment`   | Status booking.               |
| `total_ticket`        | `INTEGER`      | Not null                              | Total tiket.                  |
| `total_price`         | `BIGINT`       | Not null                              | Total harga.                  |
| `expires_at`          | `TIMESTAMPTZ`  | Not null                              | Batas waktu pembayaran.       |
| `paid_at`             | `TIMESTAMPTZ`  | Nullable                              | Waktu pembayaran berhasil.    |
| `cancelled_at`        | `TIMESTAMPTZ`  | Nullable                              | Waktu booking dibatalkan.     |
| `created_at`          | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.            |
| `updated_at`          | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui.        |
| `deleted_at`          | `TIMESTAMP`    | Nullable                              | Waktu soft delete.            |

Indexes:

- `idx_bookings_guest_id` pada `guest_id`.
- `idx_bookings_event_id` pada `event_id`.
- `idx_bookings_status` pada `status`.

---

### 06. **booking_items**

**Purpose**: Menyimpan detail kategori tiket dalam booking.

| Column                        | Data Type      | Constraint / Default                   | Description                  |
| ----------------------------- | -------------- | -------------------------------------- | ---------------------------- |
| `id`                          | `BIGSERIAL`    | Primary key                            | ID item booking.             |
| `booking_id`                  | `BIGINT`       | Not null, FK ke `bookings.id`          | Booking pemilik item.        |
| `ticket_category_id`          | `BIGINT`       | Not null, FK ke `ticket_categories.id` | Kategori tiket.              |
| `ticket_category_name`        | `VARCHAR(100)` | Not null                               | Snapshot nama kategori.      |
| `ticket_category_description` | `TEXT`         | Nullable                               | Snapshot deskripsi kategori. |
| `quantity`                    | `INTEGER`      | Not null                               | Jumlah tiket.                |
| `unit_price`                  | `BIGINT`       | Not null                               | Harga satuan saat booking.   |
| `subtotal_price`              | `BIGINT`       | Not null                               | Subtotal item.               |
| `created_at`                  | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP`  | Waktu data dibuat.           |
| `updated_at`                  | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP`  | Waktu data diperbarui.       |

Indexes:

- `idx_booking_items_booking_id` pada `booking_id`.
- `idx_booking_items_ticket_category_id` pada `ticket_category_id`.

---

### 07. **payment_transactions**

**Purpose**: Menyimpan transaksi pembayaran dari payment provider.

| Column             | Data Type      | Constraint / Default                  | Description                            |
| ------------------ | -------------- | ------------------------------------- | -------------------------------------- |
| `id`               | `BIGSERIAL`    | Primary key                           | ID payment.                            |
| `booking_id`       | `BIGINT`       | Not null, FK ke `bookings.id`         | Booking yang dibayar.                  |
| `transaction_code` | `VARCHAR(80)`  | Not null                              | Kode transaksi internal.               |
| `provider`         | `VARCHAR(80)`  | Not null                              | Nama provider, misalnya Midtrans/DOKU. |
| `ref_id`           | `VARCHAR(120)` | Not null                              | ID transaksi atau referensi provider.  |
| `payment_method`   | `VARCHAR(80)`  | Nullable                              | Metode pembayaran.                     |
| `status`           | `VARCHAR(50)`  | Not null                              | Status payment.                        |
| `amount`           | `BIGINT`       | Not null                              | Nominal pembayaran.                    |
| `payload`          | `JSONB`        | Not null, default `{}`                | Payload asli provider.                 |
| `paid_at`          | `TIMESTAMPTZ`  | Nullable                              | Waktu pembayaran berhasil.             |
| `expired_at`       | `TIMESTAMPTZ`  | Nullable                              | Waktu payment kedaluwarsa.             |
| `created_at`       | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.                     |
| `updated_at`       | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui.                 |

Indexes:

- `idx_payment_transactions_booking_id` pada `booking_id`.
- Unique `idx_payment_transactions_transaction_code` pada `transaction_code`.
- Unique `idx_payment_transactions_provider_ref_id` pada kombinasi `provider, ref_id`.

---

### 08. **waiting_rooms**

**Purpose**: Menyimpan antrean user sebelum memperoleh akses checkout.

| Column               | Data Type      | Constraint / Default                   | Description                       |
| -------------------- | -------------- | -------------------------------------- | --------------------------------- |
| `id`                 | `BIGSERIAL`    | Primary key                            | ID antrean.                       |
| `event_id`           | `BIGINT`       | Not null, FK ke `events.id`            | Event yang dituju.                |
| `event_name`         | `VARCHAR(200)` | Not null                               | Snapshot nama event.              |
| `ticket_category_id` | `BIGINT`       | Not null, FK ke `ticket_categories.id` | Kategori tiket yang dituju.       |
| `queue_token`        | `VARCHAR(80)`  | Not null, unique                       | Token antrean.                    |
| `checkout_token`     | `VARCHAR(80)`  | Nullable, unique                       | Token akses checkout.             |
| `booking_id`         | `BIGINT`       | Nullable, FK ke `bookings.id`          | Booking yang dibuat dari antrean. |
| `status`             | `VARCHAR(50)`  | Not null, default `waiting`            | Status antrean.                   |
| `failed_reason`      | `TEXT`         | Nullable                               | Alasan proses antrean gagal.      |
| `expired_at`         | `TIMESTAMPTZ`  | Nullable                               | Waktu token kedaluwarsa.          |
| `created_at`         | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP`  | Waktu data dibuat.                |
| `updated_at`         | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP`  | Waktu data diperbarui.            |

Indexes:

- `idx_waiting_rooms_event_name` pada `event_name`.
- `idx_waiting_rooms_ticket_category_id` pada `ticket_category_id`.
- `idx_waiting_rooms_status` pada `status`.
- `idx_waiting_rooms_queue_token` pada `queue_token`.
- `idx_waiting_rooms_checkout_token` pada `checkout_token`.

---

### 09. **outbox_events**

**Purpose**: Menyimpan event yang akan dipublikasikan oleh outbox worker.

| Column            | Data Type      | Constraint / Default                  | Description                    |
| ----------------- | -------------- | ------------------------------------- | ------------------------------ |
| `id`              | `BIGSERIAL`    | Primary key                           | ID outbox event.               |
| `aggregate_type`  | `VARCHAR(80)`  | Not null                              | Tipe aggregate sumber event.   |
| `aggregate_id`    | `BIGINT`       | Not null                              | ID aggregate sumber event.     |
| `event_type`      | `VARCHAR(120)` | Not null                              | Jenis event.                   |
| `payload`         | `JSONB`        | Not null                              | Payload event.                 |
| `status`          | `VARCHAR(50)`  | Not null, default `pending`           | Status pemrosesan outbox.      |
| `attempts`        | `INTEGER`      | Not null, default `0`                 | Jumlah percobaan pemrosesan.   |
| `next_attempt_at` | `TIMESTAMPTZ`  | Not null, default `CURRENT_TIMESTAMP` | Waktu percobaan berikutnya.    |
| `processed_at`    | `TIMESTAMPTZ`  | Nullable                              | Waktu event berhasil diproses. |
| `last_error`      | `TEXT`         | Nullable                              | Error terakhir.                |
| `created_at`      | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data dibuat.             |
| `updated_at`      | `TIMESTAMP`    | Not null, default `CURRENT_TIMESTAMP` | Waktu data diperbarui.         |

Check constraint `outbox_events_status_check` membatasi status menjadi:

```text
pending, processing, sent
```

Indexes:

- `idx_outbox_events_pending` pada kombinasi `status, next_attempt_at, id`.
- `idx_outbox_events_aggregate` pada kombinasi `aggregate_type, aggregate_id`.
