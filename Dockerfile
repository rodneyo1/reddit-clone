# Stage 1: Build the application
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

# Enable CGO
ENV CGO_ENABLED=1

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o forum .

# Stage 2: Create the final lightweight image
FROM alpine:latest
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/forum .

# Copy the templates directory from the builder stage
COPY --from=builder /app/templates /app/templates

# Copy the static directory from the builder stage
COPY --from=builder /app/static /app/static

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./forum"]
