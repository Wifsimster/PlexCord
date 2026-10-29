#!/usr/bin/env bash

export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_DIR="$(dirname "$SCRIPT_DIR")"

VERBOSITY_FILE="$CLAUDE_DIR/tts-verbosity.txt"
GLOBAL_VERBOSITY_FILE="$HOME/.claude/tts-verbosity.txt"

get_verbosity() {
  if [[ -f "$VERBOSITY_FILE" ]]; then
    cat "$VERBOSITY_FILE"
  elif [[ -f "$GLOBAL_VERBOSITY_FILE" ]]; then
    cat "$GLOBAL_VERBOSITY_FILE"
  else
    echo "low"
  fi
}

set_verbosity() {
  local level="$1"

  if [[ ! "$level" =~ ^(low|medium|high)$ ]]; then
    echo "❌ Invalid verbosity level: $level"
    echo "Valid options: low, medium, high"
    return 1
  fi

  if [[ -d "$CLAUDE_DIR" ]]; then
    echo "$level" > "$VERBOSITY_FILE"
    echo "✅ Verbosity set to: $level (project-local)"
  else
    mkdir -p "$(dirname "$GLOBAL_VERBOSITY_FILE")"
    echo "$level" > "$GLOBAL_VERBOSITY_FILE"
    echo "✅ Verbosity set to: $level (global)"
  fi

  echo ""
  echo "💡 Verbosity levels:"
  echo "   • LOW: Acknowledgments + Completions only"
  echo "   • MEDIUM: + Major decisions and findings"
  echo "   • HIGH: All reasoning (maximum transparency)"
  echo ""
  echo "⚠️  Restart Claude Code for changes to take effect"

  return 0
}

show_info() {
  local current_level
  current_level=$(get_verbosity)

  echo "🎙️  AgentVibes Verbosity Control"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Current level: $current_level"
  echo ""
  echo "Available levels:"
  echo ""
  echo "LOW (Minimal)"
  echo "  ✅ Acknowledgments only"
  echo "  ✅ Completions only"
  echo "  🔇 No reasoning spoken"
  echo ""
  echo "MEDIUM (Balanced)"
  echo "  ✅ Acknowledgments"
  echo "  🤔 Major decisions"
  echo "  ✓ Key findings"
  echo "  ✅ Completions"
  echo ""
  echo "HIGH (Maximum Transparency)"
  echo "  ✅ Acknowledgments"
  echo "  💭 All reasoning"
  echo "  🤔 All decisions"
  echo "  ✓ All findings"
  echo "  ✅ Completions"
  echo ""
  echo "Usage:"
  echo "  $0 get                # Show current level"
  echo "  $0 set low|medium|high # Change level"
}

case "$1" in
  get)
    get_verbosity
    ;;
  set)
    if [[ -z "$2" ]]; then
      echo "❌ Error: Missing verbosity level"
      echo "Usage: $0 set low|medium|high"
      exit 1
    fi
    set_verbosity "$2"
    ;;
  info|"")
    show_info
    ;;
  *)
    echo "❌ Unknown command: $1"
    echo "Usage: $0 {get|set|info} [level]"
    exit 1
    ;;
esac
