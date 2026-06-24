# Duplicate Request Pada Payment Webhook

## Analisa

Pada proses payment webhook, bagian yang paling rawan duplicate request adalah saat
service menerima status pembayaran dari provider seperti Midtrans atau DOKU.

Endpoint webhook didaftarkan pada file:

`internal/handler/handler.go`

```go
api.POST("/payments/webhook", h.paymentWebhook)
```

Request kemudian diterima oleh `paymentWebhook` pada file:

`internal/handler/payment_handler.go`

Handler membaca payload, memverifikasi signature, lalu meneruskan request ke:

```go
result, err := h.paymentService.ProcessWebhook(req, body)
```

Provider dapat mengirim webhook yang sama lebih dari sekali, misalnya ketika response
dari booking service tidak diterima karena gangguan jaringan. Dua webhook yang sama
juga dapat masuk hampir bersamaan.

Jika tidak ada pencegahan, setiap request dapat:

1. Membuat row payment baru
2. Mengubah status booking menjadi `paid`
3. Memindahkan stock dari `reserved` menjadi `sold`
4. Membuat outbox event accounting

Karena itu, pencegahan duplicate harus dilakukan sebelum payment dan data terkait
diubah.

## Idempotency Key

Pencegahan utama dilakukan di file:

`internal/service/payment_service.go`

Identitas payment menggunakan kombinasi:

```text
provider + ref_id
```

`provider` berisi sumber payment seperti `midtrans` atau `doku`.
`ref_id` adalah ID transaksi atau referensi unik yang dikirim oleh provider.

Setelah memperoleh lock, service mencari payment yang sudah tersimpan:

```go
existing, err := paymentRepo.FindByRefID(req.Provider, req.RefID)
```

Jika payment sudah tersedia, service langsung mengembalikan payment tersebut sebagai
duplicate dengan response `200 OK`. Booking, stock, dan outbox tidak diproses kembali.

## Penggunaan Advisory Lock

Pengecekan payment saja belum cukup karena dua request dapat melakukan pengecekan pada
waktu yang hampir bersamaan.

Locking dilakukan di file:

`internal/service/payment_service.go`

```go
lockKey := req.Provider + "|ref|" + req.RefID
```

Setiap key dikunci menggunakan PostgreSQL transaction advisory lock:

```go
tx.Exec(
	"SELECT pg_advisory_xact_lock(hashtextextended(?, 0))",
	lockKey,
)
```

Dengan lock ini, webhook dengan `provider + ref_id` yang sama diproses secara
berurutan.

Contoh alurnya:

1. Request pertama memperoleh lock
2. Request kedua dengan `ref_id` yang sama menunggu
3. Request pertama membuat payment dan menyelesaikan transaction
4. Request kedua memperoleh lock
5. Request kedua menemukan payment yang sudah tersedia
6. Request kedua langsung mengembalikan `200 OK`

## Penggunaan Unique Index

Database tetap menjadi perlindungan terakhir terhadap duplicate row.

Unique index dibuat pada file:

`migration/000007_create_payment_transactions.up.sql`

```sql
CREATE UNIQUE INDEX idx_payment_transactions_provider_ref_id
  ON payment_transactions (provider, ref_id);
```

`transaction_code` internal juga dibuat unik sebagai referensi payment di dalam
booking service.

Advisory lock mencegah request diproses bersamaan, sedangkan unique index memastikan
database tetap menolak duplicate jika terjadi kesalahan pada logic aplikasi.

## Trade-off Pencegahan Duplicate

1. Request duplicate dapat menunggu

   Ketika webhook dengan ID yang sama sedang diproses, request berikutnya harus
   menunggu advisory lock dilepas setelah transaction selesai.

2. Unique index bergantung pada identitas dari provider

   Provider harus mengirim `ref_id` yang sama pada setiap retry. Jika provider
   mengirim ID baru, request akan dianggap sebagai payment baru.
