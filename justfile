# NeuroShell Development Commands

# Default recipe to display available commands
default:
    @just --list

# Build the main binary
build:
    @echo "Building NeuroShell..."
    go build -o bin/neuro ./cmd/neuro
    @echo "Binary built at: bin/neuro"

# Run the application
run: build
    @echo "Running NeuroShell..."
    NEURO_LOG_LEVEL=debug ./bin/neuro

# Run tests with coverage
test:
    @echo "Running tests..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run linting and formatting
lint:
    @echo "Running linters..."
    gofmt -s -w .
    go vet ./...
    @if command -v golangci-lint >/dev/null 2>&1; then \
        echo "Running golangci-lint..."; \
        golangci-lint run; \
    else \
        echo "golangci-lint not found, skipping advanced linting"; \
    fi

# Clean build artifacts and temporary files
clean:
    @echo "Cleaning up..."
    rm -rf bin/
    rm -rf dist/
    rm -f coverage.out coverage.html
    rm -f *.prof
    go clean -cache
    @echo "Clean complete"

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

# Run development mode (rebuild on changes)
dev:
    @echo "Starting development mode..."
    @if command -v entr >/dev/null 2>&1; then \
        find . -name "*.go" | entr -r just run; \
    else \
        echo "Install 'entr' for file watching: brew install entr (macOS) or apt install entr (Linux)"; \
        just run; \
    fi

# Build for multiple platforms
build-all:
    @echo "Building for multiple platforms..."
    GOOS=linux GOARCH=amd64 go build -o bin/neuro-linux-amd64 ./cmd/neuro
    GOOS=darwin GOARCH=amd64 go build -o bin/neuro-darwin-amd64 ./cmd/neuro
    GOOS=darwin GOARCH=arm64 go build -o bin/neuro-darwin-arm64 ./cmd/neuro
    GOOS=windows GOARCH=amd64 go build -o bin/neuro-windows-amd64.exe ./cmd/neuro
    @echo "Cross-platform binaries built in bin/"

# Check project health
check:
    @echo "Checking project health..."
    go mod verify
    go vet ./...
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fast; \
    fi
    @echo "Project health check complete"

# Initialize development environment
init:
    @echo "Initializing development environment..."
    go mod download
    @mkdir -p bin docs logs sessions
    @echo "Development environment ready"
