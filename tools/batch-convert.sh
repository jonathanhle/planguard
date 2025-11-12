#!/bin/bash
#
# Batch convert Terrascan policies to Planguard rules
#
# Usage:
#   ./tools/batch-convert.sh [terrascan_dir] [resource_type]
#
# Examples:
#   ./tools/batch-convert.sh /tmp/terrascan               # Convert all AWS
#   ./tools/batch-convert.sh /tmp/terrascan aws_s3_bucket # Just S3
#   ./tools/batch-convert.sh /tmp/terrascan aws_db_instance # Just RDS
#

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
TERRASCAN_DIR="${1:-/tmp/terrascan}"
RESOURCE_TYPE="${2:-}"
DELAY=2  # Seconds between conversions (to avoid rate limits)

# Find the converter binary
CONVERTER=""
if command -v convert-terrascan &> /dev/null; then
    CONVERTER="convert-terrascan"
elif [ -f "tools/convert-terrascan/convert-terrascan" ]; then
    CONVERTER="tools/convert-terrascan/convert-terrascan"
else
    echo -e "${RED}❌ Error: convert-terrascan binary not found${NC}"
    echo ""
    echo "Build it first:"
    echo "  cd tools/convert-terrascan && make build"
    echo "Or install it:"
    echo "  cd tools/convert-terrascan && make install"
    exit 1
fi

# Check API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo -e "${RED}❌ Error: ANTHROPIC_API_KEY environment variable not set${NC}"
    echo ""
    echo "Get your API key from: https://console.anthropic.com/"
    echo "Then run: export ANTHROPIC_API_KEY='your-key-here'"
    exit 1
fi

# Check if terrascan directory exists
if [ ! -d "$TERRASCAN_DIR" ]; then
    echo -e "${RED}❌ Error: Terrascan directory not found: $TERRASCAN_DIR${NC}"
    echo ""
    echo "Clone Terrascan first:"
    echo "  git clone https://github.com/tenable/terrascan.git /tmp/terrascan"
    exit 1
fi

# Determine search path
if [ -n "$RESOURCE_TYPE" ]; then
    SEARCH_PATH="$TERRASCAN_DIR/pkg/policies/opa/rego/aws/$RESOURCE_TYPE"
    if [ ! -d "$SEARCH_PATH" ]; then
        echo -e "${RED}❌ Error: Resource type directory not found: $SEARCH_PATH${NC}"
        exit 1
    fi
    echo -e "${GREEN}🔍 Converting policies for: $RESOURCE_TYPE${NC}"
else
    SEARCH_PATH="$TERRASCAN_DIR/pkg/policies/opa/rego/aws"
    echo -e "${GREEN}🔍 Converting all AWS policies${NC}"
fi

# Find all rego files
REGO_FILES=$(find "$SEARCH_PATH" -name "*.rego" -type f)
TOTAL=$(echo "$REGO_FILES" | wc -l)

if [ "$TOTAL" -eq 0 ]; then
    echo -e "${YELLOW}⚠️  No .rego files found in $SEARCH_PATH${NC}"
    exit 0
fi

echo -e "${GREEN}📊 Found $TOTAL policy files to convert${NC}"
echo ""

# Ask for confirmation
read -p "Continue with conversion? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Convert each file
CONVERTED=0
FAILED=0

for rego_file in $REGO_FILES; do
    CONVERTED=$((CONVERTED + 1))

    # Extract filename for display
    FILENAME=$(basename "$rego_file")

    echo -e "${YELLOW}[$CONVERTED/$TOTAL]${NC} Converting: $FILENAME"

    # Convert the file
    if $CONVERTER -file "$rego_file" 2>&1; then
        echo -e "  ${GREEN}✓${NC} Success"
    else
        echo -e "  ${RED}✗${NC} Failed"
        FAILED=$((FAILED + 1))
    fi

    # Add delay to avoid rate limits (except for last file)
    if [ "$CONVERTED" -lt "$TOTAL" ]; then
        echo "  ⏱️  Waiting ${DELAY}s..."
        sleep $DELAY
    fi

    echo ""
done

# Summary
echo "================================"
echo -e "${GREEN}✅ Conversion Complete${NC}"
echo "================================"
echo "Total:      $TOTAL"
echo "Converted:  $((TOTAL - FAILED))"
if [ "$FAILED" -gt 0 ]; then
    echo -e "Failed:     ${RED}$FAILED${NC}"
else
    echo "Failed:     0"
fi
echo ""
echo "Next steps:"
echo "  1. Review converted rules in rules/aws/"
echo "  2. Test with: planguard -config .planguard/config.hcl -directory ./test-terraform"
echo "  3. Add to your config and configure exceptions as needed"
