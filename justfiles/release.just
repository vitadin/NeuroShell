# =============================================================================
# RELEASE MODULE - Release management, changelog, and validation
# =============================================================================

# =============================================================================
# GORELEASER VALIDATION
# =============================================================================

# Validate GoReleaser configuration locally
release-validate-goreleaser:
    #!/bin/bash
    set -euo pipefail
    
    echo "🔧 GoReleaser Configuration Validation"
    echo "======================================"
    echo ""
    
    # Check if goreleaser is installed
    if ! command -v goreleaser >/dev/null 2>&1; then
        echo "❌ GoReleaser not found. Installing..."
        if command -v brew >/dev/null 2>&1; then
            brew install --cask goreleaser
        else
            echo "   Please install GoReleaser: https://goreleaser.com/install/"
            exit 1
        fi
    fi
    echo "✅ GoReleaser available: $(goreleaser --version | head -1)"
    echo ""
    
    # Validate configuration syntax
    echo "Validating .goreleaser.yaml syntax..."
    if goreleaser check; then
        echo "✅ GoReleaser configuration is valid"
    else
        echo "❌ GoReleaser configuration has errors"
        exit 1
    fi
    echo ""
    
    # Test build without releasing (snapshot mode)
    echo "Testing build process (snapshot mode)..."
    if goreleaser build --snapshot --clean --single-target; then
        echo "✅ GoReleaser build test successful"
        echo "   Generated binaries in ./dist/ directory"
    else
        echo "❌ GoReleaser build test failed"
        exit 1
    fi
    echo ""
    
    # Test release templates by doing a dry-run
    echo "Testing release templates (dry-run mode)..."
    if goreleaser release --snapshot --clean --skip=publish; then
        echo "✅ GoReleaser release template test successful"
        echo "   All templates and configurations validated"
    else
        echo "❌ GoReleaser release template test failed"
        exit 1
    fi
    echo ""
    
    echo "🎉 GoReleaser configuration validated successfully!"

# =============================================================================
# RELEASE PIPELINE VALIDATION
# =============================================================================

