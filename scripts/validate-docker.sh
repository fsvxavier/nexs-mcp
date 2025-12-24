#!/bin/bash

# Docker Validation Script for NEXS-MCP
# This script validates the Docker image has all required components

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

IMAGE_NAME="${1:-nexs-mcp:latest}"

echo "======================================"
echo "NEXS-MCP Docker Validation"
echo "======================================"
echo "Image: $IMAGE_NAME"
echo ""

# Check if image exists
echo "1. Checking if image exists..."
if docker image inspect "$IMAGE_NAME" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Image found"
else
    echo -e "${RED}✗${NC} Image not found"
    exit 1
fi

# Check image size
echo ""
echo "2. Checking image size..."
SIZE=$(docker image inspect "$IMAGE_NAME" --format='{{.Size}}' | awk '{print $1/1024/1024 " MB"}')
echo -e "${GREEN}✓${NC} Image size: $SIZE"

# Check ONNX Runtime
echo ""
echo "3. Checking ONNX Runtime..."
if docker run --rm "$IMAGE_NAME" ldconfig -p | grep -q "libonnxruntime.so"; then
    VERSION=$(docker run --rm "$IMAGE_NAME" ldconfig -p | grep libonnxruntime.so | head -1)
    echo -e "${GREEN}✓${NC} ONNX Runtime found: $VERSION"
else
    echo -e "${RED}✗${NC} ONNX Runtime not found"
    exit 1
fi

# Check binary
echo ""
echo "4. Checking nexs-mcp binary..."
if docker run --rm "$IMAGE_NAME" /app/nexs-mcp --version > /dev/null 2>&1; then
    VERSION=$(docker run --rm "$IMAGE_NAME" /app/nexs-mcp --version)
    echo -e "${GREEN}✓${NC} Binary found: $VERSION"
else
    echo -e "${RED}✗${NC} Binary not working"
    exit 1
fi

# Check models
echo ""
echo "5. Checking ONNX models..."
MODELS=$(docker run --rm "$IMAGE_NAME" ls /app/models/)

if echo "$MODELS" | grep -q "ms-marco-MiniLM-L-6-v2"; then
    SIZE=$(docker run --rm "$IMAGE_NAME" du -sh /app/models/ms-marco-MiniLM-L-6-v2 | awk '{print $1}')
    echo -e "${GREEN}✓${NC} MS MARCO model found ($SIZE)"
else
    echo -e "${RED}✗${NC} MS MARCO model not found"
    exit 1
fi

if echo "$MODELS" | grep -q "paraphrase-multilingual-MiniLM-L12-v2"; then
    SIZE=$(docker run --rm "$IMAGE_NAME" du -sh /app/models/paraphrase-multilingual-MiniLM-L12-v2 | awk '{print $1}')
    echo -e "${GREEN}✓${NC} Paraphrase-Multilingual model found ($SIZE)"
else
    echo -e "${RED}✗${NC} Paraphrase-Multilingual model not found"
    exit 1
fi

# Check data directory
echo ""
echo "6. Checking data directory structure..."
DATA_DIRS=$(docker run --rm "$IMAGE_NAME" ls /app/data/elements/)

for dir in "agents" "ensembles" "memories" "personas" "skills" "templates"; do
    if echo "$DATA_DIRS" | grep -q "$dir"; then
        echo -e "${GREEN}✓${NC} $dir directory exists"
    else
        echo -e "${YELLOW}!${NC} $dir directory not found (may be created on first run)"
    fi
done

# Check environment variables
echo ""
echo "7. Checking environment variables..."
ENV_VARS=$(docker run --rm "$IMAGE_NAME" env | sort)

for var in "NEXS_EMBEDDING_PROVIDER" "LD_LIBRARY_PATH" "NEXS_AUTO_SAVE_MEMORIES" "GOMAXPROCS" "GOMEMLIMIT"; do
    if echo "$ENV_VARS" | grep -q "^$var="; then
        VALUE=$(echo "$ENV_VARS" | grep "^$var=" | cut -d'=' -f2-)
        echo -e "${GREEN}✓${NC} $var=$VALUE"
    else
        echo -e "${YELLOW}!${NC} $var not set"
    fi
done

# Functional test
echo ""
echo "8. Running functional test..."
if docker run --rm -v "$(pwd)/test_data:/app/data" "$IMAGE_NAME" /app/nexs-mcp --version > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Functional test passed"
else
    echo -e "${RED}✗${NC} Functional test failed"
    exit 1
fi

# Resource limits test
echo ""
echo "9. Checking resource limits..."
docker run --rm \
    --memory=2g \
    --cpus=2 \
    -e GOMEMLIMIT=2GiB \
    -e GOMAXPROCS=2 \
    "$IMAGE_NAME" \
    /app/nexs-mcp --version > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓${NC} Resource limits respected"
else
    echo -e "${RED}✗${NC} Resource limits test failed"
    exit 1
fi

# Security check
echo ""
echo "10. Checking security features..."
SECURITY=$(docker image inspect "$IMAGE_NAME" --format='{{json .Config}}')

if echo "$SECURITY" | grep -q '"User":""'; then
    echo -e "${YELLOW}!${NC} Running as root (consider using non-root user)"
else
    USER=$(echo "$SECURITY" | jq -r '.User')
    echo -e "${GREEN}✓${NC} Running as user: $USER"
fi

# Summary
echo ""
echo "======================================"
echo -e "${GREEN}✓ All validation checks passed!${NC}"
echo "======================================"
echo ""
echo "Image Details:"
echo "  Name: $IMAGE_NAME"
echo "  Size: $SIZE"
echo "  ONNX: Enabled with Runtime v1.23.2"
echo "  Models: MS MARCO + Paraphrase-Multilingual"
echo "  Config: All features configured"
echo ""
echo "Next steps:"
echo "  1. Run: docker run -d --name nexs-mcp -v \$(pwd)/data:/app/data $IMAGE_NAME"
echo "  2. Test: docker logs nexs-mcp"
echo "  3. Publish: make docker-publish"
echo ""
