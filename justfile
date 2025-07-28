# NeuroShell Development Commands

# Default recipe to display available commands
default:
    @just --list
    @echo ""
    @echo "Test Commands:"
    @echo "  test              - Run all tests with coverage"
    @echo "  test-unit         - Run service/utils unit tests only"
    @echo "  test-commands     - Run command tests only"
    @echo "  test-parser       - Run parser tests only"
    @echo "  test-context      - Run context tests only"
    @echo "  test-shell        - Run shell tests only"
    @echo "  test-all-units    - Run all unit, command, parser, context, and shell tests"
    @echo "  test-e2e          - Run end-to-end tests"
    @echo "  record-all-e2e    - Re-record all end-to-end test cases"
    @echo "  test-bench        - Run benchmark tests"
    @echo ""
    @echo "CI/CD Commands:"
    @echo "  check-ci          - Run all CI checks locally (mirrors CI pipeline)"

# Build the main binaries
build: clean lint
    @echo "Building neurotest..."
    go build -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"
    @echo "Building NeuroShell..."
    go build -o bin/neuro ./cmd/neuro
    @echo "Binary built at: bin/neuro"


# Run the application
run: build
    @echo "Running NeuroShell..."
    NEURO_LOG_LEVEL=debug ./bin/neuro


# Run tests with coverage
test: build test-all-units
    @echo "Running tests..."
    EDITOR=echo go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Run unit tests only
test-unit:
    @echo "Running unit tests..."
    EDITOR=echo go test -v ./internal/services/... ./internal/testutils/...
    @echo "Unit tests complete"

# Run unit tests with coverage
test-unit-coverage:
    @echo "Running unit tests with coverage..."
    EDITOR=echo go test -v -coverprofile=unit-coverage.out ./internal/services/... ./internal/testutils/...
    go tool cover -html=unit-coverage.out -o unit-coverage.html
    go tool cover -func=unit-coverage.out
    @echo "Unit test coverage report generated: unit-coverage.html"

# Run command tests only
test-commands:
    @echo "Running command tests..."
    EDITOR=echo go test -v ./internal/commands/...
    @echo "Command tests complete"

# Run command tests with coverage
test-commands-coverage:
    @echo "Running command tests with coverage..."
    EDITOR=echo go test -v -coverprofile=commands-coverage.out ./internal/commands/...
    go tool cover -html=commands-coverage.out -o commands-coverage.html
    go tool cover -func=commands-coverage.out
    @echo "Command test coverage report generated: commands-coverage.html"

# Run parser tests only
test-parser:
    @echo "Running parser tests..."
    go test -v ./internal/parser/...
    @echo "Parser tests complete"

# Run parser tests with coverage
test-parser-coverage:
    @echo "Running parser tests with coverage..."
    go test -v -coverprofile=parser-coverage.out ./internal/parser/...
    go tool cover -html=parser-coverage.out -o parser-coverage.html
    go tool cover -func=parser-coverage.out
    @echo "Parser test coverage report generated: parser-coverage.html"

# Run context tests only
test-context:
    @echo "Running context tests..."
    go test -v ./internal/context/...
    @echo "Context tests complete"

# Run context tests with coverage
test-context-coverage:
    @echo "Running context tests with coverage..."
    go test -v -coverprofile=context-coverage.out ./internal/context/...
    go tool cover -html=context-coverage.out -o context-coverage.html
    go tool cover -func=context-coverage.out
    @echo "Context test coverage report generated: context-coverage.html"

# Run shell tests only
test-shell:
    @echo "Running shell tests..."
    go test -v ./internal/shell/...
    @echo "Shell tests complete"

# Run shell tests with coverage
test-shell-coverage:
    @echo "Running shell tests with coverage..."
    go test -v -coverprofile=shell-coverage.out ./internal/shell/...
    go tool cover -html=shell-coverage.out -o shell-coverage.html
    go tool cover -func=shell-coverage.out
    @echo "Shell test coverage report generated: shell-coverage.html"

