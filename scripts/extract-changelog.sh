#!/bin/bash
# Extract changelog entry for a specific version from change-logs.yaml
# Usage: ./scripts/extract-changelog.sh [version]
# Example: ./scripts/extract-changelog.sh 0.3.0

set -euo pipefail

# Default to current version from version.sh if no argument provided
VERSION="${1:-$(./scripts/version.sh | sed 's/^v//' | sed 's/+.*//')}"

# Remove 'v' prefix if present
VERSION="${VERSION#v}"

# Path to the changelog YAML file
CHANGELOG_FILE="internal/data/embedded/change-logs/change-logs.yaml"

# Check if changelog file exists
if [[ ! -f "$CHANGELOG_FILE" ]]; then
    echo "Error: Changelog file not found at $CHANGELOG_FILE" >&2
    exit 1
fi

# Check if yq is available for YAML parsing
if ! command -v yq &> /dev/null; then
    echo "Error: yq is required for YAML parsing. Install with:" >&2
    echo "  brew install yq  # macOS" >&2
    echo "  apt-get install yq  # Ubuntu/Debian" >&2
    echo "  Or download from: https://github.com/mikefarah/yq" >&2
    exit 1
fi

# Function to extract changelog entry for a specific version
extract_changelog_entry() {
    local target_version="$1"
    
    # Find the entry with matching version
    local entry=$(yq eval ".entries[] | select(.version == \"$target_version\")" "$CHANGELOG_FILE")
    
    if [[ -z "$entry" ]]; then
        echo "Error: No changelog entry found for version $target_version" >&2
        echo "Available versions:" >&2
        yq eval '.entries[].version' "$CHANGELOG_FILE" | sed 's/^/  - /' >&2
        exit 1
    fi
    
    # Extract individual fields
    local id=$(echo "$entry" | yq eval '.id' -)
    local version=$(echo "$entry" | yq eval '.version' -)
    local date=$(echo "$entry" | yq eval '.date' -)
    local type=$(echo "$entry" | yq eval '.type' -)
    local title=$(echo "$entry" | yq eval '.title' -)
    local description=$(echo "$entry" | yq eval '.description' -)
    local impact=$(echo "$entry" | yq eval '.impact' -)
    local files_changed=$(echo "$entry" | yq eval '.files_changed[]' - | tr '\n' ' ')
    
    # Format the changelog entry for GitHub release
    cat << EOF
### ${title}

**Version:** ${version}  
**Date:** ${date}  
**Type:** ${type}  
**ID:** ${id}

#### Description
${description}

#### Impact
${impact}

#### Files Changed
$(echo "$files_changed" | tr ' ' '\n' | sed 's/^/- /' | grep -v '^- $')

EOF
}

# Function to extract just the description for shorter format
extract_description_only() {
    local target_version="$1"
    
    # Find and return just the description
    yq eval ".entries[] | select(.version == \"$target_version\") | .description" "$CHANGELOG_FILE"
}

# Function to extract title only
extract_title_only() {
    local target_version="$1"
    
    # Find and return just the title
    yq eval ".entries[] | select(.version == \"$target_version\") | .title" "$CHANGELOG_FILE"
}

# Function to extract changelog in GoReleaser format
extract_goreleaser_format() {
    local target_version="$1"
    
    # Find the entry with matching version
    local entry=$(yq eval ".entries[] | select(.version == \"$target_version\")" "$CHANGELOG_FILE")
    
    if [[ -z "$entry" ]]; then
        echo "No changelog entry found for version $target_version"
        return
    fi
    
    # Extract fields
    local title=$(echo "$entry" | yq eval '.title' -)
    local description=$(echo "$entry" | yq eval '.description' -)
    local impact=$(echo "$entry" | yq eval '.impact' -)
    
    # Format for GoReleaser
    echo "**${title}**"
    echo ""
    echo "${description}"
    echo ""
    echo "**User Impact:** ${impact}"
}

# Check command line arguments for format option
FORMAT="${2:-full}"

case "$FORMAT" in
    "full")
        extract_changelog_entry "$VERSION"
        ;;
    "description")
        extract_description_only "$VERSION"
        ;;
    "title")
        extract_title_only "$VERSION"
        ;;
    "goreleaser")
        extract_goreleaser_format "$VERSION"
        ;;
    *)
        echo "Usage: $0 [version] [format]" >&2
        echo "" >&2
        echo "Formats:" >&2
        echo "  full        - Complete changelog entry (default)" >&2
        echo "  description - Description field only" >&2
        echo "  title       - Title field only" >&2
        echo "  goreleaser  - Format suitable for GoReleaser" >&2
        echo "" >&2
        echo "Examples:" >&2
        echo "  $0 0.3.0" >&2
        echo "  $0 0.3.0 description" >&2
        echo "  $0 0.3.0 goreleaser" >&2
        exit 1
        ;;
esac