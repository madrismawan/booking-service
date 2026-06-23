# High Traffic Processing Dengan Waiting Room Queue

## Analisa

Pada traffic tinggi, misalnya 10.000 request per menit, proses booking sebaiknya tidak langsung masuk ke checkout. Jika semua request langsung checkout, maka database akan menerima banyak transaction yang sama-sama mencoba mengurangi stock ticket.

Untuk mengurangi beban tersebut, user dimasukkan dulu ke waiting room. API cukup menerima request, membuat `queue_token`, menyimpan data ke database, lalu mengirim message ke RabbitMQ.

User akan mendapat response awal:

```json
{
  "message": "you are in queue",
  "queue_token": "queue_xxx",
  "ticket_category_id": 1,
  "status": "waiting"
}
```

Dengan flow ini, request tetap bisa diterima oleh sistem, tapi tidak semuanya langsung diproses menjadi booking.

## Penggunaan RabbitMQ

RabbitMQ dipakai untuk menampung antrian waiting room. Queue bisa membawa `ticket_category_id`, supaya checkout token terikat ke kategori tiket yang spesifik.

Contoh queue:

```text
waiting_room.queue
```

Misalnya:

```text
waiting_room.queue
```

Endpoint untuk masuk queue:

```http
POST /api/v1/ticket-categories/:ticket_category_id/queue/join
```

Pada endpoint ini sistem akan:

1. Validasi ticket category berdasarkan `ticket_category_id`
2. Generate `queue_token`
3. Simpan data queue ke database dengan status `waiting` dan `ticket_category_id`
4. Publish message ke RabbitMQ
5. Return response `you are in queue`

Setelah itu, user bisa cek status queue lewat endpoint:

```http
GET /api/v1/queue/:queue_token/status
```

Jika status sudah `ready`, response akan mengembalikan `checkout_token`.

```json
{
  "status": "ready",
  "ticket_category_id": 1,
  "checkout_token": "checkout_xxx",
  "expired_at": "2026-06-23T10:05:00Z"
}
```

## Worker Waiting Room

Worker bertugas membaca message dari RabbitMQ dan mengubah status user dari `waiting` menjadi `ready`.

Saat worker memproses queue, sistem akan:

1. Ambil message dari RabbitMQ
2. Cari data queue berdasarkan `queue_token`
3. Generate `checkout_token`
4. Set status menjadi `ready`
5. Set `expired_at` selama 5 menit

`checkout_token` ini dipakai user untuk lanjut ke proses checkout. Jika token sudah lewat dari 5 menit, status bisa diubah menjadi `expired`.

## Hubungan Dengan Checkout

Waiting room tidak langsung mengurangi stock ticket. Waiting room hanya mengatur giliran user agar tidak semua request langsung masuk checkout.

Stock baru dikurangi saat user sudah punya `checkout_token` dan melakukan checkout. Pada bagian ini, flow akan masuk ke proses booking yang memakai `FOR UPDATE`, seperti dijelaskan di:

`docs/1-race-condtion.md`

Jadi pembagiannya:

```text
Waiting room:
- menerima traffic besar
- menyimpan queue
- membuat checkout_token

Checkout:
- validasi checkout_token
- cek stock
- decrement stock dengan FOR UPDATE
- membuat booking
```

Dengan cara ini, RabbitMQ dipakai untuk mengatur traffic, sedangkan database transaction tetap dipakai untuk menjaga stock.

## Status Queue

Status sederhana yang bisa dipakai:

```text
waiting
ready
expired
checkout_started
completed
failed
```

Penjelasan singkat:

1. `waiting`

   User sudah masuk queue, tapi belum boleh checkout.

2. `ready`

   User sudah mendapat `checkout_token` dan boleh checkout.

3. `expired`

   `checkout_token` sudah lewat dari 5 menit.

4. `checkout_started`

   User sudah mulai proses checkout.

5. `completed`

   Booking berhasil dibuat.

6. `failed`

   Proses gagal, misalnya token tidak valid atau stock sudah habis.

## Trade-off Menggunakan Waiting Room

1. User tidak langsung checkout

   User harus menunggu sampai status queue menjadi `ready`. Ini membuat flow sedikit lebih panjang, tapi sistem lebih aman saat traffic tinggi.

2. Perlu proses worker

   Karena antrian diproses oleh worker, sistem perlu memastikan worker selalu berjalan dan bisa memproses message dari RabbitMQ.

3. Perlu handling expired token

   `checkout_token` hanya berlaku 5 menit. Sistem perlu menangani token yang sudah expired agar slot checkout tidak tertahan terlalu lama.

4. Data request bisa banyak

   Saat traffic besar, database akan menyimpan banyak data queue. Jadi perlu indexing yang baik, terutama pada `queue_token`, `ticket_category_id`, dan `status`.

## Kesimpulan

Waiting room dengan RabbitMQ membantu menahan lonjakan traffic sebelum user masuk checkout. API bisa tetap cepat memberi response `you are in queue`, lalu worker memproses antrian secara bertahap.

Request yang masuk bisa tetap disimpan di database, tapi statusnya berbeda-beda. Tidak semua request langsung menjadi booking. User baru masuk checkout ketika status queue sudah `ready` dan memiliki `checkout_token`.

Untuk keamanan stock, proses decrement tetap dilakukan di checkout menggunakan `FOR UPDATE`.
