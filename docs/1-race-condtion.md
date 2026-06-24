# Race Condition Pada Booking Ticket

## Analisa

Pada proses booking ticket, bagian yang paling rawan race condition adalah saat sistem mengurangi stock ticket. Ini terjadi di flow `CreateBooking` pada file:

`internal/service/booking_service.go`

Proses booking dijalankan di dalam database transaction:

```go
err := s.txManager.Transaction(func(tx *gorm.DB) error {
```

Di dalam transaction tersebut, service memanggil:

```go
ticketStockService.ReserveForUpdate(category.ID, req.Quantity)
```

Bagian ini penting karena stock ticket tidak hanya dibaca, tapi juga langsung diubah. Jika banyak request booking masuk bersamaan, sistem harus memastikan proses cek stock dan pengurangan stock berjalan secara aman.

Di `internal/service/ticket_stock_service.go`, logic pengurangan stock dilakukan setelah stock berhasil diambil:

```go
if stock.AvailableQuantity < quantity {
	return nil, repository.ErrInsufficientStock
}

stock.AvailableQuantity -= quantity
stock.ReservedQuantity += quantity
```

Jadi urutannya adalah:

1. Ambil data stock
2. Cek apakah stock cukup
3. Kurangi `AvailableQuantity`
4. Tambah `ReservedQuantity`
5. Simpan perubahan stock

Kalau proses ini tidak dikunci, ada kemungkinan beberapa request membaca nilai stock yang sama sebelum salah satu request selesai menyimpan perubahan.

## Penggunaan `FOR UPDATE`

Locking dilakukan di file:

`internal/repository/ticket_stock_repository.go`

```go
r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
	Where("ticket_category_id = ?", ticketCategoryID).
	First(&stock)
```

Kode ini menggunakan konsep `SELECT ... FOR UPDATE`.

Dengan `FOR UPDATE`, row stock yang sedang diproses akan dikunci selama transaction masih berjalan. Request lain yang ingin membaca row yang sama untuk diubah harus menunggu sampai transaction sebelumnya selesai.

## Trade-off Menggunakan `FOR UPDATE`

1. Request lain bisa menunggu

   Ketika satu transaction sedang mengunci row stock, request lain yang ingin mengubah stock yang sama harus menunggu. Ini bisa membuat response lebih lambat ketika traffic tinggi.

2. Ada risiko deadlock jika locking tidak konsisten

   Jika di masa depan ada proses lain yang mengunci beberapa table atau row dengan urutan berbeda, database bisa mengalami deadlock. Untuk menghindarinya, urutan akses data perlu dibuat konsisten.
