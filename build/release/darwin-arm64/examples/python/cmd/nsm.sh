#!/usr/bin/env bash
# NSM launcher for {{.ProjectName}}

set -euo pipefail

# Project configuration
PROJECT_TYPE="python"
DOMAIN="{{.Domain}}"
COMMAND="python app.py"

# Colors
readonly BLUE='\033[0;34m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}●${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }

main() {
    log "🐍 Starting {{.ProjectName}} with NSM"
    
    # Check if NSM is available
    if ! command -v nsm >/dev/null 2>&1; then
        echo "❌ NSM not found. Please install NSM first:"
        echo "   Run 'nsm-setup install' to get started"
        exit 1
    fi
    
    # Check if Python is available
    if ! command -v python3 >/dev/null 2>&1 && ! command -v python >/dev/null 2>&1; then
        echo "❌ Python not found. Please install Python first."
        exit 1
    fi
    
    # Use python3 if available, otherwise python
    PYTHON_CMD="python3"
    if ! command -v python3 >/dev/null 2>&1; then
        PYTHON_CMD="python"
    fi
    
    # Check for virtual environment
    if [[ ! -d "venv" && ! -n "${VIRTUAL_ENV:-}" ]]; then
        warn "No virtual environment detected. Creating one..."
        $PYTHON_CMD -m venv venv
        log "Virtual environment created. Activating..."
        source venv/bin/activate
    elif [[ -d "venv" && ! -n "${VIRTUAL_ENV:-}" ]]; then
        log "Activating existing virtual environment..."
        source venv/bin/activate
    fi
    
    # Install dependencies if requirements.txt exists
    if [[ -f "requirements.txt" ]]; then
        log "Installing dependencies..."
        pip install -q -r requirements.txt
    fi
    
    # Create static directory if it doesn't exist
    [[ ! -d "static" ]] && mkdir -p static
    [[ ! -d "templates" ]] && mkdir -p templates
    
    # Update command with correct Python executable
    COMMAND="$PYTHON_CMD app.py"
    
    success "Configuration ready"
    echo "  Project: {{.ProjectName}}"
    echo "  Domain: {{.Domain}}"
    echo "  Framework: Python + Flask"
    echo "  Python: $($PYTHON_CMD --version)"
    echo "  Virtual Env: ${VIRTUAL_ENV:-None}"
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
