FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка бинарника
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Копирование бинарника
COPY --from=builder /app/main .

# Копирование миграций
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./main"]
