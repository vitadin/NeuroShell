#!/bin/bash

# Coverage report script for NeuroShell
# This script generates per-folder coverage reports for the internal directory

echo "📊 Test Coverage by Internal Folder:"
echo "=================================="
echo ""

# List of internal folders to check coverage for
folders=(
    "commands"
    "context" 
    "output"
    "parser"
    "services"
    "shell"
    "statemachine"
    "stringprocessing"
    "version"
)

for folder in "${folders[@]}"; do
    echo -n "$folder: "
    # Run tests with coverage and extract the coverage percentage
    # Use head -1 to only get the first coverage line (main package coverage)
    coverage=$(EDITOR=echo go test -cover ./internal/$folder/... 2>/dev/null | \
        grep -E "coverage:|ok\s+\t" | \
        head -1 | \
        sed -E 's/.*coverage: ([0-9.]+)% of statements.*/\1%/' || echo "0.0%")
    echo "$coverage"
done

echo ""
echo "📈 Overall Coverage:"
echo "=================="
overall=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo "Total coverage: $overall"