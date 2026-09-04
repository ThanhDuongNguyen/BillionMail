#!/bin/bash
# ============================================================
# BillionMail — quick backend dev reload (local)
#
# Rebuilds the Go binary for the running core container's
# architecture, copies it in, and restarts the core process.
# Much faster than rebuilding the whole Docker image.
#
# Usage:
#   bash dev-backend.sh            # build + swap + restart core
#   bash dev-backend.sh --logs     # same, then tail core logs
#   bash dev-backend.sh --up       # start the stack first if it's down
#   bash dev-backend.sh --up --logs
#
# Env:
#   CORE_IMAGE   core image tag used by docker compose
#                (default: billionmail/core:dev-variables)
# ============================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

CORE_SERVICE="core-billionmail"
GO_IMAGE="golang:1.22-alpine"
CONTAINER_BINARY_PATH="/opt/billionmail/core/billionmail"

# docker compose needs CORE_IMAGE to resolve the core service.
export CORE_IMAGE="${CORE_IMAGE:-billionmail/core:dev-variables}"

# Parse flags (order-independent).
DO_UP=false
DO_LOGS=false
for arg in "$@"; do
	case "$arg" in
		--up) DO_UP=true ;;
		--logs) DO_LOGS=true ;;
		*) ;;
	esac
done

log() { printf '\033[1;34m[dev-backend]\033[0m %s\n' "$*"; }
err() { printf '\033[1;31m[dev-backend]\033[0m %s\n' "$*" >&2; }

# ---- 1. Ensure the core container is running -------------------------------
CORE_CONTAINER="$(docker compose ps -q "$CORE_SERVICE" 2>/dev/null || true)"

if [ -z "$CORE_CONTAINER" ] && [ "$DO_UP" = true ]; then
	log "Core not running — starting the stack (CORE_IMAGE=$CORE_IMAGE)..."
	docker compose up -d
	sleep 8
	CORE_CONTAINER="$(docker compose ps -q "$CORE_SERVICE" 2>/dev/null || true)"
fi

if [ -z "$CORE_CONTAINER" ]; then
	err "Core container is not running."
	err "Start the stack first:"
	err "  CORE_IMAGE=$CORE_IMAGE docker compose up -d"
	err "Or re-run with --up to start it automatically:"
	err "  bash dev-backend.sh --up"
	exit 1
fi

# ---- 2. Detect the container's architecture --------------------------------
# Map uname -m output to Go's GOARCH values.
CONTAINER_ARCH="$(docker exec "$CORE_CONTAINER" uname -m 2>/dev/null || echo "unknown")"
case "$CONTAINER_ARCH" in
	x86_64) GOARCH="amd64" ;;
	aarch64 | arm64) GOARCH="arm64" ;;
	*)
		err "Unsupported/unknown container arch: $CONTAINER_ARCH"
		exit 1
		;;
esac
log "Core container arch: $CONTAINER_ARCH -> GOARCH=$GOARCH"

# ---- 3. Build the Go binary (Linux) inside a throwaway builder -------------
# A named volume caches Go modules so repeat builds are fast.
BINARY_NAME="billionmail-dev-${GOARCH}"
log "Building Go binary ($BINARY_NAME)..."
docker run --rm \
	-v "$ROOT_DIR/core":/opt/core \
	-v billionmail_gocache:/go/pkg/mod \
	-w /opt/core \
	"$GO_IMAGE" \
	sh -c "GOOS=linux GOARCH=${GOARCH} go build -ldflags='-s -w' -o ${BINARY_NAME} main.go"

if [ ! -f "core/${BINARY_NAME}" ]; then
	err "Build failed: core/${BINARY_NAME} not found."
	exit 1
fi
log "Build OK ($(du -h "core/${BINARY_NAME}" | cut -f1))."

# Ensure the temp binary is removed even if later steps fail.
cleanup() { rm -f "$ROOT_DIR/core/${BINARY_NAME}"; }
trap cleanup EXIT

# ---- 4. Copy the binary into the running container -------------------------
log "Copying binary into container..."
docker cp "core/${BINARY_NAME}" "${CORE_CONTAINER}:${CONTAINER_BINARY_PATH}"
docker exec "$CORE_CONTAINER" chmod +x "$CONTAINER_BINARY_PATH"

# ---- 5. Restart the core process via supervisor ----------------------------
log "Restarting core process..."
docker exec "$CORE_CONTAINER" supervisorctl restart core >/dev/null

log "Done. Core reloaded with the latest backend code."

# ---- 6. Optionally tail logs ----------------------------------------------
if [ "$DO_LOGS" = true ]; then
	log "Tailing core logs (Ctrl+C to stop)..."
	docker compose logs -f "$CORE_SERVICE"
fi