# Run all unit, command, parser, context, and shell tests
test-all-units:
    @echo "Running all unit, command, parser, context, execution, and shell tests..."
    EDITOR=echo go test -v \
        ./internal/services/... \
        ./internal/testutils/... \
        ./internal/parser/... \
        ./internal/context/... \
        ./internal/statemachine/... \
        ./internal/shell/... \
        ./internal/stringprocessing/... \
        ./internal/commands/builtin/... \
        ./internal/commands/render/... \
        ./internal/commands/session/... \
        ./internal/commands/model/... \
        ./internal/commands/provider/... \
        ./internal/commands/llm/...
    # ./internal/commands/... # Commented out during state machine transition (except specific integrated commands)
    @echo "All unit, command, parser, context, execution, and shell tests complete"

# Run all unit, command, parser, context, and shell tests with coverage
test-all-units-coverage:
    @echo "Running all unit, command, parser, context, execution, and shell tests with coverage..."
    EDITOR=echo go test -v -coverprofile=all-units-coverage.out \
        ./internal/services/... \
        ./internal/testutils/... \
        ./internal/parser/... \
        ./internal/context/... \
        ./internal/statemachine/... \
        ./internal/shell/... \
        ./internal/stringprocessing/... \
        ./internal/commands/builtin/... \
        ./internal/commands/render/... \
        ./internal/commands/session/... \
        ./internal/commands/model/... \
        ./internal/commands/provider/... \
        ./internal/commands/llm/...
    # ./internal/commands/... # Commented out during state machine transition (except specific integrated commands)
    go tool cover -html=all-units-coverage.out -o all-units-coverage.html
    go tool cover -func=all-units-coverage.out
    @echo "All unit test coverage report generated: all-units-coverage.html"

# Run benchmark tests
test-bench:
    @echo "Running benchmark tests..."
    go test -bench=. -benchmem ./internal/services/... ./internal/commands/... ./internal/parser/... ./internal/context/... ./internal/statemachine/... ./internal/shell/...
    @echo "Benchmark tests complete"

# Check test coverage percentage
test-coverage-check:
    @echo "Checking test coverage..."
    @coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
    if [ -z "$$coverage" ]; then \
        echo "No coverage data found. Run 'just test' first."; \
        exit 1; \
    fi; \
    echo "Current coverage: $$coverage%"; \
    if [ $$(echo "$$coverage >= 90" | bc) -eq 1 ]; then \
        echo "✅ Coverage meets target (≥90%)"; \
    else \
        echo "❌ Coverage below target ($$coverage% < 90%)"; \
        exit 1; \
    fi

# Run linting and formatting
lint:
    @echo "Running linters..."
    gofmt -s -w .
    go vet ./...
    @echo "Running golangci-lint..."
    @if ! command -v golangci-lint >/dev/null 2>&1; then \
        echo "❌ golangci-lint not found. Please install it:"; \
        echo "   brew install golangci-lint"; \
        exit 1; \
    fi
    golangci-lint run

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

# Run all CI checks locally (mirrors CI pipeline)
check-ci:
    @echo "Running CI checks locally..."
    @echo "1. Updating dependencies..."
    just deps
    @echo "2. Running linter..."
    just lint
    @echo "3. Running all unit tests..."
    just test-all-units
    @echo "4. Building binary..."
    just build
    @echo "5. Running end-to-end tests..."
    just test-e2e
    @echo "✅ CI checks completed"

# Initialize development environment
init:
    @echo "Initializing development environment..."
    go mod download
    @mkdir -p bin docs logs sessions
    @echo "Development environment ready"

# Run end-to-end tests
test-e2e: build
    @echo "Running end-to-end tests..."
    ./bin/neurotest --neuro-cmd="./bin/neuro" run-all
    @echo "End-to-end tests complete"

# Re-record all end-to-end test cases
record-all-e2e: build
    #!/bin/bash
    echo "Re-recording all end-to-end test cases..."
    for test_file in $(find test/golden -maxdepth 1 -name "*.neuro" -type f | sort); do \
        test_name=$(basename "$test_file" .neuro); \
        echo "Recording $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" record "$test_name" >/dev/null 2>&1 && echo "✓ $test_name" || echo "✗ $test_name"; \
    done
    echo "All end-to-end test cases re-recorded"

# Build neurotest binary
build-neurotest:
    @echo "Building neurotest..."
    go build -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"
