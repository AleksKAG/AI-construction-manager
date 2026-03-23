# Dockerfile
FROM golang:1.22-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копирование модулей для кэширования
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Финальный образ
FROM alpine:3.19

# Установка корневых сертификатов и часового пояса
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копирование бинарника
COPY --from=builder /app/main .

# Копирование миграций
COPY --from=builder /app/migrations ./migrations

# Копирование frontend (ВАЖНО для Web UI)
COPY --from=builder /app/web ./web

# Копирование .env.example (для справки)
COPY --from=builder /app/.env.example ./.env.example

EXPOSE 8080

CMD ["./main"]
