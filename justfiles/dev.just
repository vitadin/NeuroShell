# =============================================================================
# DEV MODULE - Development utilities, CI/CD, and project management
# =============================================================================

# =============================================================================
# DEVELOPMENT COMMANDS
# =============================================================================

# Run the application
run: build
    @echo "Running NeuroShell..."
    NEURO_LOG_LEVEL=debug ./bin/neuro

# Run development mode (rebuild on changes)
dev:
    @echo "Starting development mode..."
    @if command -v entr >/dev/null 2>&1; then \
        find . -name "*.go" | entr -r just run; \
    else \
        echo "Install 'entr' for file watching: brew install entr (macOS) or apt install entr (Linux)"; \
        just run; \
    fi

# Initialize development environment
init:
    @echo "Initializing development environment..."
    go mod download
    @mkdir -p bin docs logs sessions
    @echo "Development environment ready"

# Install binary to system PATH
install: build
    @echo "Installing NeuroShell..."
    @if [ -w "/usr/local/bin" ]; then \
        cp bin/neuro /usr/local/bin/neuro; \
        echo "Installed to /usr/local/bin/neuro"; \
    elif [ -n "$GOPATH" ] && [ -w "$GOPATH/bin" ]; then \
        cp bin/neuro $GOPATH/bin/neuro; \
        echo "Installed to $GOPATH/bin/neuro"; \
    else \
        echo "Cannot install: no writable directory in PATH found"; \
        echo "Try: sudo just install"; \
        exit 1; \
    fi

# Build neurotest binary
build-neurotest:
    @echo "Building neurotest..."
    go build -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"

# =============================================================================
# PROJECT MANAGEMENT
# =============================================================================

# Update dependencies
deps:
    @echo "Updating dependencies..."
    go mod tidy
    go mod download
    @echo "Dependencies updated"

# Generate documentation
docs:
    @echo "Generating documentation..."
    go doc -all > docs/api.txt
    @echo "Documentation generated: docs/api.txt"

# Check project health
check:
    @echo "Checking project health..."
    go mod verify
    go vet ./...
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fast; \
    fi
    @echo "Project health check complete"

# =============================================================================
# CI/CD COMMANDS
# =============================================================================

# Run all CI checks locally (fast version - avoids unnecessary rebuilds)
check-ci:
    @echo "Running CI checks locally (fast)..."
    @echo "1. Updating dependencies..."
    just deps
    @echo "2. Running linter..."
    just lint
    @echo "3. Running all unit tests..."
    just test-all-units
    @echo "4. Ensuring binaries are built..."
    just ensure-build
    @echo "5. Running end-to-end tests..."
    just test-e2e
    @echo "SUCCESS CI checks completed"

# Run all CI checks locally (clean version - full rebuild)
check-ci-clean:
    @echo "Running CI checks locally (clean)..."
    @echo "1. Updating dependencies..."
    just deps
    @echo "2. Running linter..."
    just lint
    @echo "3. Running all unit tests..."
    just test-all-units
    @echo "4. Building binary (clean)..."
    just build
    @echo "5. Running end-to-end tests..."
    just test-e2e
    @echo "SUCCESS CI checks completed"

# Run fast CI checks (skips linting and dependency updates)
check-ci-fast:
    @echo "Running fast CI checks..."
    @echo "1. Running all unit tests..."
    just test-all-units
    @echo "2. Ensuring binaries are built..."
    just ensure-build
    @echo "3. Running end-to-end tests..."
    just test-e2e
    @echo "SUCCESS Fast CI checks completed"