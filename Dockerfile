# --- Stage 1: Build the Go binary ---
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build for Linux ARM64 (your Mac is arm64)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server ./cmd/api

# --- Stage 2: Runtime (Debian instead of Alpine) ---
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/server .

#RUN chmod +x /app/server

EXPOSE 8080

CMD ["./server"]
