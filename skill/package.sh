#!/bin/bash
# Package the gagent-cli agent skill
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_FILE="${SCRIPT_DIR}/../gagent-cli.skill"

echo "📦 Packaging gagent-cli skill..."

cd "$SCRIPT_DIR"

# Remove old package if exists
rm -f "$OUTPUT_FILE"

# Create zip file with .skill extension
zip -r "$OUTPUT_FILE" SKILL.md references/ -x "*.DS_Store"

echo "✅ Skill packaged: $OUTPUT_FILE"
echo ""
echo "To install:"
echo "  cp gagent-cli.skill ~/.claude/skills/"
