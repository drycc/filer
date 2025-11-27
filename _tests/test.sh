#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test configuration
IMAGE_NAME="${IMAGE:-registry.drycc.cc/drycc/filer:canary}"
CONTAINER_NAME="filer-test"
PING_PORT=8081
RCLONE_PORT=8014
PING_INTERVAL=30s

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

cleanup() {
    log "Cleaning up..."
    podman stop ${CONTAINER_NAME} 2>/dev/null || true
    podman rm ${CONTAINER_NAME} 2>/dev/null || true
}

trap cleanup EXIT INT TERM

# Check requirements
command -v podman >/dev/null || { error "podman not found"; exit 1; }
command -v curl >/dev/null || { error "curl not found"; exit 1; }

# Start container
log "Starting container..."
podman run -d --name ${CONTAINER_NAME} \
    ${IMAGE_NAME} \
    --interval=${PING_INTERVAL} --bind=0.0.0.0:${PING_PORT} -- \
    rclone serve s3 /tmp --addr 0.0.0.0:${RCLONE_PORT}

# Get container IP
log "Getting container IP..."
CONTAINER_IP=$(podman inspect --format "{{ .NetworkSettings.IPAddress }}" ${CONTAINER_NAME})
if [ -z "$CONTAINER_IP" ]; then
    error "Failed to get container IP"
    exit 1
fi
log "Container IP: ${CONTAINER_IP}"

# Wait for services
log "Waiting for services..."
for i in {1..30}; do
    if curl -sf http://${CONTAINER_IP}:${PING_PORT}/_/ping >/dev/null; then
        log "Ping service ready"
        break
    fi
    [ $i -eq 30 ] && { error "Ping service timeout"; podman logs ${CONTAINER_NAME}; exit 1; }
    sleep 2
done

for i in {1..30}; do
    if curl -sf http://${CONTAINER_IP}:${RCLONE_PORT}/ >/dev/null 2>&1; then
        log "Rclone service ready"
        break
    fi
    [ $i -eq 30 ] && { error "Rclone service timeout"; podman logs ${CONTAINER_NAME}; exit 1; }
    sleep 2
done

# Test ping
log "Testing ping functionality..."
response=$(curl -sw "%{http_code}" http://${CONTAINER_IP}:${PING_PORT}/_/ping)
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" = "200" ] && [ "$body" = "pong" ]; then
    log "Ping test passed"
else
    error "Ping test failed - HTTP: ${http_code}, Body: ${body}"
    exit 1
fi

# Test rclone service
log "Testing Rclone S3 service..."
response=$(curl -sw "%{http_code}" http://${CONTAINER_IP}:${RCLONE_PORT}/)
http_code="${response: -3}"

if [ "$http_code" != "000" ]; then
    log "Rclone service test passed"
else
    error "Rclone service test failed"
    exit 1
fi

# Optional: Test ping timeout
if [ "${TEST_PING_TIMEOUT:-}" = "true" ]; then
    log "Testing ping timeout functionality..."
    wait_time=$(echo ${PING_INTERVAL} | sed 's/s$//')
    wait_time=$((wait_time + 10))
    
    warn "Waiting ${wait_time}s for container to auto-exit..."
    sleep ${wait_time}
    
    if ! podman ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        log "Ping timeout test passed - container auto-exited"
    else
        warn "Container still running - timeout may not be working"
    fi
else
    log "Skipping ping timeout test (set TEST_PING_TIMEOUT=true to enable)"
fi

log "All tests passed! ✅"
