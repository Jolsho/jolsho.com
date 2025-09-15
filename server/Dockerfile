FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .

# Build statically linked binary for Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Final minimal image
FROM alpine:3.18
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/dst ./dst

RUN chmod +x server && adduser -D webuser
CMD ["./server"]
