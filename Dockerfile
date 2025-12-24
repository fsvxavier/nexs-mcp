# Multi-stage build for NEXS MCP Server with ONNX Runtime support
FROM ubuntu:22.04 AS builder

# Prevent interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies
RUN apt-get update && apt-get install -y \
    wget \
    tar \
    ca-certificates \
    build-essential \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.25
RUN wget -q https://go.dev/dl/go1.25.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz && \
    rm go1.25.0.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

# Install ONNX Runtime
RUN wget -q https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz && \
    tar -xzf onnxruntime-linux-x64-1.23.2.tgz && \
    cp -r onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/ && \
    cp -r onnxruntime-linux-x64-1.23.2/include/* /usr/local/include/ && \
    ldconfig && \
    rm -rf onnxruntime-linux-x64-1.23.2*

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with ONNX support
RUN CGO_ENABLED=1 \
    CGO_CFLAGS="-I/usr/local/include" \
    CGO_LDFLAGS="-L/usr/local/lib -lonnxruntime" \
    go build -ldflags="-w -s" \
    -o nexs-mcp \
    ./cmd/nexs-mcp

# Final stage
FROM ubuntu:22.04

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install ONNX Runtime (runtime only)
RUN wget -q https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz && \
    tar -xzf onnxruntime-linux-x64-1.23.2.tgz && \
    cp -r onnxruntime-linux-x64-1.23.2/lib/* /usr/local/lib/ && \
    ldconfig && \
    rm -rf onnxruntime-linux-x64-1.23.2*

# Create non-root user
RUN groupadd -g 1000 nexs && \
    useradd -u 1000 -g nexs -m -s /bin/bash nexs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/nexs-mcp .

# Create directories for models and data
RUN mkdir -p /app/models /app/data && \
    chown -R nexs:nexs /app

# Copy ONNX models from local build context
COPY --chown=nexs:nexs models/ms-marco-MiniLM-L-6-v2 /app/models/ms-marco-MiniLM-L-6-v2
COPY --chown=nexs:nexs models/paraphrase-multilingual-MiniLM-L12-v2 /app/models/paraphrase-multilingual-MiniLM-L12-v2

# Switch to non-root user
USER nexs

# Set environment variables for the application
ENV NEXS_STORAGE_TYPE=file \
    NEXS_DATA_DIR=/app/data \
    NEXS_SERVER_NAME=nexs-mcp \
    NEXS_LOG_LEVEL=info \
    NEXS_LOG_FORMAT=json \
    NEXS_AUTO_SAVE_MEMORIES=true \
    NEXS_AUTO_SAVE_INTERVAL=5m \
    NEXS_RESOURCES_ENABLED=false \
    NEXS_RESOURCES_CACHE_TTL=5m \
    NEXS_EMBEDDING_PROVIDER=onnx \
    NEXS_USE_GPU=false \
    NEXS_TEST_MODE=0 \
    HOME=/app \
    LD_LIBRARY_PATH=/usr/local/lib

# Performance tuning
ENV GOMAXPROCS=0 \
    GOMEMLIMIT=2GiB

# Expose volume for data persistence
VOLUME ["/app/data"]

# Health check (optional - uncomment if you add a health endpoint)
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD pgrep nexs-mcp || exit 1

# Run the application
ENTRYPOINT ["./nexs-mcp"]
CMD ["-storage", "file", "-data-dir", "/app/data", "-log-level", "info", "-log-format", "json"]
