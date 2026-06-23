# Database Schema - Ticket Booking System

---

### 01. **events**

**Purpose**: Menyimpan data konser atau acara.

| Column          | Data Type      | Description                                           |
| --------------- | -------------- | ----------------------------------------------------- |
| `id`            | `BIGINT (PK)`  | Primary key.                                          |
| `slug`          | `VARCHAR(160)` | Slug unik untuk URL booking event.                    |
| `name`          | `VARCHAR(200)` | Nama konser atau acara.                               |
| `description`   | `TEXT`         | Deskripsi acara.                                      |
| `venue_name`    | `VARCHAR(200)` | Nama venue.                                           |
| `venue_address` | `TEXT`         | Alamat venue.                                         |
| `starts_at`     | `TIMESTAMPTZ`  | Waktu mulai acara.                                    |
| `ends_at`       | `TIMESTAMPTZ`  | Waktu selesai acara.                                  |
| `status`        | `VARCHAR(50)`  | Status acara: draft, published, cancelled, completed. |
| `created_at`    | `TIMESTAMP`    | Timestamp column.                                     |
| `updated_at`    | `TIMESTAMP`    | Timestamp column.                                     |
| `deleted_at`    | `TIMESTAMP`    | Soft delete timestamp.                                |

---

### 02. **ticket_categories**

**Purpose**: Menyimpan data kategori tiket untuk setiap konser, seperti VIP, Festival, atau Tribune.

| Column            | Data Type      | Description                           |
| ----------------- | -------------- | ------------------------------------- |
| `id`              | `BIGINT (PK)`  | Primary key.                          |
| `event_id`        | `BIGINT (FK)`  | Foreign key reference to `events.id`. |
| `name`            | `VARCHAR(100)` | Nama kategori tiket.                  |
| `description`     | `TEXT`         | Deskripsi kategori tiket.             |
| `price`           | `BIGINT`       | Harga tiket.                          |
| `sale_starts_at`  | `TIMESTAMPTZ`  | Waktu mulai penjualan tiket.          |
| `sale_ends_at`    | `TIMESTAMPTZ`  | Waktu selesai penjualan tiket.        |
| `max_per_booking` | `INTEGER`      | Maksimal jumlah tiket per booking.    |
| `created_at`      | `TIMESTAMP`    | Timestamp column.                     |
| `updated_at`      | `TIMESTAMP`    | Timestamp column.                     |
| `deleted_at`      | `TIMESTAMP`    | Soft delete timestamp.                |

---

### 03. **ticket_stocks**

**Purpose**: Menyimpan stok tiket per kategori dan mencegah overselling saat traffic tinggi.

| Column               | Data Type     | Description                                      |
| -------------------- | ------------- | ------------------------------------------------ |
| `id`                 | `BIGINT (PK)` | Primary key.                                     |
| `ticket_category_id` | `BIGINT (FK)` | Foreign key reference to `ticket_categories.id`. |
| `total_quantity`     | `INTEGER`     | Total jumlah tiket.                              |
| `available_quantity` | `INTEGER`     | Jumlah tiket yang tersedia untuk booking.        |
| `reserved_quantity`  | `INTEGER`     | Jumlah tiket yang sedang direservasi.            |
| `sold_quantity`      | `INTEGER`     | Jumlah tiket yang sudah terjual.                 |
| `version`            | `BIGINT`      | Versi monotonik untuk sinkronisasi stok.         |
| `created_at`         | `TIMESTAMP`   | Timestamp column.                                |
| `updated_at`         | `TIMESTAMP`   | Timestamp column.                                |

---

### 04. **guests**

**Purpose**: Menyimpan data pembeli tiket tanpa akun user.

| Column       | Data Type      | Description            |
| ------------ | -------------- | ---------------------- |
| `id`         | `BIGINT (PK)`  | Primary key.           |
| `name`       | `VARCHAR(255)` | Nama lengkap pembeli.  |
| `email`      | `VARCHAR(255)` | Email pembeli.         |
| `phone`      | `VARCHAR(20)`  | Nomor telepon pembeli. |
| `address`    | `TEXT`         | Alamat pembeli.        |
| `created_at` | `TIMESTAMP`    | Timestamp column.      |
| `updated_at` | `TIMESTAMP`    | Timestamp column.      |
| `deleted_at` | `TIMESTAMP`    | Soft delete timestamp. |

---

