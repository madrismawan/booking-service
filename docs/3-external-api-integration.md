# Integrasi Payment ke Accounting

Payment provider mengirim callback ke booking service. Callback yang valid diproses
secara idempotent, lalu event accounting disimpan ke transactional outbox. Request
HTTP ke accounting tidak dilakukan dari API payment, tetapi oleh worker RabbitMQ
setelah transaksi database selesai.

## Payment Webhook

Endpoint:

```http
POST /api/v1/payments/webhook
X-Payment-Signature: sha256={hex_hmac_sha256}
Content-Type: application/json
```

Signature adalah HMAC-SHA256 dari raw request body menggunakan
`PAYMENT_WEBHOOK_SECRET`.

Payload:

```json
{
  "provider": "midtrans",
  "ref_id": "trx_123",
  "booking_id": 10,
  "payment_method": "virtual_account",
  "status": "paid",
  "amount": 500000,
  "paid_at": "2026-06-24T10:00:00Z"
}
```

Kombinasi `provider + ref_id` menjadi identitas bisnis dan idempotency key payment.
Nilai `provider` dapat berupa `midtrans`, `doku`, atau provider lain. Jika callback
dengan key yang sama sudah pernah diproses, service langsung mengembalikan payment
sebelumnya dengan `duplicate: true` tanpa mengubah booking, stok, atau outbox.

Callback sukses menjalankan satu transaksi database:

1. Lock idempotency key dan booking.
2. Validasi booking masih `pending_payment`, nominal sesuai, dan pembayaran tidak
   melewati waktu kedaluwarsa booking.
3. Simpan `payment_transactions` sebagai `paid`.
4. Ubah booking menjadi `paid`.
5. Pindahkan quantity stok dari `reserved` ke `sold`, naikkan stock version, dan
   buat event perubahan stok.
6. Buat outbox event `accounting.payment_succeeded`.

Jika salah satu proses gagal, seluruh perubahan rollback.

Status response:

- `200`: callback diproses atau merupakan callback duplikat.
- `400`: payload tidak valid.
- `401`: signature tidak valid.
- `404`: booking tidak ditemukan.
- `409`: booking tidak dapat dibayar, nominal berbeda, stok reservasi tidak sesuai,
  atau pembayaran sudah kedaluwarsa.

Contoh response:

```json
{
  "success": true,
  "message": "payment webhook processed",
  "data": {
    "payment_transaction_id": 12,
    "transaction_code": "PAY-8F21A1C53093F16008A7C635",
    "booking_id": 10,
    "booking_status": "paid",
    "payment_status": "paid",
    "duplicate": false
  }
}
```

## Event Accounting

Outbox worker mengirim event ke:

```text
accounting.payment_succeeded.queue
```

Message:

```json
{
  "event_id": 25,
  "event_type": "accounting.payment_succeeded",
  "schema_version": 1,
  "payment_transaction_id": 12,
  "booking_id": 10,
  "booking_code": "BKG-123",
  "amount": 500000,
  "paid_at": "2026-06-24T10:00:00Z",
  "attempt": 1
}
```

Accounting worker melakukan HTTP `POST` ke `ACCOUNTING_API_URL` dengan:

```http
Authorization: Bearer {ACCOUNTING_API_TOKEN}
Idempotency-Key: accounting-payment-{event_id}
Content-Type: application/json
```

Sistem accounting wajib menyimpan idempotency key karena delivery RabbitMQ bersifat
at-least-once.

## Retry dan DLQ

Maksimal terdapat tiga HTTP call:

1. Attempt 1 langsung dari queue utama.
2. Jika gagal, attempt 2 menunggu 5 detik di
   `accounting.payment_succeeded.retry.5s.queue`.
3. Jika gagal lagi, attempt 3 menunggu 10 detik di
   `accounting.payment_succeeded.retry.10s.queue`.

Retry dilakukan untuk network error, HTTP `429`, dan HTTP `5xx`. HTTP `4xx` selain
`429` dianggap permanent failure dan langsung dikirim ke:

```text
accounting.payment_succeeded.dlq
```

Setelah attempt ketiga gagal, message juga masuk ke DLQ dengan `attempt` dan
`last_error` terakhir.

## Environment

```env
PAYMENT_WEBHOOK_SECRET=change-me
ACCOUNTING_API_URL=http://host.docker.internal:8081/api/v1/payments
ACCOUNTING_API_TOKEN=change-me
ACCOUNTING_API_TIMEOUT=10s
```
