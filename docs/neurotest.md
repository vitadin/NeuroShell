# NeuroTest CLI Documentation

NeuroTest is an end-to-end testing tool for the Neuro CLI that uses a golden file approach to verify expected behavior and catch regressions.

## Overview

NeuroTest enables you to:
- Record expected output from `.neuro` scripts as golden files
- Compare actual output against expected output
- Update golden files when behavior changes are verified
- Run comprehensive test suites with detailed reporting

## Installation

Build the neurotest binary:

```bash
# Build both neuro and neurotest
just build

# Or build neurotest specifically
just build-neurotest

# Or manually
go build -o bin/neurotest ./cmd/neurotest
```

## Quick Start

1. **Create a test script**:
   ```bash
   mkdir -p test/golden/example
   cat > test/golden/example/example.neuro << 'EOF'
   # Example test
   \set[greeting="Hello"]
   \set[name="World"]
   \get[greeting]
   \get[name]
   EOF
   ```

2. **Record the expected output**:
   ```bash
   ./bin/neurotest record example
   ```

3. **Run the test**:
   ```bash
   ./bin/neurotest run example
   ```

4. **Run all tests**:
   ```bash
   ./bin/neurotest run-all
   ```

## Commands

### `neurotest record <testname>`

Records a new test case by running a `.neuro` script and capturing its output as the expected golden file.

```bash
# Record a test case
./bin/neurotest record basic

# Record with custom neuro binary path
./bin/neurotest --neuro-cmd=./custom/neuro record basic
```

**What it does:**
- Finds the `.neuro` script file for the test
- Executes it using the neuro CLI with `--test-mode`
- Captures and cleans the output
- Saves the output as `<testname>.expected`

### `neurotest run <testname>`

Runs a specific test case and compares its output with the expected golden file.

```bash
# Run a single test
./bin/neurotest run basic

# Run with verbose output
./bin/neurotest --verbose run basic

# Run with custom timeout
./bin/neurotest --timeout=60 run basic
```

**Exit codes:**
- `0`: Test passed
- `1`: Test failed or error occurred

### `neurotest run-all`

Runs all test cases in the test directory and provides a summary report.

```bash
# Run all tests
./bin/neurotest run-all

# Run all tests with verbose output
./bin/neurotest --verbose run-all
```

**Output example:**
```
Running basic... PASS
Running variables... PASS
Running system... FAIL

Results: 2 passed, 1 failed
```

### `neurotest accept <testname>`

Updates the golden file for a test case with the current output. Use this after verifying that new behavior is correct.

```bash
# Accept new output as correct
./bin/neurotest accept basic

# This is equivalent to re-recording the test
./bin/neurotest record basic
```

### `neurotest diff <testname>`

Shows detailed differences between the expected golden file output and the actual output from running the test.

```bash
# Show differences
./bin/neurotest diff basic
```

**Output example:**
```
=== Expected ===
Setting name = test
name = test

=== Actual ===
Setting name = test
Setting status = working
name = test
status = working

=== Differences found ===
```

### `neurotest version`

Shows version information.

```bash
./bin/neurotest version
# Output: neurotest v0.1.0
```

## Directory Structure

NeuroTest expects the following directory structure:

```
test/
├── golden/                 # Golden file tests
│   ├── basic/
│   │   ├── basic.neuro    # Test script
│   │   └── basic.expected # Expected output (auto-generated)
│   ├── variables/
│   │   ├── variables.neuro
│   │   └── variables.expected
│   └── system/
│       ├── system.neuro
│       └── system.expected
├── scripts/               # Standalone test scripts (optional)
└── fixtures/              # Test data files (optional)
```

## Global Flags

- `--neuro-cmd string`: Neuro command to test (default: "neuro")
- `--test-dir string`: Test directory (default: "test/golden")
- `--timeout int`: Test timeout in seconds (default: 30)
- `--verbose, -v`: Verbose output

## Example Workflows

### Creating a New Test Case

1. **Write the test script**:
   ```bash
   mkdir -p test/golden/interpolation
   cat > test/golden/interpolation/interpolation.neuro << 'EOF'
   # Test variable interpolation
   \set[first="Hello"]
   \set[second="World"]
   \set[combined="${first}, ${second}!"]
   \get[combined]
   \get[#test_mode]
   EOF
   ```

