#!/usr/bin/env bash
# NSM launcher for {{.ProjectName}}

set -euo pipefail

# Project configuration
PROJECT_TYPE="java"
DOMAIN="{{.Domain}}"
COMMAND="mvn spring-boot:run"

# Colors
readonly BLUE='\033[0;34m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m'

log() { echo -e "${BLUE}●${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC} $*"; }

main() {
    log "☕ Starting {{.ProjectName}} with NSM"
    
    # Check if NSM is available
    if ! command -v nsm >/dev/null 2>&1; then
        echo "❌ NSM not found. Please install NSM first:"
        echo "   Run 'nsm-setup install' to get started"
        exit 1
    fi
    
    # Check if Java is available
    if ! command -v java >/dev/null 2>&1; then
        echo "❌ Java not found. Please install Java 17+ first."
        exit 1
    fi
    
    # Check Java version
    JAVA_VERSION=$(java -version 2>&1 | head -n1 | cut -d'"' -f2 | cut -d'.' -f1)
    if [[ "$JAVA_VERSION" -lt 17 ]]; then
        warn "Java $JAVA_VERSION detected. Java 17+ recommended for Spring Boot 3."
    fi
    
    # Check if Maven is available
    if ! command -v mvn >/dev/null 2>&1; then
        echo "❌ Maven not found. Please install Maven first:"
        echo "   brew install maven  # macOS"
        echo "   apt install maven   # Ubuntu/Debian"
        exit 1
    fi
    
    # Check if pom.xml exists
    if [[ ! -f "pom.xml" ]]; then
        echo "❌ pom.xml not found. Are you in a Maven project directory?"
        exit 1
    fi
    
    # Check if we can use Spring Boot DevTools for hot reload
    if grep -q "spring-boot-devtools" pom.xml; then
        log "Spring Boot DevTools detected - hot reload enabled"
    else
        warn "Spring Boot DevTools not found. Add it to pom.xml for hot reload."
    fi
    
    success "Configuration ready"
    echo "  Project: {{.ProjectName}}"
    echo "  Domain: {{.Domain}}"
    echo "  Framework: Java + Spring Boot"
    echo "  Java: $(java -version 2>&1 | head -n1 | cut -d'"' -f2)"
    echo "  Maven: $(mvn --version | head -n1 | cut -d' ' -f3)"
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
