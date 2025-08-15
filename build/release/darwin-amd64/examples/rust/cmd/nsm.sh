#!/usr/bin/env bash
# NSM launcher for {{.ProjectName}}

set -euo pipefail

# Project configuration
PROJECT_TYPE="rust"
DOMAIN="{{.Domain}}"
COMMAND="cargo run"

# Colors
readonly BLUE='\033[0;34m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}●${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }

main() {
    log "🦀 Starting {{.ProjectName}} with NSM"
    
    # Check if NSM is available
    if ! command -v nsm >/dev/null 2>&1; then
        echo "❌ NSM not found. Please install NSM first:"
        echo "   Run 'nsm-setup install' to get started"
        exit 1
    fi
    
    # Check if Rust is available
    if ! command -v cargo >/dev/null 2>&1; then
        echo "❌ Rust/Cargo not found. Please install Rust first:"
        echo "   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh"
        exit 1
    fi
    
    # Check if cargo-watch is available (for hot reload)
    if command -v cargo-watch >/dev/null 2>&1; then
        COMMAND="cargo watch -x run"
        log "Using cargo-watch for hot reload"
    else
        warn "cargo-watch not found. Install for hot reload:"
        warn "  cargo install cargo-watch"
    fi
    
    # Create static directory if it doesn't exist
    [[ ! -d "static" ]] && mkdir -p static
    
    success "Configuration ready"
    echo "  Project: {{.ProjectName}}"
    echo "  Domain: {{.Domain}}"
    echo "  Framework: Rust + Axum"
    echo "  Hot Reload: $(command -v cargo-watch >/dev/null 2>&1 && echo "✅ Enabled" || echo "❌ Disabled")"
    echo
    
    # Start NSM
    exec nsm \
        --project-type "$PROJECT_TYPE" \
        --domain "$DOMAIN" \
        --command "$COMMAND"
}

# Only run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
