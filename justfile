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
    @echo "  test-e2e          - Run end-to-end tests (includes .neurorc tests)"
    @echo "  test-neurorc      - Run .neurorc startup tests only"
    @echo "  record-all-e2e    - Re-record all end-to-end test cases (includes .neurorc)"
    @echo "  record-neurorc    - Re-record all .neurorc startup test cases"
    @echo "  test-bench        - Run benchmark tests"
    @echo ""
    @echo "Build Commands:"
    @echo "  build             - Build all binaries (clean + lint + build)"
    @echo "  build-if-needed   - Build binaries only if sources are newer"
    @echo "  ensure-build      - Ensure binaries are built (alias for build-if-needed)"
    @echo "  build-all         - Build for multiple platforms"
    @echo ""
    @echo "Code Quality:"
    @echo "  format            - Format Go code and organize imports"
    @echo "  imports           - Organize Go imports only"
    @echo "  lint              - Run all linters and formatting"
    @echo ""
    @echo "CI/CD Commands:"
    @echo "  check-ci          - Run all CI checks locally (fast, avoids rebuilds)"
    @echo "  check-ci-clean    - Run all CI checks with clean rebuild"
    @echo "  check-ci-fast     - Run tests only (skips lint and deps)"

# Build the main binaries with version injection
build: clean lint
    @echo "Building neurotest..."
    go build -ldflags="-X 'neuroshell/internal/version.Version=$(./scripts/version.sh)' -X 'neuroshell/internal/version.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'neuroshell/internal/version.BuildDate=$(date -u +%Y-%m-%d)' " -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"
    @echo "Building NeuroShell..."
    go build -ldflags="-X 'neuroshell/internal/version.Version=$(./scripts/version.sh)' -X 'neuroshell/internal/version.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'neuroshell/internal/version.BuildDate=$(date -u +%Y-%m-%d)' " -o bin/neuro ./cmd/neuro
    @echo "Binary built at: bin/neuro"

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
        go build -ldflags="-X 'neuroshell/internal/version.Version=$(./scripts/version.sh)' -X 'neuroshell/internal/version.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'neuroshell/internal/version.BuildDate=$(date -u +%Y-%m-%d)' " -o bin/neurotest ./cmd/neurotest
        echo "Binary built at: bin/neurotest"
    else
        echo "neurotest is up to date"
    fi
    
    # Build neuro if needed
    if needs_rebuild "bin/neuro" "cmd/neuro"; then
        echo "Building NeuroShell..."
        go build -ldflags="-X 'neuroshell/internal/version.Version=$(./scripts/version.sh)' -X 'neuroshell/internal/version.GitCommit=$(git rev-parse --short HEAD 2>/dev/null || echo unknown)' -X 'neuroshell/internal/version.BuildDate=$(date -u +%Y-%m-%d)' " -o bin/neuro ./cmd/neuro
        echo "Binary built at: bin/neuro"
    else
        echo "neuro is up to date"    
    fi

# Ensure binaries are built (build if needed, but skip clean and lint)
ensure-build: build-if-needed


# Run the application
run: build
    @echo "Running NeuroShell..."
    NEURO_LOG_LEVEL=debug ./bin/neuro


# Run tests with coverage
test: ensure-build test-all-units
    @echo "Running tests..."
    EDITOR=echo go test -v -coverprofile=coverage.out \
        ./internal/... \
        ./cmd/...
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
        ./internal/version/... \
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
        ./internal/version/... \
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
        echo "PASS Coverage meets target (>=90%)"; \
    else \
        echo "FAIL Coverage below target ($$coverage% < 90%)"; \
        exit 1; \
    fi

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

# Check project health
check:
    @echo "Checking project health..."
    go mod verify
    go vet ./...
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fast; \
    fi
    @echo "Project health check complete"

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

# Initialize development environment
init:
    @echo "Initializing development environment..."
    go mod download
    @mkdir -p bin docs logs sessions
    @echo "Development environment ready"

# Run end-to-end tests
test-e2e: ensure-build
    @echo "Running end-to-end tests..."
    ./bin/neurotest --neuro-cmd="./bin/neuro" run-all
    @echo "Running .neurorc startup tests..."
    #!/bin/bash
    for test_file in $(find test/golden/neurorc -maxdepth 1 -name "*.neurorc-test" -type f | sort); do \
        test_name=$(basename "$test_file" .neurorc-test); \
        echo "Testing $test_name..."; \
        ./bin/neurotest run-neurorc "$test_name" >/dev/null 2>&1 && echo "PASS $test_name" || echo "FAIL $test_name"; \
    done
    @echo "End-to-end tests complete"

