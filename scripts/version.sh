#!/bin/bash
# Script to generate dynamic version with commit count and hash
# Usage: ./scripts/version.sh [base_version]

set -euo pipefail

# Default base version
BASE_VERSION="${1:-0.2.0}"

# Get the current commit hash (short)
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "${BASE_VERSION}.0+${COMMIT_HASH}"
    exit 0
fi

# Try to find the latest tag matching the base version pattern
TAG_PATTERN="v${BASE_VERSION}"
LATEST_TAG=$(git tag -l "${TAG_PATTERN}" | sort -V | tail -1 2>/dev/null || echo "")

if [ -z "$LATEST_TAG" ]; then
    # No matching tag found, count all commits
    COMMIT_COUNT=$(git rev-list --count HEAD 2>/dev/null || echo "0")
else
    # Count commits since the tag
    COMMIT_COUNT=$(git rev-list --count "${LATEST_TAG}..HEAD" 2>/dev/null || echo "0")
fi

# Generate the version
# Use semver-compliant format: BASE_VERSION+COMMIT_COUNT.COMMIT_HASH
if [ "$COMMIT_COUNT" -gt 0 ]; then
    echo "${BASE_VERSION}+${COMMIT_COUNT}.${COMMIT_HASH}"
else
    echo "${BASE_VERSION}+${COMMIT_HASH}"
fi