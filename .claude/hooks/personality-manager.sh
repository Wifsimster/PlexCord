#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PERSONALITIES_DIR="$SCRIPT_DIR/../personalities"

if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -d "$CLAUDE_PROJECT_DIR/.claude" ]]; then
  CLAUDE_DIR="$CLAUDE_PROJECT_DIR/.claude"
else
  CLAUDE_DIR="$(cd "$SCRIPT_DIR/.." 2>/dev/null && pwd)"

  if [[ "$CLAUDE_DIR" == "$HOME/.claude" ]]; then
    CLAUDE_DIR="$HOME/.claude"
  elif [[ ! -d "$CLAUDE_DIR" ]]; then
    CLAUDE_DIR="$HOME/.claude"
  fi
fi

PERSONALITY_FILE="$CLAUDE_DIR/tts-personality.txt"

get_personality_data() {
  local personality="$1"
  local field="$2"
  local file="$PERSONALITIES_DIR/${personality}.md"

  if [[ ! -f "$file" ]]; then
    return 1
  fi

  case "$field" in
    prefix)
      sed -n '/^## Prefix/,/^##/p' "$file" | sed '1d;$d' | tr -d '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
    suffix)
      sed -n '/^## Suffix/,/^##/p' "$file" | sed '1d;$d' | tr -d '\n' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
    description)
      grep "^description:" "$file" | cut -d: -f2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
    voice)
      grep "^piper_voice:" "$file" | cut -d: -f2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
    piper_voice)
      grep "^piper_voice:" "$file" | cut -d: -f2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
    instructions)
      sed -n '/^## AI Instructions/,/^##/p' "$file" | sed '1d;$d'
      ;;
  esac
}