2. **Record the expected output**:
   ```bash
   ./bin/neurotest record interpolation
   ```

3. **Verify the test passes**:
   ```bash
   ./bin/neurotest run interpolation
   # Output: PASS: interpolation
   ```

### Handling Test Failures

1. **Run the failing test**:
   ```bash
   ./bin/neurotest run basic
   # Output: FAIL: basic
   ```

2. **Check the differences**:
   ```bash
   ./bin/neurotest diff basic
   ```

3. **If the change is expected, accept it**:
   ```bash
   ./bin/neurotest accept basic
   ```

4. **If the change is a bug, fix the code and re-test**:
   ```bash
   # Fix the code...
   ./bin/neurotest run basic
   # Output: PASS: basic
   ```

### Continuous Integration

Add to your CI pipeline:

```bash
# Build binaries
just build

# Run all end-to-end tests
just test-e2e

# Or run directly
./bin/neurotest run-all
```

**CI-friendly command:**
```bash
# Exit with proper code for CI
./bin/neurotest run-all && echo "All tests passed" || exit 1
```

## Best Practices

### Test Script Guidelines

1. **Keep tests focused**: Each test should verify specific functionality
2. **Avoid non-deterministic elements**: Don't test timestamp-based values like `#session_id`
3. **Use descriptive names**: Name tests based on what they verify
4. **Add comments**: Document what each test is checking

### Example Test Scripts

**Basic functionality test:**
```neuro
# Basic variable operations
\set[name="test"]
\get[name]
```

**Variable interpolation test:**
```neuro
# Test nested variable interpolation
\set[greeting="Hello"]
\set[name="World"]
\set[message="${greeting}, ${name}!"]
\get[message]
```

**System variables test:**
```neuro
# Test system variables (avoid timestamps)
\get[@user]
\get[#test_mode]
```

### Managing Golden Files

1. **Review changes carefully**: Always inspect diffs before accepting
2. **Version control**: Commit both `.neuro` and `.expected` files
3. **Clean up**: Remove obsolete test cases when features are removed
4. **Organize**: Group related tests in subdirectories

### Debugging Test Failures

1. **Use verbose mode**:
   ```bash
   ./bin/neurotest --verbose run failing-test
   ```

2. **Check the diff**:
   ```bash
   ./bin/neurotest diff failing-test
   ```

3. **Test manually**:
   ```bash
   # Run the neuro script directly
   ./bin/neuro --test-mode
   # Then: \run test/golden/failing-test/failing-test.neuro
   ```

## Integration with Development Workflow

### Justfile Integration

The project includes justfile targets for easy testing:

```bash
# Run all end-to-end tests
just test-e2e

# Build neurotest only
just build-neurotest
```

### Pre-commit Hook Example

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Running end-to-end tests..."
just test-e2e

if [ $? -ne 0 ]; then
    echo "End-to-end tests failed. Commit aborted."
    exit 1
fi

echo "All tests passed!"
```

## Troubleshooting

### Common Issues

**"neuro command not found"**
- Ensure the neuro binary is built: `just build`
- Check the path: `./bin/neurotest --neuro-cmd=./bin/neuro`

**"neuro script not found"**
- Verify the `.neuro` file exists in the expected location
- Check the test directory: `--test-dir=custom/path`

**"Test timeout"**
- Increase timeout: `--timeout=60`
- Check for infinite loops in test scripts

**Non-deterministic test failures**
- Avoid testing timestamp-based values
- Use `--test-mode` flag (automatically applied)
- Set consistent environment variables

### Debug Mode

For detailed debugging, run neuro directly:

```bash
# Run with debug logging
NEURO_LOG_LEVEL=debug ./bin/neuro --test-mode

# Then execute commands manually:
# \run test/golden/mytest/mytest.neuro
```

## Contributing

When adding new features to Neuro CLI:

1. Create corresponding test cases
2. Record expected outputs
3. Ensure tests pass before submitting PRs
4. Update this documentation if adding new test patterns

For questions or issues, refer to the main project documentation or create an issue in the repository.