# Check release pipeline locally before pushing tags
release-check VERSION:
    #!/bin/bash
    set -euo pipefail
    
    VERSION="{{VERSION}}"
    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"
    
    echo "🔍 Release Pipeline Check for v${VERSION}"
    echo "======================================"
    echo ""
    
    # 1. Validate version format (semantic versioning)
    echo "1. Validating version format..."
    if ! echo "${VERSION}" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        echo "❌ Invalid version format: ${VERSION}"
        echo "   Expected format: X.Y.Z (e.g., 0.2.3, 1.0.0)"
        exit 1
    fi
    echo "✅ Version format valid: ${VERSION}"
    echo ""
    
    # 2. Check if yq is available (required for changelog processing)
    echo "2. Checking dependencies..."
    if ! command -v yq >/dev/null 2>&1; then
        echo "❌ yq is required for changelog processing"
        echo "   Install with: brew install yq"
        exit 1
    fi
    echo "✅ yq available: $(yq --version)"
    echo ""
    
    # 3. Validate Go modules and build
    echo "3. Validating Go environment..."
    go mod verify
    go mod tidy
    echo "✅ Go modules verified"
    echo ""
    
    # 4. Test builds for main binaries
    echo "4. Testing binary builds..."
    echo "   Building neuro..."
    go build -v ./cmd/neuro > /dev/null
    echo "   Building neurotest..."
    go build -v ./cmd/neurotest > /dev/null
    echo "✅ All binaries build successfully"
    echo ""
    
    # 5. Validate changelog entry exists
    echo "5. Validating changelog entry..."
    if ! yq eval ".entries[] | select(.version == \"${VERSION}\")" internal/data/embedded/change-logs/change-logs.yaml | grep -q .; then
        echo "❌ No changelog entry found for version ${VERSION}"
        echo "   Available versions:"
        yq eval '.entries[].version' internal/data/embedded/change-logs/change-logs.yaml | sed 's/^/     /'
        echo ""
        echo "   Add a changelog entry before releasing."
        echo "   See docs/CHANGELOG_TEMPLATE.md for format."
        exit 1
    fi
    echo "✅ Changelog entry found for version ${VERSION}"
    
    # Extract and display changelog entry
    echo ""
    echo "📋 Changelog Entry Preview:"
    echo "─────────────────────────────"
    scripts/extract-changelog.sh "${VERSION}" goreleaser
    echo ""
    
    # 6. Check version consistency with scripts/version.sh
    echo "6. Checking version consistency..."
    SCRIPT_VERSION=$(./scripts/version.sh | sed 's/+.*//')
    if [[ "${VERSION}" != "${SCRIPT_VERSION}" ]]; then
        echo "❌ Version mismatch detected:"
        echo "   Requested version: ${VERSION}"
        echo "   scripts/version.sh base: ${SCRIPT_VERSION}"
        echo ""
        echo "   Update scripts/version.sh BASE_VERSION to ${VERSION}"
        exit 1
    fi
    echo "✅ Version consistency verified"
    echo ""
    
    # 7. Check for codename mapping
    echo "7. Checking codename mapping..."
    # Use Go to check codename (create a small test)
    CODENAME=$(printf 'package main\nimport (\n    "fmt"\n    "neuroshell/internal/version"\n)\nfunc main() {\n    fmt.Print(version.GetCodenameForVersion("%s"))\n}' "${VERSION}" | go run -)
    if [[ -n "${CODENAME}" ]]; then
        echo "✅ Codename found: '${CODENAME}'"
        echo "   Release title will be: NeuroShell v${VERSION} '${CODENAME}'"
    else
        echo "⚠️  No codename found for version ${VERSION}"
        echo "   Release title will be: NeuroShell v${VERSION}"
    fi
    echo ""
    
    # 8. Validate GoReleaser configuration (if goreleaser is available)
    echo "8. Validating GoReleaser configuration..."
    if command -v goreleaser >/dev/null 2>&1; then
        if goreleaser check; then
            echo "✅ GoReleaser configuration valid"
        else
            echo "❌ GoReleaser configuration has issues"
            exit 1
        fi
    else
        echo "⚠️  GoReleaser not installed locally (will use GitHub Actions version)"
        echo "   Install locally for validation: brew install goreleaser"
    fi
    echo ""
    
    # 9. Check git status
    echo "9. Checking git status..."
    if [[ -n "$(git status --porcelain)" ]]; then
        echo "⚠️  Working directory has uncommitted changes:"
        git status --short
        echo ""
        echo "   Consider committing changes before creating release tag."
    else
        echo "✅ Working directory clean"
    fi
    echo ""
    
    # 10. Check if tag already exists
    echo "10. Checking if tag exists..."
    if git rev-parse --verify "v${VERSION}" >/dev/null 2>&1; then
        echo "❌ Tag v${VERSION} already exists"
        echo "   Delete with: git tag -d v${VERSION} && git push origin :refs/tags/v${VERSION}"
        exit 1
    fi
    echo "✅ Tag v${VERSION} is available"
    echo ""
    
    echo "🎉 Release Check Complete!"
    echo "========================="
    echo ""
    echo "All checks passed for version ${VERSION}."
    echo ""
    echo "To create the release:"
    echo "  1. git tag v${VERSION}"
    echo "  2. git push origin v${VERSION}"
    echo ""
    echo "The GitHub Actions release pipeline will automatically:"
    echo "  • Build cross-platform binaries"
    echo "  • Create GitHub release with changelog"
    echo "  • Upload release assets"
    if [[ -n "${CODENAME}" ]]; then
        echo "  • Use codename '${CODENAME}' in release title"
    fi

# =============================================================================
# CHANGELOG MANAGEMENT
# =============================================================================