list_personalities() {
  local personalities=()

  if [[ -d "$PERSONALITIES_DIR" ]]; then
    for file in "$PERSONALITIES_DIR"/*.md; do
      if [[ -f "$file" ]]; then
        basename "$file" .md
      fi
    done
  fi
}

case "$1" in
  list)
    CURRENT="normal"
    if [ -f "$PERSONALITY_FILE" ]; then
      CURRENT=$(cat "$PERSONALITY_FILE")
    fi

    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
    FORMATTER="$PROJECT_ROOT/src/cli/list-personalities.js"

    if [[ -f "$FORMATTER" ]] && command -v node &> /dev/null; then
      node "$FORMATTER" "$PERSONALITIES_DIR" "$CURRENT"
    else
      echo "🎭 Available Personalities:"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

      echo "Built-in personalities:"
      for personality in $(list_personalities | sort); do
        desc=$(get_personality_data "$personality" "description")
        if [[ "$personality" == "$CURRENT" ]]; then
          echo "  ✓ $personality - $desc (current)"
        else
          echo "  - $personality - $desc"
        fi
      done

      if [[ "$CURRENT" == "random" ]]; then
        echo "  ✓ random - Picks randomly each time (current)"
      else
        echo "  - random - Picks randomly each time"
      fi

      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo ""
      echo "Usage: /agent-vibes:personality <name>"
      echo "       /agent-vibes:personality add <name>"
      echo "       /agent-vibes:personality edit <name>"
    fi
    ;;

  set|switch)
    PERSONALITY="$2"

    if [[ -z "$PERSONALITY" ]]; then
      echo "❌ Please specify a personality name"
      echo "Usage: $0 set <personality>"
      exit 1
    fi

    if [[ "$PERSONALITY" != "random" ]]; then
      if [[ ! -f "$PERSONALITIES_DIR/${PERSONALITY}.md" ]]; then
        echo "❌ Personality not found: $PERSONALITY"
        echo ""
        echo "Available personalities:"
        for p in $(list_personalities | sort); do
          echo "  • $p"
        done
        exit 1
      fi
    fi

    echo "$PERSONALITY" > "$PERSONALITY_FILE"
    echo "🎭 Personality set to: $PERSONALITY"

    PROVIDER_FILE=""
    if [[ -f "$CLAUDE_DIR/tts-provider.txt" ]]; then
      PROVIDER_FILE="$CLAUDE_DIR/tts-provider.txt"
    elif [[ -f "$HOME/.claude/tts-provider.txt" ]]; then
      PROVIDER_FILE="$HOME/.claude/tts-provider.txt"
    fi

    ACTIVE_PROVIDER="piper"
    if [[ -n "$PROVIDER_FILE" ]]; then
      ACTIVE_PROVIDER=$(cat "$PROVIDER_FILE")
    fi

    ASSIGNED_VOICE=""
    if [[ "$ACTIVE_PROVIDER" == "piper" ]]; then
      ASSIGNED_VOICE=$(get_personality_data "$PERSONALITY" "piper_voice")
      if [[ -z "$ASSIGNED_VOICE" ]]; then
        ASSIGNED_VOICE="en_US-lessac-medium"
      fi
    else
      ASSIGNED_VOICE=$(get_personality_data "$PERSONALITY" "voice")
    fi

    if [[ -n "$ASSIGNED_VOICE" ]]; then
      VOICE_MANAGER="$SCRIPT_DIR/voice-manager.sh"
      if [[ -x "$VOICE_MANAGER" ]]; then
        echo "🎤 Switching to assigned voice: $ASSIGNED_VOICE"
        "$VOICE_MANAGER" switch "$ASSIGNED_VOICE" --silent >/dev/null 2>&1
      fi
    fi

    if [[ "$PERSONALITY" != "random" ]]; then
      echo ""

      TTS_SCRIPT="$SCRIPT_DIR/play-tts.sh"

      PERSONALITY_FILE_PATH="$PERSONALITIES_DIR/${PERSONALITY}.md"
      REMARK=""

      if [[ -f "$PERSONALITY_FILE_PATH" ]]; then
        mapfile -t EXAMPLES < <(grep '^- "' "$PERSONALITY_FILE_PATH" | sed 's/^- "//; s/"$//')

        if [[ ${#EXAMPLES[@]} -gt 0 ]]; then
          REMARK="${EXAMPLES[$RANDOM % ${#EXAMPLES[@]}]}"
        fi
      fi

      if [[ -z "$REMARK" ]]; then
        REMARK="Personality set to ${PERSONALITY}!"
      fi

      echo "💬 $REMARK"
      "$TTS_SCRIPT" "$REMARK"

      echo ""
      echo "Note: AI will generate unique ${PERSONALITY} responses - no fixed templates!"
      echo ""
      echo "💡 Tip: To hear automatic TTS narration, enable the Agent Vibes output style:"
      echo "   /output-style Agent Vibes"
    fi
    ;;

  get)
    if [ -f "$PERSONALITY_FILE" ]; then
      CURRENT=$(cat "$PERSONALITY_FILE")
      echo "Current personality: $CURRENT"

      if [[ "$CURRENT" != "random" ]]; then
        desc=$(get_personality_data "$CURRENT" "description")
        [[ -n "$desc" ]] && echo "Description: $desc"
      fi
    else
      echo "Current personality: normal (default)"
    fi
    ;;

  add)
    NAME="$2"
    if [[ -z "$NAME" ]]; then
      echo "❌ Please specify a personality name"
      echo "Usage: $0 add <name>"
      exit 1
    fi

    FILE="$PERSONALITIES_DIR/${NAME}.md"
    if [[ -f "$FILE" ]]; then
      echo "❌ Personality '$NAME' already exists"
      echo "Use 'edit' to modify it"
      exit 1
    fi

    cat > "$FILE" << 'EOF'
---
name: NAME
description: Custom personality
---

# NAME Personality

## Prefix


## Suffix


## AI Instructions
Describe how the AI should generate messages for this personality.

## Example Responses
- "Example response 1"
- "Example response 2"
EOF

    sed -i "s/NAME/$NAME/g" "$FILE"

    echo "✅ Created new personality: $NAME"
    echo "📝 Edit the file: $FILE"
    echo ""
    echo "You can now customize:"
    echo "  • Prefix: Text before messages"
    echo "  • Suffix: Text after messages"
    echo "  • AI Instructions: How AI should speak"
    echo "  • Example Responses: Sample messages"
    ;;

  edit)
    NAME="$2"
    if [[ -z "$NAME" ]]; then
      echo "❌ Please specify a personality name"
      echo "Usage: $0 edit <name>"
      exit 1
    fi

    FILE="$PERSONALITIES_DIR/${NAME}.md"
    if [[ ! -f "$FILE" ]]; then
      echo "❌ Personality '$NAME' not found"
      echo "Use 'add' to create it first"
      exit 1
    fi

    echo "📝 Edit this file to customize the personality:"
    echo "$FILE"
    ;;

  reset)
    echo "normal" > "$PERSONALITY_FILE"
    echo "🎭 Personality reset to: normal"
    ;;

  set-favorite-voice)
    PERSONALITY="$2"
    NEW_VOICE="$3"

    if [[ -z "$PERSONALITY" ]] || [[ -z "$NEW_VOICE" ]]; then
      echo "❌ Please specify both personality name and voice name"
      echo "Usage: $0 set-favorite-voice <personality> <voice>"
      exit 1
    fi

    FILE="$PERSONALITIES_DIR/${PERSONALITY}.md"
    if [[ ! -f "$FILE" ]]; then
      echo "❌ Personality '$PERSONALITY' not found"
      exit 1
    fi

    PROVIDER_FILE=""
    if [[ -f "$CLAUDE_DIR/tts-provider.txt" ]]; then
      PROVIDER_FILE="$CLAUDE_DIR/tts-provider.txt"
    elif [[ -f "$HOME/.claude/tts-provider.txt" ]]; then
      PROVIDER_FILE="$HOME/.claude/tts-provider.txt"
    fi

    ACTIVE_PROVIDER="piper"
    if [[ -n "$PROVIDER_FILE" ]]; then
      ACTIVE_PROVIDER=$(cat "$PROVIDER_FILE")
    fi

    if [[ "$ACTIVE_PROVIDER" == "piper" ]]; then
      VOICE_FIELD="piper_voice"
      CURRENT_VOICE=$(get_personality_data "$PERSONALITY" "piper_voice")
    else
      VOICE_FIELD="macos_voice"
      CURRENT_VOICE=$(get_personality_data "$PERSONALITY" "macos_voice")
    fi

    if [[ -n "$CURRENT_VOICE" ]] && [[ "$CURRENT_VOICE" != "$NEW_VOICE" ]]; then
      echo "⚠️  WARNING: Personality '$PERSONALITY' already has a favorite voice assigned!"
      echo ""
      echo "   Current favorite ($ACTIVE_PROVIDER): $CURRENT_VOICE"
      echo "   New voice: $NEW_VOICE"
      echo ""
      echo "Do you want to replace the favorite voice?"
      echo ""
      read -p "Enter your choice (yes/no): " CHOICE

      case "$CHOICE" in
        yes|y|YES|Y)
          echo "✅ Replacing favorite voice..."
          ;;
        no|n|NO|N)
          echo "❌ Keeping current favorite voice: $CURRENT_VOICE"
          exit 0
          ;;
        *)
          echo "❌ Invalid choice. Keeping current favorite voice: $CURRENT_VOICE"
          exit 1
          ;;
      esac
    fi

    if grep -q "^${VOICE_FIELD}:" "$FILE"; then
      sed -i "s/^${VOICE_FIELD}:.*/${VOICE_FIELD}: ${NEW_VOICE}/" "$FILE"
    else
      sed -i "/^---$/,/^---$/ { /^---$/a\\
${VOICE_FIELD}: ${NEW_VOICE}
}" "$FILE"
    fi

    echo "✅ Favorite voice for '$PERSONALITY' personality set to: $NEW_VOICE ($ACTIVE_PROVIDER)"
    echo "📝 Updated file: $FILE"
    ;;

  *)
    if [[ -n "$1" ]] && [[ -f "$PERSONALITIES_DIR/${1}.md" || "$1" == "random" ]]; then
      exec "$0" set "$1"
    else
      echo "AgentVibes Personality Manager"
      echo ""
      echo "Commands:"
      echo "  list                              - List all personalities"
      echo "  set/switch <name>                 - Set personality"
      echo "  add <name>                        - Create new personality"
      echo "  edit <name>                       - Show path to edit personality"
      echo "  get                               - Show current personality"
      echo "  set-favorite-voice <name> <voice> - Set favorite voice for a personality"
      echo "  reset                             - Reset to normal"
      echo ""
      echo "Examples:"
      echo "  /agent-vibes:personality flirty"
      echo "  /agent-vibes:personality add cowboy"
      echo "  /agent-vibes:personality set-favorite-voice flirty \"Aria\""
    fi
    ;;
esac