# Build Stage (Debian-based golang image for full toolchain & glibc compatibility)
FROM golang:latest AS builder

WORKDIR /app

# Install git and CA certificates
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/api/main.go

# Production Stage (Lightweight Alpine container)
FROM alpine:latest

WORKDIR /app

# Install CA certificates and timezone data for external API calls
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/server /app/server

# Create uploads folder
RUN mkdir -p /app/uploads

# Expose port
EXPOSE 8080

# Environment variables default fallbacks
ENV PORT=8080
ENV UPLOADS_DIR=/app/uploads

CMD ["/app/server"]
