# ---------- BUILD STAGE ----------
FROM golang:1.23-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Кэш зависимостей
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Копируем код
COPY . .

# Сборка
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main ./cmd/api

# ---------- FINAL STAGE ----------
FROM alpine:3.19

# Сертификаты + timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем бинарник
COPY --from=builder /app/main .

# Копируем нужные файлы
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/web ./web
COPY --from=builder /app/.env.example ./.env.example

# Порт
EXPOSE 8080

# Healthcheck (очень полезно)
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./main"]