### 05. **bookings**

**Purpose**: Menyimpan data booking tiket yang dibuat oleh guest.

| Column                | Data Type      | Description                                                |
| --------------------- | -------------- | ---------------------------------------------------------- |
| `id`                  | `BIGINT (PK)`  | Primary key.                                               |
| `booking_code`        | `VARCHAR(40)`  | Kode booking unik yang ditampilkan ke pembeli.             |
| `guest_id`            | `BIGINT (FK)`  | Foreign key reference to `guests.id`.                      |
| `guest_name`          | `VARCHAR(255)` | Snapshot nama lengkap pembeli saat booking dibuat.         |
| `guest_email`         | `VARCHAR(255)` | Snapshot email pembeli saat booking dibuat.                |
| `guest_phone`         | `VARCHAR(20)`  | Snapshot nomor telepon pembeli saat booking dibuat.        |
| `guest_address`       | `TEXT`         | Snapshot alamat pembeli saat booking dibuat.               |
| `event_id`            | `BIGINT (FK)`  | Foreign key reference to `events.id`.                      |
| `event_slug`          | `VARCHAR(160)` | Snapshot slug event saat booking dibuat.                   |
| `event_name`          | `VARCHAR(200)` | Snapshot nama event saat booking dibuat.                   |
| `event_venue_name`    | `VARCHAR(200)` | Snapshot nama venue saat booking dibuat.                   |
| `event_venue_address` | `TEXT`         | Snapshot alamat venue saat booking dibuat.                 |
| `event_starts_at`     | `TIMESTAMPTZ`  | Snapshot waktu mulai event saat booking dibuat.            |
| `event_ends_at`       | `TIMESTAMPTZ`  | Snapshot waktu selesai event saat booking dibuat.          |
| `status`              | `VARCHAR(50)`  | Status booking: pending_payment, paid, cancelled, expired. |
| `total_ticket`        | `INTEGER`      | Total jumlah tiket yang dibooking.                         |
| `total_price`         | `BIGINT`       | Total harga booking.                                       |
| `expires_at`          | `TIMESTAMPTZ`  | Waktu kedaluwarsa pembayaran.                              |
| `paid_at`             | `TIMESTAMPTZ`  | Waktu pembayaran berhasil.                                 |
| `cancelled_at`        | `TIMESTAMPTZ`  | Waktu booking dibatalkan.                                  |
| `created_at`          | `TIMESTAMP`    | Timestamp column.                                          |
| `updated_at`          | `TIMESTAMP`    | Timestamp column.                                          |
| `deleted_at`          | `TIMESTAMP`    | Soft delete timestamp.                                     |

---

### 06. **booking_items**

**Purpose**: Menyimpan detail kategori tiket yang dipilih dalam satu booking.

| Column                        | Data Type      | Description                                            |
| ----------------------------- | -------------- | ------------------------------------------------------ |
| `id`                          | `BIGINT (PK)`  | Primary key.                                           |
| `booking_id`                  | `BIGINT (FK)`  | Foreign key reference to `bookings.id`.                |
| `ticket_category_id`          | `BIGINT (FK)`  | Foreign key reference to `ticket_categories.id`.       |
| `ticket_category_name`        | `VARCHAR(100)` | Snapshot nama kategori tiket saat booking dibuat.      |
| `ticket_category_description` | `TEXT`         | Snapshot deskripsi kategori tiket saat booking dibuat. |
| `quantity`                    | `INTEGER`      | Jumlah tiket untuk kategori ini.                       |
| `unit_price`                  | `BIGINT`       | Harga satuan tiket saat booking dibuat.                |
| `subtotal_price`              | `BIGINT`       | Total harga untuk item ini.                            |
| `created_at`                  | `TIMESTAMP`    | Timestamp column.                                      |
| `updated_at`                  | `TIMESTAMP`    | Timestamp column.                                      |

---

### 07. **payment_transactions**

**Purpose**: Menyimpan transaksi pembayaran untuk sebuah booking.

