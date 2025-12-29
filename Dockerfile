# Build stage
FROM golang:1.23-alpine AS builder

# Set Go toolchain to auto to allow downloading required version
ENV GOTOOLCHAIN=auto

WORKDIR /build

# Copy all services (needed for local replace directives)
COPY . .

# Set working directory to this service
WORKDIR /build/ecommerce-go-inventory-service

# Download dependencies
RUN go mod download

# Build the application
# This handles both cmd/api/main.go and main.go structures
RUN if [ -d "cmd/api" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .; \
    fi

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /build/ecommerce-go-inventory-service/main .

# Copy config files if they exist
COPY --from=builder /build/ecommerce-go-inventory-service/config* ./config/ 2>/dev/null || true

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
