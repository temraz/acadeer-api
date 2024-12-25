# Build stage
FROM golang:1.22-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Final stage
FROM alpine:latest

# Install ca-certificates and gettext (for envsubst)
RUN apk --no-cache add ca-certificates gettext

WORKDIR /root/

# Copy the binary and config template from builder
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml ./config.yaml.template

# Create entrypoint script with proper line endings
RUN printf '#!/bin/sh\nenvsubst < /root/config.yaml.template > /root/config.yaml\nexec "$@"\n' > /root/entrypoint.sh && \
    chmod +x /root/entrypoint.sh

# Expose port
EXPOSE 9090

# Set the entrypoint script
ENTRYPOINT ["/root/entrypoint.sh"]

# Command to run the application
CMD ["./main"] 