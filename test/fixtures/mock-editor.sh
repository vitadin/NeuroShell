#!/bin/bash
# Mock editor script for testing _editor functionality
# Usage: mock-editor.sh <file_path>
# This script simulates an editor by:
# 1. Reading the initial content from the file
# 2. Adding distinctive markers at TOP and BOTTOM
# 3. Writing it back to the file so we can verify the range

if [ $# -ne 1 ]; then
    echo "Usage: $0 <file_path>" >&2
    exit 1
fi

FILE="$1"

# Read existing content
if [ -f "$FILE" ]; then
    CONTENT=$(cat "$FILE")
else
    CONTENT=""
fi

# Simulate "editing" by adding markers at top and bottom
# Use a fixed timestamp for consistent testing
cat > "$FILE" << EOF
%% === TOP MARKER: Mock editor started ===
%% Editor: mock-editor.sh
%% Timestamp: 2025-01-01-12:00:00
%% Original content follows:

$CONTENT

%% Original content ends above
%% Editor test completed successfully
%% === BOTTOM MARKER: Mock editor finished ===
EOF

exit 0