# Validate current changelog format and syntax
release-validate-changelog:
    #!/bin/bash
    set -euo pipefail
    
    echo "📋 Validating changelog format and syntax..."
    echo "============================================="
    echo ""
    
    # Check YAML syntax
    echo "1. Checking YAML syntax..."
    if yq eval '.' internal/data/embedded/change-logs/change-logs.yaml >/dev/null; then
        echo "✅ YAML syntax valid"
    else
        echo "❌ YAML syntax errors detected"
        exit 1
    fi
    echo ""
    
    # Check required fields for each entry
    echo "2. Validating entry structure..."
    yq eval '.entries[] | [.id, .version, .date, .type, .title, .description, .impact] | @csv' internal/data/embedded/change-logs/change-logs.yaml | \
    while IFS=, read -r id version date type title description impact; do
        if [[ -z "$id" || -z "$version" || -z "$date" || -z "$type" || -z "$title" ]]; then
            echo "❌ Missing required fields in entry: $id"
            exit 1
        fi
    done
    echo "✅ All entries have required fields"
    echo ""
    
    # Check ID uniqueness
    echo "3. Checking ID uniqueness..."
    DUPLICATE_IDS=$(yq eval '.entries[].id' internal/data/embedded/change-logs/change-logs.yaml | sort | uniq -d)
    if [[ -n "$DUPLICATE_IDS" ]]; then
        echo "❌ Duplicate IDs found: $DUPLICATE_IDS"
        exit 1
    fi
    echo "✅ All IDs are unique"
    echo ""
    
    # Show latest entries
    echo "4. Latest changelog entries:"
    echo "────────────────────────────"
    yq eval '.entries | sort_by(.date) | reverse | .[0:3] | .[] | "[" + .id + "] " + .version + " (" + .date + ") - " + .title' internal/data/embedded/change-logs/change-logs.yaml
    echo ""
    echo "✅ Changelog validation complete"

# Extract and preview changelog for a specific version
release-preview-changelog VERSION:
    #!/bin/bash
    VERSION="{{VERSION}}"
    VERSION="${VERSION#v}"  # Remove 'v' prefix if present
    
    echo "📋 Changelog Preview for v${VERSION}"
    echo "=================================="
    echo ""
    
    if ! scripts/extract-changelog.sh "${VERSION}" >/dev/null 2>&1; then
        echo "❌ No changelog entry found for version ${VERSION}"
        echo ""
        echo "Available versions:"
        yq eval '.entries[].version' internal/data/embedded/change-logs/change-logs.yaml | sed 's/^/  - v/'
        exit 1
    fi
    
    echo "Full Format:"
    echo "────────────"
    scripts/extract-changelog.sh "${VERSION}" full
    echo ""
    echo "GoReleaser Format (for GitHub release):"
    echo "───────────────────────────────────────"
    scripts/extract-changelog.sh "${VERSION}" goreleaser

