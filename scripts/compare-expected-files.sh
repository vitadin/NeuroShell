#!/bin/bash

# Script to compare old nested .expected files with new flat .expected files
# Reports any differences for manual verification

set -e

GOLDEN_DIR="test/golden"
BACKUP_DIR="test/golden-backup-20250728-100735"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Expected Files Comparison Report${NC}"
echo -e "${BLUE}=================================${NC}"
echo

# Check if backup directory exists
if [ ! -d "$BACKUP_DIR" ]; then
    echo -e "${RED}Error: Backup directory not found: $BACKUP_DIR${NC}"
    exit 1
fi

echo -e "${BLUE}Comparing old (nested) vs new (flat) expected files...${NC}"
echo

identical=0
different=0
missing_old=0
missing_new=0

# Find all current flat .expected files
while IFS= read -r -d '' new_file; do
    test_name=$(basename "$new_file" .expected)
    old_file="$BACKUP_DIR/$test_name/$test_name.expected"
    
    if [ ! -f "$old_file" ]; then
        echo -e "${YELLOW}⚠️  No old file found for: $test_name${NC}"
        missing_old=$((missing_old + 1))
        continue
    fi
    
    # Compare files
    if diff -q "$old_file" "$new_file" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Identical: $test_name${NC}"
        identical=$((identical + 1))
    else
        echo -e "${RED}✗ Different: $test_name${NC}"
        different=$((different + 1))
        
        # Show the differences
        echo -e "${YELLOW}  Differences:${NC}"
        diff -u "$old_file" "$new_file" | head -20 | sed 's/^/    /'
        echo
    fi
    
done < <(find "$GOLDEN_DIR" -maxdepth 1 -name "*.expected" -print0)

# Check for old files that don't have new counterparts
while IFS= read -r -d '' old_file; do
    test_name=$(basename "$(dirname "$old_file")")
    new_file="$GOLDEN_DIR/$test_name.expected"
    
    if [ ! -f "$new_file" ]; then
        echo -e "${YELLOW}⚠️  No new file found for: $test_name${NC}"
        missing_new=$((missing_new + 1))
    fi
    
done < <(find "$BACKUP_DIR" -name "*.expected" -print0)

echo -e "${BLUE}Comparison Summary:${NC}"
echo -e "${GREEN}  ✓ Identical files: $identical${NC}"
if [ $different -gt 0 ]; then
    echo -e "${RED}  ✗ Different files: $different${NC}"
fi
if [ $missing_old -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️  Missing old files: $missing_old${NC}"
fi
if [ $missing_new -gt 0 ]; then
    echo -e "${YELLOW}  ⚠️  Missing new files: $missing_new${NC}"
fi
echo

if [ $different -gt 0 ]; then
    echo -e "${YELLOW}⚠️  There are differences in expected files.${NC}"
    echo -e "${YELLOW}   Please review the differences above and verify they are expected.${NC}"
    echo -e "${YELLOW}   If differences are due to improved output or formatting, they're likely correct.${NC}"
    echo
    exit 1
else
    echo -e "${GREEN}✅ All expected files are identical or changes are expected.${NC}"
    echo -e "${GREEN}   The flattening was successful with no content changes.${NC}"
fi