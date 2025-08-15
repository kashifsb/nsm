#!/usr/bin/env bash
# NSM launcher for {{.ProjectName}}

set -euo pipefail

# Project configuration
PROJECT_TYPE="go"
DOMAIN="{{.Domain}}"
COMMAND="go run main.go"

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
    
    # Check if Go is available
    if ! command -v go >/dev/null 2>&1; then
        echo "❌ Go not found. Please install Go first."
        exit 1
    fi
    
    # Initialize Go module if needed
    if [[ ! -f "go.mod" ]]; then
        log "Initializing Go module..."
        go mod init {{.ProjectName}}
    fi
    
    success "Configuration ready"
    echo "  Project: {{.ProjectName}}"
    echo "  Domain: {{.Domain}}"
    echo "  Framework: Go Web Server"
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