# Create a new changelog entry template
release-new-changelog-entry VERSION:
    #!/bin/bash
    VERSION="{{VERSION}}"
    VERSION="${VERSION#v}"  # Remove 'v' prefix if present
    
    echo "📝 Creating changelog entry template for v${VERSION}"
    echo "=================================================="
    echo ""
    
    # Check if entry already exists
    if yq eval ".entries[] | select(.version == \"${VERSION}\")" internal/data/embedded/change-logs/change-logs.yaml | grep -q .; then
        echo "❌ Changelog entry for version ${VERSION} already exists"
        exit 1
    fi
    
    # Get next changelog ID
    LAST_ID=$(yq eval '.entries[].id' internal/data/embedded/change-logs/change-logs.yaml | sort -V | tail -1)
    LAST_NUM=$(echo "${LAST_ID}" | sed 's/CL//')
    NEXT_NUM=$((LAST_NUM + 1))
    NEXT_ID="CL${NEXT_NUM}"
    
    # Get current date
    CURRENT_DATE=$(date +%Y-%m-%d)
    
    # Create template
    echo "# Add this entry to internal/data/embedded/change-logs/change-logs.yaml"
    echo "# at the TOP of the entries array (newest first):"
    echo ""
    echo "- id: \"${NEXT_ID}\""
    echo "  version: \"${VERSION}\""
    echo "  date: \"${CURRENT_DATE}\""
    echo "  type: \"feature\"  # feature, enhancement, bugfix, performance, security, testing, refactor, docs, chore, breaking"
    echo "  title: \"[Brief Summary - Max 80 Characters]\""
    echo "  description: \"[Detailed description of changes - 2-4 sentences]\""
    echo "  impact: \"[User-facing impact description - 1-3 sentences]\""
    echo "  files_changed: ["
    echo "    \"path/to/modified/file.go\","
    echo "    \"internal/commands/*/\","
    echo "    \"test/golden/feature-*\""
    echo "  ]"
    echo ""
    echo "# Template generated for version ${VERSION}"
    echo "# See docs/CHANGELOG_TEMPLATE.md for detailed formatting guidelines"
    
    echo ""
    echo "✅ Template created! Copy the above YAML into the changelog file."
    echo ""
    echo "Next steps:"
    echo "1. Edit internal/data/embedded/change-logs/change-logs.yaml"
    echo "2. Add the template entry at the TOP of the entries array"
    echo "3. Fill in the title, description, impact, and files_changed"
    echo "4. Run: just release-validate-changelog"
    echo "5. Run: just release-check ${VERSION}"

# =============================================================================
# VERSION INFORMATION
# =============================================================================

# Show current version and codename information
release-version-info:
    #!/bin/bash
    echo "📊 Version Information"
    echo "====================="
    echo ""
    
    # Current version from script
    CURRENT_VERSION=$(./scripts/version.sh)
    BASE_VERSION=$(echo "${CURRENT_VERSION}" | sed 's/+.*//')
    BUILD_META=$(echo "${CURRENT_VERSION}" | sed 's/^[^+]*//' | sed 's/^+//')
    
    echo "Current Version: ${CURRENT_VERSION}"
    echo "Base Version:    ${BASE_VERSION}"
    echo "Build Metadata:  ${BUILD_META}"
    echo ""
    
    # Get codename
    CODENAME=$(printf 'package main\nimport (\n    "fmt"\n    "neuroshell/internal/version"\n)\nfunc main() {\n    fmt.Print(version.GetCodenameForVersion("%s"))\n}' "${BASE_VERSION}" | go run -)
    
    if [[ -n "${CODENAME}" ]]; then
        echo "Codename:        '${CODENAME}'"
        echo "Release Title:   NeuroShell v${BASE_VERSION} '${CODENAME}'"
    else
        echo "Codename:        (none)"
        echo "Release Title:   NeuroShell v${BASE_VERSION}"
    fi
    echo ""
    
    # Show available codenames
    echo "Available Codenames:"
    echo "───────────────────"
    echo "0.1.0 = Hydra      (Simple nerve net)"
    echo "0.2.0 = Planaria   (Simple brain, basic learning) ← Current series"
    echo "0.3.0 = Aplysia    (Sea slug, neuroscience research)"
    echo "0.4.0 = Octopus    (Highly intelligent invertebrate)"
    echo "0.5.0 = Corvus     (Crow, exceptional intelligence)"
    echo "0.6.0 = Rattus     (Rat, neuroscience model)"
    echo "0.7.0 = Macaca     (Macaque, advanced cognition)"
    echo "0.8.0 = Pan        (Chimpanzee, tool use)"
    echo "0.9.0 = Tursiops   (Dolphin, self-awareness)"
    echo "1.0.0 = Sapiens    (Human-level milestone)"
    echo "2.0.0 = Synthia    (Synthetic intelligence)"

# =============================================================================
# COMPREHENSIVE RELEASE VALIDATION
# =============================================================================

# Comprehensive pre-release validation
release-validate VERSION:
    @echo "🔍 Comprehensive Release Validation for v{{VERSION}}"
    @echo "=================================================="
    @just release-validate-changelog
    @echo ""
    @just release-check "{{VERSION}}"