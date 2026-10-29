#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PERSONALITIES_DIR="$SCRIPT_DIR/../personalities"

CLAUDE_DIR="$(cd "$SCRIPT_DIR/.." 2>/dev/null && pwd)"

if [[ -d "$CLAUDE_DIR" ]] && [[ "$CLAUDE_DIR" != "$HOME/.claude" ]]; then
  SENTIMENT_FILE="$CLAUDE_DIR/tts-sentiment.txt"
else
  SENTIMENT_FILE="$HOME/.claude/tts-sentiment.txt"
fi

get_personality_data() {
  local personality="$1"
  local field="$2"
  local file="$PERSONALITIES_DIR/${personality}.md"

  if [[ ! -f "$file" ]]; then
    return 1
  fi

  case "$field" in
    description)
      grep "^description:" "$file" | cut -d: -f2- | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
      ;;
  esac
}

list_personalities() {
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
    echo "🎭 Available Sentiments:"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    CURRENT="none"
    if [ -f "$SENTIMENT_FILE" ]; then
      CURRENT=$(cat "$SENTIMENT_FILE")
    fi

    echo "Available sentiment styles:"
    for personality in $(list_personalities | sort); do
      desc=$(get_personality_data "$personality" "description")
      if [[ "$personality" == "$CURRENT" ]]; then
        echo "  ✓ $personality - $desc (current)"
      else
        echo "  - $personality - $desc"
      fi
    done

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "Usage: /agent-vibes:sentiment <name>"
    echo "       /agent-vibes:sentiment clear"
    ;;

  set)
    SENTIMENT="$2"

    if [[ -z "$SENTIMENT" ]]; then
      echo "❌ Please specify a sentiment name"
      echo "Usage: $0 set <sentiment>"
      exit 1
    fi

    if [[ ! -f "$PERSONALITIES_DIR/${SENTIMENT}.md" ]]; then
      echo "❌ Sentiment not found: $SENTIMENT"
      echo ""
      echo "Available sentiments:"
      for p in $(list_personalities | sort); do
        echo "  • $p"
      done
      exit 1
    fi

    echo "$SENTIMENT" > "$SENTIMENT_FILE"
    echo "🎭 Sentiment set to: $SENTIMENT"
    echo "🎤 Voice remains unchanged"
    echo ""

    TTS_SCRIPT="$SCRIPT_DIR/play-tts.sh"

    PERSONALITY_FILE_PATH="$PERSONALITIES_DIR/${SENTIMENT}.md"
    REMARK=""

    if [[ -f "$PERSONALITY_FILE_PATH" ]]; then
      mapfile -t EXAMPLES < <(grep '^- "' "$PERSONALITY_FILE_PATH" | sed 's/^- "//; s/"$//')

      if [[ ${#EXAMPLES[@]} -gt 0 ]]; then
        REMARK="${EXAMPLES[$RANDOM % ${#EXAMPLES[@]}]}"
      fi
    fi

    if [[ -z "$REMARK" ]]; then
      REMARK="Sentiment set to ${SENTIMENT} while maintaining current voice"
    fi

    echo "💬 $REMARK"
    "$TTS_SCRIPT" "$REMARK"
    ;;

  get)
    if [ -f "$SENTIMENT_FILE" ]; then
      CURRENT=$(cat "$SENTIMENT_FILE")
      echo "Current sentiment: $CURRENT"

      desc=$(get_personality_data "$CURRENT" "description")
      [[ -n "$desc" ]] && echo "Description: $desc"
    else
      echo "Current sentiment: none (voice personality only)"
    fi
    ;;

  clear)
    rm -f "$SENTIMENT_FILE"
    echo "🎭 Sentiment cleared - using voice personality only"
    ;;

  *)
    if [[ -n "$1" ]] && [[ -f "$PERSONALITIES_DIR/${1}.md" ]]; then
      exec "$0" set "$1"
    else
      echo "AgentVibes Sentiment Manager"
      echo ""
      echo "Commands:"
      echo "  list                - List all sentiments"
      echo "  set <name>          - Set sentiment for current voice"
      echo "  get                 - Show current sentiment"
      echo "  clear               - Clear sentiment"
      echo ""
      echo "Examples:"
      echo "  /agent-vibes:sentiment flirty     # Add flirty style to current voice"
      echo "  /agent-vibes:sentiment sarcastic  # Add sarcasm to current voice"
      echo "  /agent-vibes:sentiment clear      # Remove sentiment"
    fi
    ;;
esac
