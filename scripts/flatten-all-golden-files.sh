#!/bin/bash

# Script to completely flatten neurotest golden files from nested to flat structure
# This script moves all files to flat structure and creates backup for comparison

set -e

GOLDEN_DIR="test/golden"
BACKUP_DIR="test/golden-backup-$(date +%Y%m%d-%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}NeuroTest Complete Golden Files Flattening Script${NC}"
echo -e "${BLUE}==================================================${NC}"
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

# Find all nested test directories and files
echo -e "${BLUE}Analyzing current structure...${NC}"
nested_tests=()
auxiliary_files=()

while IFS= read -r -d '' dir; do
    test_name=$(basename "$dir")
    neuro_file="$dir/$test_name.neuro"
    expected_file="$dir/$test_name.expected"
    
    if [ -f "$neuro_file" ]; then
        nested_tests+=("$test_name")
        
        # Check for auxiliary files
        for file in "$dir"/*; do
            if [ -f "$file" ]; then
                filename=$(basename "$file")
                if [[ "$filename" != "$test_name.neuro" && "$filename" != "$test_name.expected" ]]; then
                    auxiliary_files+=("$test_name/$filename")
                fi
            fi
        done
    fi
done < <(find "$GOLDEN_DIR" -mindepth 1 -maxdepth 1 -type d -print0)

echo -e "${BLUE}Found ${#nested_tests[@]} tests to flatten${NC}"
if [ ${#auxiliary_files[@]} -gt 0 ]; then
    echo -e "${YELLOW}Found ${#auxiliary_files[@]} auxiliary files:${NC}"
    for aux in "${auxiliary_files[@]}"; do
        echo "  - $aux"
    done
fi
echo

# Ask for confirmation
read -p "Proceed with complete flattening? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Flattening cancelled${NC}"
    exit 0
fi

echo -e "${BLUE}Starting complete flattening...${NC}"
echo

moved=0
conflicts=0
aux_handled=0

# Flatten each test
for test_name in "${nested_tests[@]}"; do
    echo -e "${BLUE}Processing $test_name...${NC}"
    
    src_dir="$GOLDEN_DIR/$test_name"
    src_neuro="$src_dir/$test_name.neuro"
    src_expected="$src_dir/$test_name.expected"
    
    dest_neuro="$GOLDEN_DIR/$test_name.neuro"
    dest_expected="$GOLDEN_DIR/$test_name.expected"
    
    # Check for conflicts
    if [ -f "$dest_neuro" ] || [ -f "$dest_expected" ]; then
        echo -e "${RED}  ⚠️  Conflict: Flat files already exist for $test_name${NC}"
        conflicts=$((conflicts + 1))
        continue
    fi
    
    # Move .neuro file
    if [ -f "$src_neuro" ]; then
        mv "$src_neuro" "$dest_neuro"
        echo -e "${GREEN}  ✓ Moved $test_name.neuro${NC}"
    else
        echo -e "${RED}  ✗ Missing $src_neuro${NC}"
        continue
    fi
    
    # Move .expected file if it exists
    if [ -f "$src_expected" ]; then
        mv "$src_expected" "$dest_expected"
        echo -e "${GREEN}  ✓ Moved $test_name.expected${NC}"
    else
        echo -e "${YELLOW}  ⚠️  No .expected file (will be generated)${NC}"
    fi
    
    # Handle auxiliary files by keeping them in place
    # (They will remain in subdirectories for tests that need them)
    aux_count=0
    for file in "$src_dir"/*; do
        if [ -f "$file" ]; then
            aux_count=$((aux_count + 1))
        fi
    done
    
    if [ $aux_count -gt 0 ]; then
        echo -e "${YELLOW}  ⚠️  Keeping $aux_count auxiliary files in $src_dir${NC}"
        aux_handled=$((aux_handled + aux_count))
    else
        # Remove empty directory
        rmdir "$src_dir"
        echo -e "${GREEN}  ✓ Removed empty directory${NC}"
    fi
    
    moved=$((moved + 1))
    echo
done

echo -e "${BLUE}Flattening Summary:${NC}"
echo -e "${GREEN}  ✓ Successfully flattened: $moved tests${NC}"
if [ $conflicts -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️  Conflicts (skipped): $conflicts tests${NC}"
fi
if [ $aux_handled -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️  Auxiliary files preserved: $aux_handled files${NC}"
fi
echo -e "${BLUE}  📁 Backup location: $BACKUP_DIR${NC}"
echo

# Verify flattening
echo -e "${BLUE}Verifying flattening...${NC}"
flat_neuro_count=$(find "$GOLDEN_DIR" -maxdepth 1 -name "*.neuro" | wc -l)
flat_expected_count=$(find "$GOLDEN_DIR" -maxdepth 1 -name "*.expected" | wc -l)

echo -e "${GREEN}Found $flat_neuro_count flat .neuro files${NC}"
echo -e "${GREEN}Found $flat_expected_count flat .expected files${NC}"
echo

echo -e "${GREEN}Complete flattening finished!${NC}"
echo -e "${BLUE}Next steps:${NC}"
echo "  1. Update justfile record-all-e2e command"
echo "  2. Build neurotest: just build-neurotest"
echo "  3. Regenerate expected files: just record-all-e2e"
echo "  4. Compare with backup to verify changes"
echo
echo -e "${YELLOW}Note: Some tests may have auxiliary files remaining in subdirectories.${NC}"
echo -e "${YELLOW}These are test data files and should be preserved.${NC}"