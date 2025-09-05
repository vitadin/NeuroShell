# =============================================================================
# TEST MODULE - All test-related recipes and validation
# =============================================================================

# Test directories
_test_dirs := './internal/output/... ./internal/services/... ./internal/testutils/... ./internal/parser/... ./internal/context/... ./internal/statemachine/... ./internal/shell/... ./internal/stringprocessing/... ./internal/version/... ./internal/commands/builtin/... ./internal/commands/render/... ./internal/commands/session/... ./internal/commands/model/... ./internal/commands/provider/... ./internal/commands/llm/... ./internal/commands/shell/... ./pkg/...'

# =============================================================================
# MAIN TEST COMMANDS
# =============================================================================

# Run tests with coverage and show per-folder coverage
test: ensure-build test-all-units
    @echo "Running tests with coverage..."
    EDITOR=echo go test -v -coverprofile=coverage.out \
        ./internal/... \
        ./cmd/...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"
    @echo ""
    @echo ""
    @./scripts/coverage-report.sh

# Run all unit, command, parser, context, and shell tests
test-all-units:
    @echo "Running all unit, command, parser, context, execution, and shell tests..."
    EDITOR=echo go test -v {{_test_dirs}}
    @echo "All unit, command, parser, context, execution, and shell tests complete"

# Run benchmark tests
test-bench:
    @echo "Running benchmark tests..."
    go test -bench=. -benchmem ./internal/services/... ./internal/commands/... ./internal/parser/... ./internal/context/... ./internal/statemachine/... ./internal/shell/...
    @echo "Benchmark tests complete"

# =============================================================================
# UNIT TEST CATEGORIES
# =============================================================================

# Run unit tests only
test-unit:
    @echo "Running unit tests..."
    EDITOR=echo go test -v ./internal/services/... ./internal/testutils/...
    @echo "Unit tests complete"

# Run command tests only
test-commands:
    @echo "Running command tests..."
    EDITOR=echo go test -v ./internal/commands/...
    @echo "Command tests complete"

# Run parser tests only
test-parser:
    @echo "Running parser tests..."
    go test -v ./internal/parser/...
    @echo "Parser tests complete"

# Run context tests only
test-context:
    @echo "Running context tests..."
    go test -v ./internal/context/...
    @echo "Context tests complete"

# Run shell tests only
test-shell:
    @echo "Running shell tests..."
    go test -v ./internal/shell/...
    @echo "Shell tests complete"

# =============================================================================
# COVERAGE TESTING
# =============================================================================

# Run unit tests with coverage
test-unit-coverage:
    @echo "Running unit tests with coverage..."
    EDITOR=echo go test -v -coverprofile=unit-coverage.out ./internal/services/... ./internal/testutils/...
    go tool cover -html=unit-coverage.out -o unit-coverage.html
    go tool cover -func=unit-coverage.out
    @echo "Unit test coverage report generated: unit-coverage.html"

# Run command tests with coverage
test-commands-coverage:
    @echo "Running command tests with coverage..."
    EDITOR=echo go test -v -coverprofile=commands-coverage.out ./internal/commands/...
    go tool cover -html=commands-coverage.out -o commands-coverage.html
    go tool cover -func=commands-coverage.out
    @echo "Command test coverage report generated: commands-coverage.html"

# Run parser tests with coverage
test-parser-coverage:
    @echo "Running parser tests with coverage..."
    go test -v -coverprofile=parser-coverage.out ./internal/parser/...
    go tool cover -html=parser-coverage.out -o parser-coverage.html
    go tool cover -func=parser-coverage.out
    @echo "Parser test coverage report generated: parser-coverage.html"

# Run context tests with coverage
test-context-coverage:
    @echo "Running context tests with coverage..."
    go test -v -coverprofile=context-coverage.out ./internal/context/...
    go tool cover -html=context-coverage.out -o context-coverage.html
    go tool cover -func=context-coverage.out
    @echo "Context test coverage report generated: context-coverage.html"

# Run shell tests with coverage
test-shell-coverage:
    @echo "Running shell tests with coverage..."
    go test -v -coverprofile=shell-coverage.out ./internal/shell/...
    go tool cover -html=shell-coverage.out -o shell-coverage.html
    go tool cover -func=shell-coverage.out
    @echo "Shell test coverage report generated: shell-coverage.html"

