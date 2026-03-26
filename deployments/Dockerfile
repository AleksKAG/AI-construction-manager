FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x  # ← -x для отладки

COPY . .
# Игнорируем ненужные файлы при копировании:
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web 2>/dev/null || true

EXPOSE 8080
CMD ["./main"]