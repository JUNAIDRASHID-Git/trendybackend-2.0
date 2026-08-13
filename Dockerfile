# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools and CA certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/api/main.go

# Production Stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates and timezone data for external API calls (Gemini, Zoho)
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