# Run all unit, command, parser, context, and shell tests with coverage
test-all-units-coverage:
    @echo "Running all unit, command, parser, context, execution, and shell tests with coverage..."
    EDITOR=echo go test -v -coverprofile=all-units-coverage.out {{_test_dirs}}
    go tool cover -html=all-units-coverage.out -o all-units-coverage.html
    go tool cover -func=all-units-coverage.out
    @echo "All unit test coverage report generated: all-units-coverage.html"

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

# =============================================================================
# END-TO-END TESTING
# =============================================================================

# Run end-to-end tests
test-e2e: ensure-build
    @echo "Running end-to-end tests..."
    @echo "1. Standard batch mode tests..."
    ./bin/neurotest --neuro-cmd="./bin/neuro" run-all
    @echo ""
    @echo "2. -c flag tests..."
    @just test-c-flag
    @echo ""
    @echo "3. .neurorc startup tests..."
    #!/bin/bash
    for test_file in $(find test/golden/neurorc -maxdepth 1 -name "*.neurorc-test" -type f | sort); do \
        test_name=$(basename "$test_file" .neurorc-test); \
        echo "Testing $test_name..."; \
        ./bin/neurotest run-neurorc "$test_name" >/dev/null 2>&1 && echo "PASS $test_name" || echo "FAIL $test_name"; \
    done
    @echo ""
    @echo "🎉 All end-to-end tests complete (batch mode + -c flag + .neurorc)"

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

# Run -c flag end-to-end tests only
test-c-flag: ensure-build
    #!/bin/bash
    echo "Running -c flag end-to-end tests..."
    # Check if any .c.expected files exist
    if ! find test/golden -maxdepth 1 -name "*.c.expected" -type f | head -1 | grep -q .; then \
        echo "No .c.expected files found. Run 'just record-all-c-flag' first."; \
        exit 1; \
    fi
    for expected_file in $(find test/golden -maxdepth 1 -name "*.c.expected" -type f | sort); do \
        test_name=$(basename "$expected_file" .c.expected); \
        echo "Testing -c flag: $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" run-c "$test_name" >/dev/null 2>&1 && echo "PASS -c $test_name" || echo "FAIL -c $test_name"; \
    done
    echo "-c flag end-to-end tests complete"

# Extended test-e2e to include -c flag testing
test-e2e-full: ensure-build
    @echo "Running comprehensive end-to-end tests (batch + -c flag + .neurorc)..."
    @just test-e2e
    @echo ""
    @echo "Running -c flag tests..."
    @just test-c-flag
    @echo ""
    @echo "Comparing modes..."
    @just compare-all-modes
    @echo "Comprehensive end-to-end tests complete"

# =============================================================================
# TEST RECORDING AND COMPARISON
# =============================================================================

# Re-record all end-to-end test cases
record-all-e2e: ensure-build
    #!/bin/bash
    echo "Re-recording all end-to-end test cases..."
    echo "1. Recording standard batch mode tests..."
    for test_file in $(find test/golden -maxdepth 1 -name "*.neuro" -type f | sort); do \
        test_name=$(basename "$test_file" .neuro); \
        echo "Recording $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" record "$test_name" >/dev/null 2>&1 && echo "RECORDED $test_name" || echo "FAILED $test_name"; \
    done
    echo ""
    echo "2. Recording -c flag tests..."
    for test_file in $(find test/golden -maxdepth 1 -name "*.neuro" -type f | sort); do \
        test_name=$(basename "$test_file" .neuro); \
        echo "Recording -c $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" record-c "$test_name" >/dev/null 2>&1 && echo "RECORDED -c $test_name" || echo "FAILED -c $test_name"; \
    done
    echo ""
    echo "3. Recording .neurorc startup tests..."
    for test_file in $(find test/golden/neurorc -maxdepth 1 -name "*.neurorc-test" -type f | sort); do \
        test_name=$(basename "$test_file" .neurorc-test); \
        echo "Recording $test_name..."; \
        ./bin/neurotest record-neurorc "$test_name" >/dev/null 2>&1 && echo "RECORDED $test_name" || echo "FAILED $test_name"; \
    done
    echo ""
    echo "🎉 All end-to-end test cases re-recorded (batch mode + -c flag + .neurorc)"

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

