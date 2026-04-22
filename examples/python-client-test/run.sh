#!/usr/bin/env bash
# run.sh - One-command runner for luner Python SDK example
# Supports: macOS / Linux / WSL

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info()    { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }

# Check dependencies
check_deps() {
  if ! command -v uv &> /dev/null; then
    error "uv not found. Please install: https://docs.astral.sh/uv/getting-started/installation/"
    exit 1
  fi
  if ! command -v curl &> /dev/null; then
    error "curl not found. Please install via your package manager."
    exit 1
  fi
}

# Check .env file
check_env() {
  if [[ ! -f ".env" ]]; then
    warn ".env not found. Creating from .env.example..."
    if [[ -f ".env.example" ]]; then
      cp .env.example .env
      info "Please edit .env to configure:"
      echo "  - AI_GATEWAY_BASE_URL (default: http://localhost:8080/v1)"
      echo "  - AI_GATEWAY_API_KEY  (placeholder value is fine)"
      echo "  - AI_GATEWAY_MODEL    (must match your config.yaml)"
      echo ""
      read -p "Press Enter after editing .env, or Ctrl+C to cancel..."
    else
      error ".env.example not found. Cannot proceed."
      exit 1
    fi
  fi
}

# Check gateway health
check_gateway() {
  info "Checking luner gateway health..."
  if ! curl -sf http://localhost:8080/health &> /dev/null; then
    warn "Gateway not responding at http://localhost:8080"
    echo ""
    echo "Please start luner first:"
    echo "  docker run -d --name luner -p 8080:8080 \\ "
    echo "    -v \$(pwd)/config/config.yaml:/app/config.yaml:ro \\ "
    echo "    --env-file .env \\ "
    echo "    ghcr.io/skylunna/luner:v0.4.2"
    echo ""
    read -p "Press Enter to continue anyway (test may fail), or Ctrl+C to cancel..."
  fi
}

# Run tests
run_tests() {
  info "Running integration tests with uv..."
  uv run python test/test_integration.py
}

# Main
main() {
  if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "One-command runner for luner Python SDK example."
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo ""
    echo "Steps performed:"
    echo "  1. Check dependencies (uv, curl)"
    echo "  2. Ensure .env exists (copy from .env.example if needed)"
    echo "  3. Verify luner gateway is running"
    echo "  4. Run integration tests via uv"
    exit 0
  fi

  info "🚀 Starting luner Python SDK example runner"
  check_deps
  check_env
  check_gateway
  run_tests
  info "✅ All done!"
}

main "$@"