#!/usr/bin/env bash
# NSM launcher for {{.ProjectName}}

set -euo pipefail

# Project configuration
PROJECT_TYPE="vite"
DOMAIN="{{.Domain}}"
COMMAND="npm run dev"

# Colors
readonly BLUE='\033[0;34m'
readonly GREEN='\033[0;32m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}●${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }

main() {
    log "🚀 Starting {{.ProjectName}} with NSM"
    
    # Check if NSM is available
    if ! command -v nsm >/dev/null 2>&1; then
        echo "❌ NSM not found. Please install NSM first:"
        echo "   Run 'nsm-setup install' to get started"
        exit 1
    fi
    
    # Check if node_modules exists
    if [[ ! -d "node_modules" ]]; then
        log "Installing dependencies..."
        npm install
    fi
    
    success "Configuration ready"
    echo "  Project: {{.ProjectName}}"
    echo "  Domain: {{.Domain}}"
    echo "  Framework: React + Vite + TypeScript"
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
