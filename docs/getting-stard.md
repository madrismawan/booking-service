# Getting Started

## Persiapan

Pastikan perangkat sudah memiliki:

- Docker dan Docker Compose
- `make`

## Menjalankan Project

1. Buat file environment:

   ```bash
   cp .env.example .env
   ```

2. Jalankan migration:

   ```bash
   make docker-migrate
   ```

3. Isi data awal:

   ```bash
   make docker-seed
   ```

4. Jalankan API, worker, PostgreSQL, dan RabbitMQ:

   ```bash
   make docker-up
   ```

Booking API dapat diakses melalui:

```text
http://localhost:8080
```

RabbitMQ Management dapat diakses melalui:

```text
http://localhost:15672
```

## Menjalankan Test

```bash
make test
make test-docs
```

## Menghentikan Project

```bash
make docker-down
```

Untuk menghapus data lama, menjalankan ulang migration, dan mengisi seed:

```bash
make docker-fresh
```
