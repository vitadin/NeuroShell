# =============================================================================
# BUILD MODULE - Build, format, lint, and cleanup recipes
# =============================================================================

# Version and build information
_version := `./scripts/version.sh`
_git_commit := `git rev-parse --short HEAD 2>/dev/null || echo unknown`
_build_date := `date -u +%Y-%m-%d`

# Build flags
_build_flags := '-ldflags="-X neuroshell/internal/version.Version=' + _version + ' -X neuroshell/internal/version.GitCommit=' + _git_commit + ' -X neuroshell/internal/version.BuildDate=' + _build_date + '"'

# Build the main binaries with version injection (full clean build)
build: clean lint
    @echo "Building neurotest..."
    go build {{_build_flags}} -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"
    @echo "Building NeuroShell..."
    go build {{_build_flags}} -o bin/neuro ./cmd/neuro
    @echo "Binary built at: bin/neuro"

# Fast development build (preserves Go cache, full linting for code quality)
build-fast: clean-artifacts lint
    @echo "Fast building neurotest..."
    go build {{_build_flags}} -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"
    @echo "Fast building NeuroShell..."
    go build {{_build_flags}} -o bin/neuro ./cmd/neuro
    @echo "Binary built at: bin/neuro"

# Ultra-fast development build (preserves cache, skip all validation - use sparingly!)
build-dev: clean-artifacts
    @echo "⚡ Ultra-fast building (skipping validation)..."
    go build {{_build_flags}} -o bin/neurotest ./cmd/neurotest
    go build {{_build_flags}} -o bin/neuro ./cmd/neuro
    @echo "⚠️  Remember to run 'just lint' before committing!"

# Parallel build (build both binaries simultaneously)
build-parallel: clean lint
    #!/bin/bash
    set -euo pipefail
    echo "Building binaries in parallel..."
    mkdir -p bin

    # Build both binaries in parallel
    (
        echo "Building neurotest..."
        go build {{_build_flags}} -o bin/neurotest ./cmd/neurotest
        echo "Binary built at: bin/neurotest"
    ) &

    (
        echo "Building NeuroShell..."
        go build {{_build_flags}} -o bin/neuro ./cmd/neuro
        echo "Binary built at: bin/neuro"
    ) &

    # Wait for both builds to complete
    wait
    echo "Parallel build complete"

# Build binaries only if they don't exist or sources are newer
build-if-needed:
    #!/bin/bash
    set -euo pipefail
    
    # Function to check if binary needs rebuilding
    needs_rebuild() {
        local binary="$1"
        local source_dir="$2"
        
        # If binary doesn't exist, rebuild
        if [ ! -f "$binary" ]; then
            return 0
        fi
        
        # Check if any source files are newer than the binary
        if find "$source_dir" -name "*.go" -newer "$binary" | grep -q .; then
            return 0
        fi
        
        # Check if go.mod or go.sum are newer
        if [ -f "go.mod" ] && [ "go.mod" -nt "$binary" ]; then
            return 0
        fi
        if [ -f "go.sum" ] && [ "go.sum" -nt "$binary" ]; then
            return 0
        fi
        
        return 1
    }
    
    # Ensure bin directory exists
    mkdir -p bin
    
    # Build neurotest if needed
    if needs_rebuild "bin/neurotest" "cmd/neurotest"; then
        echo "Building neurotest..."
        go build {{_build_flags}} -o bin/neurotest ./cmd/neurotest
        echo "Binary built at: bin/neurotest"
    else
        echo "neurotest is up to date"
    fi
    
    # Build neuro if needed
    if needs_rebuild "bin/neuro" "cmd/neuro"; then
        echo "Building NeuroShell..."
        go build {{_build_flags}} -o bin/neuro ./cmd/neuro
        echo "Binary built at: bin/neuro"
    else
        echo "neuro is up to date"    
    fi

# Ensure binaries are built (build if needed, but skip clean and lint)
ensure-build: build-if-needed

# Build for multiple platforms with version injection
build-all:
    #!/bin/bash
    set -euo pipefail
    echo "Building for multiple platforms..."
    
    # Get version info
    VERSION=$(./scripts/version.sh)
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    BUILD_DATE=$(date -u +%Y-%m-%d)
    
    LDFLAGS="-X 'neuroshell/internal/version.Version=${VERSION}' -X 'neuroshell/internal/version.GitCommit=${GIT_COMMIT}' -X 'neuroshell/internal/version.BuildDate=${BUILD_DATE}'"
    
    echo "Building for Linux amd64..."
    GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/neuro-linux-amd64 ./cmd/neuro
    
    echo "Building for macOS amd64..."
    GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/neuro-darwin-amd64 ./cmd/neuro
    
    echo "Building for macOS arm64..."
    GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o bin/neuro-darwin-arm64 ./cmd/neuro
    
    echo "Building for Windows amd64..."
    GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o bin/neuro-windows-amd64.exe ./cmd/neuro
    
    echo "Cross-platform binaries built in bin/"

# Format Go code and organize imports
format:
    @echo "Formatting Go code and organizing imports..."
    @if ! command -v goimports >/dev/null 2>&1; then \
        echo "Installing goimports..."; \
        go install golang.org/x/tools/cmd/goimports@latest; \
    fi
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    elif [ -f "$(go env GOPATH)/bin/goimports" ]; then \
        $(go env GOPATH)/bin/goimports -w .; \
    elif [ -f "$HOME/go/bin/goimports" ]; then \
        $HOME/go/bin/goimports -w .; \
    else \
        echo "ERROR goimports not found after installation"; \
        exit 1; \
    fi
    gofmt -s -w .
    @echo "Code formatting complete"

# Organize imports only (without other formatting)
imports:
    @echo "Organizing Go imports..."
    @if ! command -v goimports >/dev/null 2>&1; then \
        echo "Installing goimports..."; \
        go install golang.org/x/tools/cmd/goimports@latest; \
    fi
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    elif [ -f "$(go env GOPATH)/bin/goimports" ]; then \
        $(go env GOPATH)/bin/goimports -w .; \
    elif [ -f "$HOME/go/bin/goimports" ]; then \
        $HOME/go/bin/goimports -w .; \
    else \
        echo "ERROR goimports not found after installation"; \
        exit 1; \
    fi
    @echo "Import organization complete"

# Run linting and formatting
lint:
    @echo "Running linters..."
    just format
    go vet ./...
    @echo "Running golangci-lint..."
    @if ! command -v golangci-lint >/dev/null 2>&1; then \
        echo "ERROR golangci-lint not found. Please install it:"; \
        echo "   brew install golangci-lint"; \
        exit 1; \
    fi
    golangci-lint run

# Clean build artifacts and temporary files (full clean including Go cache)
clean:
    @echo "Cleaning up..."
    rm -rf bin/
    rm -rf dist/
    rm -f coverage.out coverage.html
    rm -f *.prof
    go clean -cache
    @echo "Clean complete"

# Clean only build artifacts (preserves Go cache for faster rebuilds)
clean-artifacts:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -rf dist/
    rm -f coverage.out coverage.html
    rm -f *.prof
    @echo "Artifact cleanup complete"

# Quick format (skip import organization for speed)
format-quick:
    @echo "Quick formatting Go code..."
    gofmt -s -w .
    @echo "Quick formatting complete"