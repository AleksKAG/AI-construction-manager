FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/api

FROM alpine:latest 
RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

RUN mkdir -p /app/data

ENV DB_PATH=/app/data/app.db

EXPOSE 8080

CMD ["./main"]
