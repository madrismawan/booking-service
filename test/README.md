Folder ini berisi integration test untuk skenario pada dokumen 1–5:

- `doc_1_race_condition_test.go`: dua booking bersamaan tidak menyebabkan overselling.
- `doc_2_high_traffic_processing_test.go`: banyak goroutine masuk waiting room dan
  setiap request memiliki outbox event.
- `doc_3_external_api_integration_test.go`: webhook payment mengubah booking, stok,
  payment transaction, dan accounting outbox secara atomik.
- `doc_4_duplicate_request_test.go`: dua webhook identik diproses satu kali.
- `doc_5_data_synchronization_test.go`: perubahan stok menaikkan version, membuat
  outbox, dan menghasilkan ETag yang sesuai.

Jalankan:

```bash
make test-docs
```
