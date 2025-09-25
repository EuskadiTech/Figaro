# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o figaro ./cmd/figaro

# Final stage
FROM scratch

# Copy the binary from builder stage
COPY --from=builder /app/figaro /figaro

# Copy SSL certificates for HTTPS requests (if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8080

# Set default environment variables
ENV PORT=8080
ENV HOST=0.0.0.0
ENV DATA_DIR=/data

# Create volume for data persistence
VOLUME ["/data"]

# Run the binary
ENTRYPOINT ["/figaro"]