| Column                    | Data Type      | Description                                                 |
| ------------------------- | -------------- | ----------------------------------------------------------- |
| `id`                      | `BIGINT (PK)`  | Primary key.                                                |
| `booking_id`              | `BIGINT (FK)`  | Foreign key reference to `bookings.id`.                     |
| `transaction_code`        | `VARCHAR(80)`  | Kode transaksi internal.                                    |
| `provider`                | `VARCHAR(80)`  | Nama payment provider.                                      |
| `provider_transaction_id` | `VARCHAR(120)` | ID transaksi dari payment provider.                         |
| `provider_event_id`       | `VARCHAR(120)` | ID event webhook dari payment provider.                     |
| `payment_method`          | `VARCHAR(80)`  | Metode pembayaran, misalnya virtual_account atau ewallet.   |
| `status`                  | `VARCHAR(50)`  | Status transaksi: pending, paid, failed, expired, refunded. |
| `amount`                  | `BIGINT`       | Nominal transaksi.                                          |
| `payload`                 | `JSONB`        | Payload dari payment provider.                              |
| `paid_at`                 | `TIMESTAMPTZ`  | Waktu transaksi berhasil dibayar.                           |
| `expired_at`              | `TIMESTAMPTZ`  | Waktu transaksi kedaluwarsa.                                |
| `created_at`              | `TIMESTAMP`    | Timestamp column.                                           |
| `updated_at`              | `TIMESTAMP`    | Timestamp column.                                           |

<!--

---

### 08. **idempotency_keys**

**Purpose**: Menyimpan data idempotency untuk mencegah booking ganda dari request yang sama.

| Column          | Data Type      | Description                             |
| --------------- | -------------- | --------------------------------------- |
| `id`            | `BIGINT (PK)`  | Primary key.                            |
| `key`           | `VARCHAR(150)` | Idempotency key dari client.            |
| `request_hash`  | `CHAR(64)`     | Hash dari payload request.              |
| `booking_id`    | `BIGINT (FK)`  | Foreign key reference to `bookings.id`. |
| `response_code` | `INTEGER`      | HTTP response code yang disimpan.       |
| `response_body` | `JSONB`        | HTTP response body yang disimpan.       |
| `locked_until`  | `TIMESTAMPTZ`  | Waktu kedaluwarsa lock sementara.       |
| `created_at`    | `TIMESTAMP`    | Timestamp column.                       |
| `updated_at`    | `TIMESTAMP`    | Timestamp column.                       |

---

### 09. **outbox_events**

**Purpose**: Menyimpan event untuk integrasi external, seperti sinkronisasi accounting.

| Column            | Data Type      | Description                                              |
| ----------------- | -------------- | -------------------------------------------------------- |
| `id`              | `BIGINT (PK)`  | Primary key.                                             |
| `aggregate_type`  | `VARCHAR(80)`  | Nama aggregate, misalnya booking atau transaction.       |
| `aggregate_id`    | `BIGINT`       | ID aggregate terkait.                                    |
| `event_type`      | `VARCHAR(120)` | Jenis event, misalnya booking.created atau booking.paid. |
| `payload`         | `JSONB`        | Payload event.                                           |
| `status`          | `VARCHAR(50)`  | Status outbox: pending, processing, sent, failed.        |
| `attempts`        | `INTEGER`      | Jumlah percobaan pengiriman.                             |
| `next_attempt_at` | `TIMESTAMPTZ`  | Waktu retry berikutnya.                                  |
| `processed_at`    | `TIMESTAMPTZ`  | Waktu event berhasil diproses.                           |
| `last_error`      | `TEXT`         | Pesan error terakhir saat pengiriman.                    |
| `created_at`      | `TIMESTAMP`    | Timestamp column.                                        |
| `updated_at`      | `TIMESTAMP`    | Timestamp column.                                        |

---

### 10. **audit_logs**

**Purpose**: Menyimpan aktivitas penting sistem untuk kebutuhan traceability.

| Column        | Data Type      | Description                           |
| ------------- | -------------- | ------------------------------------- |
| `id`          | `BIGINT (PK)`  | Primary key.                          |
| `guest_id`    | `BIGINT (FK)`  | Foreign key reference to `guests.id`. |
| `action`      | `VARCHAR(120)` | Nama aksi.                            |
| `entity_type` | `VARCHAR(80)`  | Tipe entity yang terdampak aksi.      |
| `entity_id`   | `BIGINT`       | ID entity yang terdampak aksi.        |
| `metadata`    | `JSONB`        | Metadata audit tambahan.              |
| `ip_address`  | `VARCHAR(45)`  | Alamat IP actor.                      |
| `user_agent`  | `TEXT`         | User agent actor.                     |
| `created_at`  | `TIMESTAMP`    | Timestamp column.                     |

-->
