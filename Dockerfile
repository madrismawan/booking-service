FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/seed ./cmd/seed
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker

FROM alpine:3.22

WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/migrate /app/migrate
COPY --from=builder /bin/seed /app/seed
COPY --from=builder /bin/worker /app/worker
COPY migration /app/migration

USER app

EXPOSE 8080

CMD ["/app/api"]
