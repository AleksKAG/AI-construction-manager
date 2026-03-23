FROM golang:1.22-alpine AS builder

# Устанавливаем git + необходимые пакеты (Alpine требует --no-cache)
RUN apk add --no-cache git

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/web ./web

EXPOSE 8080
CMD ["./main"]
