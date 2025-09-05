# NeuroShell Development Commands
# Modular justfile for improved maintainability and readability

# =============================================================================
# MODULE IMPORTS
# =============================================================================

import 'justfiles/build.just'
import 'justfiles/test.just'
import 'justfiles/dev.just'
import 'justfiles/release.just'

# =============================================================================
# HELP AND DOCUMENTATION
# =============================================================================

# Default recipe to display available commands
default:
    @just --list
    @echo ""
    @echo "📋 NeuroShell Development Commands"
    @echo "=================================="
    @echo ""
    @echo "🔨 Build Commands:"
    @echo "  build             - Build all binaries (clean + lint + build)"
    @echo "  build-if-needed   - Build binaries only if sources are newer"
    @echo "  ensure-build      - Ensure binaries are built (alias for build-if-needed)"
    @echo "  build-all         - Build for multiple platforms"
    @echo "  build-neurotest   - Build neurotest binary only"
    @echo "  clean             - Clean build artifacts and temporary files"
    @echo "  format            - Format Go code and organize imports"
    @echo "  imports           - Organize Go imports only"
    @echo "  lint              - Run all linters and formatting"
    @echo ""
    @echo "🧪 Test Commands:"
    @echo "  test              - Run all tests with per-folder coverage report"
    @echo "  test-unit         - Run service/utils unit tests only"
    @echo "  test-commands     - Run command tests only"
    @echo "  test-parser       - Run parser tests only"
    @echo "  test-context      - Run context tests only"
    @echo "  test-shell        - Run shell tests only"
    @echo "  test-all-units    - Run all unit, command, parser, context, and shell tests"
    @echo "  test-bench        - Run benchmark tests"
    @echo ""
    @echo "  test-e2e          - Run comprehensive end-to-end tests"
    @echo "  test-neurorc      - Run .neurorc startup tests only"
    @echo "  test-c-flag       - Run -c flag e2e tests only"
    @echo "  test-e2e-full     - Run comprehensive tests (batch + -c flag + .neurorc)"
    @echo ""
    @echo "  record-all-e2e    - Re-record all end-to-end test cases"
    @echo "  record-neurorc    - Re-record all .neurorc startup test cases"
    @echo "  record-all-c-flag - Record all -c flag test cases"
    @echo "  compare-all-modes - Compare batch vs -c flag outputs for all tests"
    @echo ""
    @echo "🔬 Experiment Commands:"
    @echo "  experiment-record <NAME>          - Record experiment with real API calls"
    @echo "  experiment-record-all             - Record all experiments"
    @echo "  experiment-run <NAME> <SESSION>   - Run experiment and compare"
    @echo "  experiment-list                   - List all available experiments"
    @echo "  experiment-recordings <NAME>      - Show recordings for experiment"
    @echo ""
    @echo "⚡ Development Commands:"
    @echo "  run               - Run the application with debug logging"
    @echo "  dev               - Run development mode (rebuild on changes)"
    @echo "  init              - Initialize development environment"
    @echo "  install           - Install binary to system PATH"
    @echo "  deps              - Update dependencies"
    @echo "  docs              - Generate documentation"
    @echo "  check             - Check project health"
    @echo ""
    @echo "🚀 CI/CD Commands:"
    @echo "  check-ci          - Run all CI checks locally (fast, avoids rebuilds)"
    @echo "  check-ci-clean    - Run all CI checks with clean rebuild"
    @echo "  check-ci-fast     - Run tests only (skips lint and deps)"
    @echo ""
    @echo "📦 Release Commands:"
    @echo "  release-check <VERSION>           - Comprehensive pre-release validation"
    @echo "  release-validate <VERSION>        - Full validation (changelog + pipeline)"
    @echo "  release-validate-changelog        - Validate changelog format and syntax"
    @echo "  release-validate-goreleaser       - Validate GoReleaser configuration"
    @echo "  release-preview-changelog <VERSION> - Preview changelog for version"
    @echo "  release-version-info              - Show current version and codename info"
    @echo "  release-new-changelog-entry <VERSION> - Generate changelog entry template"
    @echo ""
    @echo "📁 Module Files:"
    @echo "  justfiles/build.just    - Build, format, lint, and cleanup recipes"
    @echo "  justfiles/test.just     - All test-related recipes and validation"
    @echo "  justfiles/dev.just      - Development utilities and CI/CD commands"
    @echo "  justfiles/release.just  - Release management and changelog tools"
    @echo ""
    @echo "💡 Tips:"
    @echo "  • Use 'just <recipe-name>' to run any command"
    @echo "  • All test commands use EDITOR=echo to prevent editor popups"
    @echo "  • CI commands are designed for both local and remote execution"
    @echo "  • Release commands validate before any destructive actions"