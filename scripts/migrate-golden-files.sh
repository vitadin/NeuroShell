#!/bin/bash

# Script to migrate neurotest golden files from nested structure to flat structure
# This script flattens the test/golden/ directory structure by moving:
# - test/golden/testname/testname.neuro -> test/golden/testname.neuro  
# - test/golden/testname/testname.expected -> test/golden/testname.expected

set -e

GOLDEN_DIR="test/golden"
BACKUP_DIR="test/golden-backup-$(date +%Y%m%d-%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}NeuroTest Golden Files Migration Script${NC}"
echo -e "${BLUE}======================================${NC}"
echo

# Check if golden directory exists
if [ ! -d "$GOLDEN_DIR" ]; then
    echo -e "${RED}Error: $GOLDEN_DIR directory not found${NC}"
    exit 1
fi

# Create backup
echo -e "${YELLOW}Creating backup at $BACKUP_DIR...${NC}"
cp -r "$GOLDEN_DIR" "$BACKUP_DIR"
echo -e "${GREEN}Backup created successfully${NC}"
echo

# Find all nested test directories
nested_tests=()
while IFS= read -r -d '' dir; do
    test_name=$(basename "$dir")
    neuro_file="$dir/$test_name.neuro"
    expected_file="$dir/$test_name.expected"
    
    if [ -f "$neuro_file" ] && [ -f "$expected_file" ]; then
        nested_tests+=("$test_name")
    fi
done < <(find "$GOLDEN_DIR" -mindepth 1 -maxdepth 1 -type d -print0)

if [ ${#nested_tests[@]} -eq 0 ]; then
    echo -e "${YELLOW}No nested test structure found. All tests may already be flattened.${NC}"
    exit 0
fi

echo -e "${BLUE}Found ${#nested_tests[@]} tests to migrate:${NC}"
for test in "${nested_tests[@]}"; do
    echo "  - $test"
done
echo

# Ask for confirmation
read -p "Proceed with migration? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Migration cancelled${NC}"
    exit 0
fi

echo -e "${BLUE}Starting migration...${NC}"
echo

migrated=0
conflicts=0

# Migrate each test
for test_name in "${nested_tests[@]}"; do
    echo -e "${BLUE}Migrating $test_name...${NC}"
    
    src_dir="$GOLDEN_DIR/$test_name"
    src_neuro="$src_dir/$test_name.neuro"
    src_expected="$src_dir/$test_name.expected"
    
    dest_neuro="$GOLDEN_DIR/$test_name.neuro"
    dest_expected="$GOLDEN_DIR/$test_name.expected"
    
    # Check for conflicts
    if [ -f "$dest_neuro" ] || [ -f "$dest_expected" ]; then
        echo -e "${RED}  ⚠️  Conflict: Flat files already exist for $test_name${NC}"
        echo -e "${RED}     Skipping to avoid overwriting${NC}"
        conflicts=$((conflicts + 1))
        continue
    fi
    
    # Move files
    if [ -f "$src_neuro" ]; then
        mv "$src_neuro" "$dest_neuro"
        echo -e "${GREEN}  ✓ Moved $test_name.neuro${NC}"
    else
        echo -e "${RED}  ✗ Missing $src_neuro${NC}"
        continue
    fi
    
    if [ -f "$src_expected" ]; then
        mv "$src_expected" "$dest_expected"
        echo -e "${GREEN}  ✓ Moved $test_name.expected${NC}"
    else
        echo -e "${RED}  ✗ Missing $src_expected${NC}"
        continue
    fi
    
    # Check if directory has any other files
    if [ -z "$(ls -A "$src_dir")" ]; then
        rmdir "$src_dir"
        echo -e "${GREEN}  ✓ Removed empty directory${NC}"
    else
        echo -e "${YELLOW}  ⚠️  Directory not empty, keeping: $src_dir${NC}"
        ls -la "$src_dir"
    fi
    
    migrated=$((migrated + 1))
    echo
done

echo -e "${BLUE}Migration Summary:${NC}"
echo -e "${GREEN}  ✓ Successfully migrated: $migrated tests${NC}"
if [ $conflicts -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️  Conflicts (skipped): $conflicts tests${NC}"
fi
echo -e "${BLUE}  📁 Backup location: $BACKUP_DIR${NC}"
echo

# Verify migration
echo -e "${BLUE}Verifying migration...${NC}"
flat_tests=0
while IFS= read -r -d '' file; do
    if [[ $file == *.neuro ]]; then
        test_name=$(basename "$file" .neuro)
        expected_file="$GOLDEN_DIR/$test_name.expected"
        if [ -f "$expected_file" ]; then
            flat_tests=$((flat_tests + 1))
        fi
    fi
done < <(find "$GOLDEN_DIR" -maxdepth 1 -name "*.neuro" -print0)

echo -e "${GREEN}Found $flat_tests flat test pairs${NC}"

# Test with neurotest if available
if command -v ./bin/neurotest &> /dev/null; then
    echo -e "${BLUE}Testing with neurotest...${NC}"
    if ./bin/neurotest run-all > /dev/null 2>&1; then
        echo -e "${GREEN}  ✓ All tests pass with new structure${NC}"
    else
        echo -e "${YELLOW}  ⚠️  Some tests may need attention${NC}"
        echo -e "${YELLOW}     Run './bin/neurotest run-all' for details${NC}"
    fi
else
    echo -e "${YELLOW}  ⚠️  neurotest binary not found, skipping test verification${NC}"
    echo -e "${YELLOW}     Build with 'just build-neurotest' and test manually${NC}"
fi

echo
echo -e "${GREEN}Migration completed!${NC}"
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Test the migration with: ./bin/neurotest run-all"
echo "  2. If tests pass, you can remove the backup: rm -rf $BACKUP_DIR"
echo "  3. Update any scripts or documentation that reference the old structure"