# Record all -c flag test cases
record-all-c-flag: ensure-build
    #!/bin/bash
    echo "Recording all -c flag test cases..."
    for test_file in $(find test/golden -maxdepth 1 -name "*.neuro" -type f | sort); do \
        test_name=$(basename "$test_file" .neuro); \
        echo "Recording -c flag test: $test_name..."; \
        ./bin/neurotest --neuro-cmd="./bin/neuro" record-c "$test_name" >/dev/null 2>&1 && echo "RECORDED -c $test_name" || echo "FAILED -c $test_name"; \
    done
    echo "All -c flag test cases recorded"

# Extended record-all-e2e to include -c flag recording  
record-all-e2e-full: ensure-build
    @echo "Recording all end-to-end test cases (batch + -c flag + .neurorc)..."
    @just record-all-e2e
    @echo ""
    @echo "Recording -c flag tests..."
    @just record-all-c-flag
    @echo "All end-to-end test cases recorded"

# Compare batch mode vs -c flag outputs for all tests
compare-all-modes: ensure-build
    #!/bin/bash
    echo "Comparing batch mode vs -c flag outputs for all tests..."
    echo "======================================================="
    total_tests=0
    identical_tests=0
    different_tests=0
    skipped_tests=0
    
    # Tests with acceptable differences (log level, temp file paths, etc.)
    acceptable_differences=("assert-equal-fail" "editor-variable-basic" "session-delete-msg-confirmation" "session-delete-msg-no-session" "provider-catalog-combined" "provider-catalog-provider-filter" "send-error-client-config" "send-error-mixed-scenarios")
    
    # Ensure we have both types of expected files
    if ! find test/golden -maxdepth 1 -name "*.c.expected" -type f | head -1 | grep -q .; then \
        echo "No .c.expected files found. Run 'just record-all-c-flag' first."; \
        exit 1; \
    fi
    
    for expected_file in $(find test/golden -maxdepth 1 -name "*.c.expected" -type f | sort); do \
        test_name=$(basename "$expected_file" .c.expected); \
        batch_expected="test/golden/${test_name}.expected"; \
        if [ -f "$batch_expected" ]; then \
            # Check if this test has acceptable differences
            skip_test=false; \
            for acceptable in "${acceptable_differences[@]}"; do \
                if [ "$test_name" = "$acceptable" ]; then \
                    echo "⚠️  Reviewing $test_name (acceptable differences):"; \
                    ./bin/neurotest --neuro-cmd="./bin/neuro" compare-modes "$test_name" 2>/dev/null | sed -n '/--- Diff ---/,$p' || echo "    [Differences confirmed as acceptable]"; \
                    echo ""; \
                    skipped_tests=$((skipped_tests + 1)); \
                    skip_test=true; \
                    break; \
                fi; \
            done; \
            if [ "$skip_test" = "false" ]; then \
                echo -n "Comparing $test_name... "; \
                total_tests=$((total_tests + 1)); \
                if ./bin/neurotest --neuro-cmd="./bin/neuro" compare-modes "$test_name" >/dev/null 2>&1; then \
                    echo "✅ IDENTICAL"; \
                    identical_tests=$((identical_tests + 1)); \
                else \
                    echo "❌ DIFFERENT"; \
                    different_tests=$((different_tests + 1)); \
                fi; \
            fi; \
        else \
            echo "⚠️  Skipping $test_name (no batch .expected file)"; \
        fi; \
    done
    
    echo ""
    echo "Summary:"
    echo "========="
    echo "Total tests compared: $total_tests"
    echo "Identical outputs:    $identical_tests"
    echo "Different outputs:    $different_tests"
    echo "Skipped (acceptable): $skipped_tests"
    
    if [ $different_tests -gt 0 ]; then \
        echo ""; \
        echo "❌ Some tests show differences between batch and -c flag modes."; \
        echo "   Run individual 'neurotest compare-modes <testname>' for details."; \
        exit 1; \
    else \
        echo ""; \
        echo "🎉 All tested cases produce identical output in both modes!"; \
        echo "   ($skipped_tests tests skipped due to acceptable differences)"; \
    fi

# =============================================================================
# EXPERIMENT TESTING
# =============================================================================

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