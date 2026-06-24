# Booking Service

Booking Service adalah backend sederhana untuk proses pemesanan tiket event.

## Arsitektur

Project menggunakan repository pattern dengan pembagian tanggung jawab:

- `handler` menerima dan mengembalikan HTTP request.
- `service` berisi proses dan aturan bisnis.
- `repository` menangani akses serta perubahan data di PostgreSQL.
- `worker` memproses event asynchronous melalui RabbitMQ.
- `model` dan `dto` mendefinisikan struktur data internal dan API.

Handler hanya mengakses service, sedangkan service menggunakan repository atau service
lain yang dibutuhkan. Pembagian ini membuat proses bisnis lebih mudah dipahami,
dikembangkan, dan diuji.

## Alur Utama

1. User masuk ke waiting room sebelum melakukan checkout.
2. Worker memeriksa ketersediaan stok dan memberikan checkout token.
3. User membuat booking menggunakan checkout token.
4. Stock direservasi menggunakan database transaction dan row locking untuk mencegah
   overselling.
5. Payment provider seperti Midtrans atau DOKU mengirim webhook ketika pembayaran
   berhasil.
6. Booking diubah menjadi `paid` dan stock dipindahkan dari `reserved` menjadi `sold`.

## Messaging dan Outbox Pattern

Aplikasi menggunakan RabbitMQ untuk proses asynchronous. Event tidak langsung
dipublikasikan dari request API, tetapi disimpan terlebih dahulu ke tabel
`outbox_events` dalam database transaction yang sama dengan perubahan data utama.

Outbox worker kemudian mengambil dan memublikasikan event tersebut ke RabbitMQ. Pola
ini mencegah event hilang ketika database berhasil diperbarui tetapi RabbitMQ sedang
tidak tersedia.

## Dokumentasi

- [Race condition pada booking](1-race-condtion.md)
- [High traffic processing](2-high-traffic-processing.md)
- [Integrasi payment dan accounting](3-external-api-integration.md)
- [Duplicate payment webhook](4-duplicate-request.md)
- [Sinkronisasi data stock](5-data-synchronization.md)
- [Database schema](db-schema.md)

Integration test untuk setiap skenario tersedia pada folder `test/` dan dapat
dijalankan dengan:

```bash
make test-docs
```
