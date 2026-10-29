#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_DIR="$(dirname "$SCRIPT_DIR")"

if [[ -d "$CLAUDE_DIR" ]] && [[ "$CLAUDE_DIR" != "$HOME/.claude" ]]; then
  REMINDER_FILE="$CLAUDE_DIR/github-star-reminder.txt"
else
  REMINDER_FILE="$HOME/.claude/github-star-reminder.txt"
  mkdir -p "$HOME/.claude"
fi

GITHUB_REPO="https://github.com/paulpreibisch/AgentVibes"

is_reminder_disabled() {
  local disable_file_local="$CLAUDE_DIR/github-star-reminder-disabled.flag"
  local disable_file_global="$HOME/.claude/github-star-reminder-disabled.flag"

  if [[ -f "$disable_file_local" ]] || [[ -f "$disable_file_global" ]]; then
    return 0
  fi

  if [[ -f "$REMINDER_FILE" ]]; then
    local content=$(cat "$REMINDER_FILE" 2>/dev/null)
    if [[ "$content" == "disabled" ]]; then
      return 0
    fi
  fi

  return 1
}

should_show_reminder() {
  if is_reminder_disabled; then
    return 1
  fi

  if [[ ! -f "$REMINDER_FILE" ]]; then
    return 0
  fi

  LAST_REMINDER=$(cat "$REMINDER_FILE" 2>/dev/null || echo "0")
  CURRENT_DATE=$(date +%Y%m%d)

  if [[ "$LAST_REMINDER" != "$CURRENT_DATE" ]]; then
    return 0
  fi

  return 1
}

show_reminder() {
  echo ""
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "⭐ Enjoying AgentVibes?"
  echo ""
  echo "   If you find this project helpful, please consider giving us"
  echo "   a star on GitHub! It helps others discover AgentVibes and"
  echo "   motivates us to keep improving it."
  echo ""
  echo "   👉 $GITHUB_REPO"
  echo ""
  echo "   Thank you for your support! 🙏"
  echo ""
  echo "   💡 To disable these reminders, run:"
  echo "   echo \"disabled\" > $REMINDER_FILE"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  date +%Y%m%d > "$REMINDER_FILE"
}

if should_show_reminder; then
  show_reminder
fi
