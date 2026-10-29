#!/usr/bin/env bash

set -euo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

BG_DIR="$SCRIPT_DIR/../audio/tracks"
CONFIG_FILE="$SCRIPT_DIR/../config/audio-effects.cfg"

CHANGES_MADE=false

echo "🔄 Checking for old background music structure..."

if [[ -d "$BG_DIR/optimized" ]]; then
    echo "  → Removing old optimized/ subdirectory..."

    if [[ -n "$(find "$BG_DIR/optimized" -type f -name "*.mp3" 2>/dev/null)" ]]; then
        find "$BG_DIR/optimized" -type f -name "*.mp3" 2>/dev/null | while IFS= read -r file; do
            basename_file=$(basename "$file")
            snake_case_file=$(echo "$basename_file" | tr '[:upper:]' '[:lower:]' | tr ' ' '_' | tr '-' '_')

            if [[ ! -f "$BG_DIR/$snake_case_file" ]]; then
                echo "    • Migrating: $basename_file → $snake_case_file"
                mv "$file" "$BG_DIR/$snake_case_file"
                CHANGES_MADE=true
            fi
        done
    fi

    rm -rf "$BG_DIR/optimized"
    echo "  ✓ Removed optimized/ subdirectory"
    CHANGES_MADE=true
fi

if [[ -d "$BG_DIR" ]]; then
    find "$BG_DIR" -maxdepth 1 -type f -name "*.mp3" 2>/dev/null | while IFS= read -r file; do
        basename_file=$(basename "$file")

        if [[ "$basename_file" =~ [[:space:]] ]] || [[ "$basename_file" =~ [A-Z] ]]; then
            snake_case_file=$(echo "$basename_file" | tr '[:upper:]' '[:lower:]' | tr ' ' '_' | tr '-' '_')

            if [[ -f "$BG_DIR/$snake_case_file" ]] && [[ "$basename_file" != "$snake_case_file" ]]; then
                echo "  → Removing old format: $basename_file (replaced by $snake_case_file)"
                rm -f "$file"
                CHANGES_MADE=true
            fi
        fi
    done
fi

if [[ -f "$CONFIG_FILE" ]]; then
    if grep -q "optimized/" "$CONFIG_FILE" 2>/dev/null; then
        echo "  → Updating config to remove optimized/ prefixes..."

        cp "$CONFIG_FILE" "${CONFIG_FILE}.backup-migration"

        sed -i.bak 's|optimized/||g' "$CONFIG_FILE"
        rm -f "${CONFIG_FILE}.bak"

        echo "  ✓ Updated audio-effects.cfg"
        CHANGES_MADE=true
    fi

    if grep -E '\|[^|]*[A-Z ][^|]*\.mp3\|' "$CONFIG_FILE" 2>/dev/null | grep -v '^#' > /dev/null; then
        echo "  → Converting config entries to snake_case..."

        temp_file=$(mktemp)
        while IFS='|' read -r field1 field2 field3 field4 rest; do
            if [[ "$field1" =~ ^#.* ]] || [[ -z "$field1" ]]; then
                echo "$field1|$field2|$field3|$field4$rest"
                continue
            fi

            if [[ -n "$field3" ]] && ([[ "$field3" =~ [[:space:]] ]] || [[ "$field3" =~ [A-Z] ]]); then
                new_field3=$(echo "$field3" | tr '[:upper:]' '[:lower:]' | tr ' ' '_' | tr '-' '_')
                echo "$field1|$field2|$new_field3|$field4$rest"
            else
                echo "$field1|$field2|$field3|$field4$rest"
            fi
        done < "$CONFIG_FILE" > "$temp_file"

        mv "$temp_file" "$CONFIG_FILE"
        echo "  ✓ Converted config entries to snake_case"
        CHANGES_MADE=true
    fi
fi

if [[ "$CHANGES_MADE" == "true" ]]; then
    echo "✅ Migration complete! Old background music structure cleaned up."
else
    echo "✅ No migration needed - structure is already up to date."
fi
