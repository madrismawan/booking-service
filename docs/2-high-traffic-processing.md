# High Traffic Processing Dengan Waiting Room

## Analisa

Masalah high traffic terjadi ketika banyak user masuk ke flow beli tiket di waktu yang sama.

Solusi di aplikasi ini adalah membuat waiting room sederhana sebelum checkout. Waiting room hanya menampung request user, menyimpan queue ke database, lalu mengirim message ke RabbitMQ lewat publisher.

Flow ini terjadi sebelum proses pada:

`docs/1-race-condtion.md`

Jadi waiting room bukan tempat mengurangi stock. Waiting room hanya mengatur apakah user boleh lanjut checkout atau tidak.

## Flow Waiting Room

User masuk lewat endpoint:

```http
POST /api/v1/ticket-categories/:ticket_category_id/queue/join
```

Service akan membuat data `waiting_rooms` dengan status `waiting`:

```go
waitingRoom := model.WaitingRoom{
	EventID:          category.EventID,
	EventName:        category.Event.Name,
	TicketCategoryID: category.ID,
	QueueToken:       queueToken,
	Status:           model.WaitingRoomStatusWaiting,
}
```

Setelah data tersimpan, service publish message ke RabbitMQ:

```go
err = s.publisher.PublishJSON(ctx, rabbitmq.WaitingRoomQueue, rabbitmq.WaitingRoomMessage{
	TicketCategoryID: waitingRoom.TicketCategoryID,
	QueueToken:       waitingRoom.QueueToken,
	CreatedAt:        time.Now(),
})
```

Publisher ini menjadi pemisah antara request user dan proses worker. Response ke user cukup berisi `queue_token`, supaya user bisa cek status queue tanpa menunggu proses checkout.

## Proses Worker

Worker membaca message dari RabbitMQ secara bertahap. Setelah menerima message, worker memanggil:

```go
waitingRoom, processed, err := w.waitingRoomService.MarkReady(message.QueueToken, checkoutTokenTTL)
```

Di `MarkReady`, sistem mencari data waiting room berdasarkan `queue_token`:

Lookup ini tidak memakai `FOR UPDATE` karena worker hanya memproses queue spesifik berdasarkan token. Prosesnya tidak perlu buru-buru mengunci row seperti proses pengurangan stock.

Setelah itu worker mengecek stock ticket category:

```go
stock, err := service.ticketStockService.FindByTicketCategoryID(record.TicketCategoryID)
```

Jika stock masih memungkinkan, worker membuat `checkout_token` dan mengubah status menjadi `ready`. Jika stock habis atau slot checkout aktif sudah penuh, status menjadi `failed` dan `failed_reason` disimpan.

## Sebelum Masuk Checkout

User yang statusnya `ready` bisa lanjut ke booking memakai `checkout_token` dan `quantity`.

Di booking, sistem tetap wajib validasi bahwa waiting room masih `ready`:

```go
if waitingRoom.TicketCategoryID != category.ID || waitingRoom.Status != model.WaitingRoomStatusReady {
	return repository.ErrInvalidCheckout
}
```

Baru setelah itu stock dikurangi lewat flow booking yang memakai `FOR UPDATE`, seperti dijelaskan di:

`docs/1-race-condtion.md`

Ringkasnya:

```text
Waiting room:
- menampung request
- menyimpan queue
- publish message ke RabbitMQ
- worker menentukan ready atau failed
- membuat checkout_token jika user boleh checkout

Checkout:
- menerima checkout_token dan quantity
- validasi status ready
- mengurangi stock dengan FOR UPDATE
- membuat booking
```

## Trade-off

1. User harus menunggu

   User tidak langsung checkout. User perlu cek status queue sampai menjadi `ready`.

2. Perlu worker

   Jika worker mati, request tetap tersimpan, tapi status bisa tertahan di `waiting`.

3. Ada kemungkinan gagal setelah masuk queue

   Semua request bisa ditampung, tapi tidak semuanya pasti mendapat checkout token. Jika stock tidak cukup, status menjadi `failed`.
