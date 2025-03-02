# Start from golang 1.19
FROM golang:1.19-alpine

# Install NATS Server and required build tools
RUN apk add --no-cache nats-server build-base

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o tilt-validator

# Expose ports for NATS and HTTP server
EXPOSE 4222 5000

# Create a script to start both NATS and the validator
COPY start.sh .
RUN chmod +x start.sh

# Set the entrypoint
ENTRYPOINT ["./start.sh"]