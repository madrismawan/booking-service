# Sinkronisasi Data Stock

`ticket_stocks` menjadi source of truth untuk jumlah stok tiket. Setiap perubahan stok
operasional menaikkan `version` dan membuat event outbox dalam transaksi database yang
sama. Seeder tidak membuat event.

## Event RabbitMQ

Outbox worker mengirim event ke queue:

```text
ticket_stock.changed.queue
```

Payload:

```json
{
  "event_id": 123,
  "event_type": "ticket_stock.changed",
  "schema_version": 1,
  "ticket_category_id": 10,
  "stock_version": 7,
  "snapshot_url": "/api/v1/ticket-categories/10/stock",
  "changed_at": "2026-06-24T10:00:00Z"
}
```

Delivery menggunakan pola at-least-once. Consumer harus memakai `event_id` untuk
deduplikasi dan tidak boleh menganggap event selalu datang berurutan.

## API Snapshot

Consumer mengambil data terbaru melalui:

```http
GET /api/v1/ticket-categories/:ticket_category_id/stock
```

Response memuat total, available, reserved, sold, version, dan waktu update. Response
juga membawa ETag:

```text
"ticket-stock-{ticket_category_id}-v{version}"
```

Consumer dapat mengirim ETag lokal:

```http
If-None-Match: "ticket-stock-10-v7"
```

API mengembalikan `304 Not Modified` jika version masih sama, atau `200 OK` dengan
snapshot dan ETag terbaru jika sudah berubah.

## Aturan Consumer

1. Abaikan event jika `stock_version <= local_version`.
2. Jika event lebih baru, panggil `snapshot_url` menggunakan `If-None-Match`.
3. Simpan seluruh snapshot dan `version` dari response API.
4. Jangan menurunkan local version ketika event lama atau duplikat datang.

Outbox event dibuat atomik bersama perubahan stok. Jika transaksi stok rollback,
event juga rollback. Jika worker crash setelah publish tetapi sebelum menandai event
sebagai sent, event dapat terkirim ulang dan harus ditangani sebagai duplikat.
