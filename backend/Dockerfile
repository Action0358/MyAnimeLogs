# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy the source code
COPY . .

# Download dependencies and build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /main .

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./main"]