# Re-record all end-to-end test cases
record-all-e2e: ensure-build
    #!/bin/bash
    echo "Re-recording all end-to-end test cases..."
    echo "Recording standard e2e tests..."
    for test_file in $(find test/golden -maxdepth 1 -name "*.neuro" -type f | sort); do \
        test_name=$(basename "$test_file" .neuro); \
        echo "Recording $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" record "$test_name" >/dev/null 2>&1 && echo "RECORDED $test_name" || echo "FAILED $test_name"; \
    done
    echo "Recording .neurorc startup tests..."
    for test_file in $(find test/golden/neurorc -maxdepth 1 -name "*.neurorc-test" -type f | sort); do \
        test_name=$(basename "$test_file" .neurorc-test); \
        echo "Recording $test_name..."; \
        ./bin/neurotest record-neurorc "$test_name" >/dev/null 2>&1 && echo "RECORDED $test_name" || echo "FAILED $test_name"; \
    done
    echo "All end-to-end test cases re-recorded"

# Build neurotest binary
build-neurotest:
    @echo "Building neurotest..."
    go build -o bin/neurotest ./cmd/neurotest
    @echo "Binary built at: bin/neurotest"

# Record a single experiment with real API calls
experiment-record EXPERIMENT: ensure-build
    @echo "Recording experiment: {{EXPERIMENT}}"
    ./bin/neurotest --neuro-cmd="./bin/neuro" record-experiment "{{EXPERIMENT}}"

# Record all available experiments with real API calls
experiment-record-all: ensure-build
    #!/bin/bash
    echo "Recording all experiments with real API calls..."
    ./bin/neurotest --neuro-cmd="./bin/neuro" record-all-experiments

# Run an experiment and compare with a specific recording
experiment-run EXPERIMENT SESSION_ID: ensure-build
    @echo "Running experiment: {{EXPERIMENT}} (session: {{SESSION_ID}})"
    ./bin/neurotest --neuro-cmd="./bin/neuro" run-experiment "{{EXPERIMENT}}" "{{SESSION_ID}}"

# List all available experiments
experiment-list:
    @echo "Available experiments:"
    @find examples/experiments -name "*.neuro" -type f 2>/dev/null | sed 's|examples/experiments/||' | sed 's|\.neuro$||' | sort || echo "No experiments found"

# Show experiment recordings for a specific experiment
experiment-recordings EXPERIMENT:
    @echo "Recordings for experiment: {{EXPERIMENT}}"
    @if [ -d "experiments/recordings/{{EXPERIMENT}}" ]; then \
        ls -la "experiments/recordings/{{EXPERIMENT}}" | grep "\.expected$" | awk '{print $9}' | sed 's|\.expected$||' | sort; \
    else \
        echo "No recordings found for {{EXPERIMENT}}"; \
    fi

# Run .neurorc startup tests only
test-neurorc: ensure-build
    @echo "Running .neurorc startup tests..."
    #!/bin/bash
    for test_file in test/golden/neurorc/*.neurorc-test; do \
        if [ -f "$test_file" ]; then \
            test_name=$(basename "$test_file" .neurorc-test); \
            echo "Testing $test_name..."; \
            ./bin/neurotest run-neurorc "$test_name" >/dev/null 2>&1 && echo "PASS $test_name" || echo "FAIL $test_name"; \
        fi; \
    done

# Re-record all .neurorc startup test cases
record-neurorc: ensure-build
    @echo "Re-recording all .neurorc startup test cases..."
    #!/bin/bash
    for test_file in test/golden/neurorc/*.neurorc-test; do \
        if [ -f "$test_file" ]; then \
            test_name=$(basename "$test_file" .neurorc-test); \
            echo "Recording $test_name..."; \
            ./bin/neurotest record-neurorc "$test_name" >/dev/null 2>&1 && echo "RECORDED $test_name" || echo "FAILED $test_name"; \
        fi; \
    done; \
    echo "All .neurorc test cases re-recorded"
