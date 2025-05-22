FROM golang:1.20-alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy and download Go dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/admin ./cmd/admin

# Use Python base image for the final stage
FROM python:3.10-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy Python requirements and install dependencies
COPY scripts/python/requirements.txt /app/scripts/python/
RUN pip install --no-cache-dir -r /app/scripts/python/requirements.txt
RUN pip install --no-cache-dir flash-attn optimum auto-gptq

# Copy Python code
COPY scripts/python/ /app/scripts/python/

# Copy binaries from Go builder
COPY --from=go-builder /app/bin/ /app/bin/

# Copy configuration and other files
COPY config/ /app/config/
COPY deployments/ /app/deployments/

# Make binaries executable
RUN chmod +x /app/bin/server /app/bin/worker /app/bin/admin

# Add binaries to PATH
ENV PATH="/app/bin:${PATH}"

# Set default command
CMD ["/app/bin/server"]
