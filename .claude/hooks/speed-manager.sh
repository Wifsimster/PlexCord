#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -d "$CLAUDE_PROJECT_DIR/.claude" ]]; then
  CONFIG_DIR="$CLAUDE_PROJECT_DIR/.claude/config"
else
  CURRENT_DIR="$PWD"
  while [[ "$CURRENT_DIR" != "/" ]]; do
    if [[ -d "$CURRENT_DIR/.claude" ]]; then
      CONFIG_DIR="$CURRENT_DIR/.claude/config"
      break
    fi
    CURRENT_DIR=$(dirname "$CURRENT_DIR")
  done
  if [[ -z "$CONFIG_DIR" ]]; then
    CONFIG_DIR="$HOME/.claude/config"
  fi
fi

mkdir -p "$CONFIG_DIR"

MAIN_SPEED_FILE="$CONFIG_DIR/tts-speech-rate.txt"
TARGET_SPEED_FILE="$CONFIG_DIR/tts-target-speech-rate.txt"

LEGACY_MAIN_SPEED_FILE="$CONFIG_DIR/piper-speech-rate.txt"
LEGACY_TARGET_SPEED_FILE="$CONFIG_DIR/piper-target-speech-rate.txt"

parse_speed_value() {
  local input="$1"

  case "$input" in
    normal|1x|1.0)
      echo "1.0"
      return
      ;;
    slow|slower|0.5x)
      echo "0.5"
      return
      ;;
    fast|2x|2.0)
      echo "2.0"
      return
      ;;
    faster|3x|3.0)
      echo "3.0"
      return
      ;;
  esac

  input="${input#+}"
  input="${input#-}"

  input="${input%x}"

  if [[ "$input" =~ ^[0-9]+\.?[0-9]*$ ]]; then
    echo "$input"
  else
    echo "ERROR"
  fi
}

set_speed() {
  local is_target=false
  local speed_input=""

  if [[ "$1" == "target" ]]; then
    is_target=true
    speed_input="$2"
  else
    speed_input="$1"
  fi

  if [[ -z "$speed_input" ]]; then
    echo "❌ Error: Speed value required"
    echo "Usage: /agent-vibes:set-speed [target] <speed>"
    echo "Examples: 2x, 0.5x, normal, +3x"
    return 1
  fi

  local speed_value
  speed_value=$(parse_speed_value "$speed_input")

  if [[ "$speed_value" == "ERROR" ]]; then
    echo "❌ Invalid speed value: $speed_input"
    echo "Valid values: normal, 0.5x, 1x, 2x, 3x, +2x, -2x"
    return 1
  fi

  local config_file
  local voice_type
  if [[ "$is_target" == true ]]; then
    config_file="$TARGET_SPEED_FILE"
    voice_type="target language"
  else
    config_file="$MAIN_SPEED_FILE"
    voice_type="main voice"
  fi

  echo "$speed_value" > "$config_file"

  echo "✓ Speech speed set for $voice_type"
  echo ""
  echo "Speed: ${speed_value}x"

  case "$speed_value" in
    0.5)
      echo "Effect: Half speed (slower)"
      ;;
    1.0)
      echo "Effect: Normal speed"
      ;;
    2.0)
      echo "Effect: Double speed (faster)"
      ;;
    3.0)
      echo "Effect: Triple speed (very fast)"
      ;;
    *)
      if (( $(echo "$speed_value > 1.0" | bc -l) )); then
        echo "Effect: Faster speech"
      else
        echo "Effect: Slower speech"
      fi
      ;;
  esac

  echo ""
  echo "Note: Speed control works with Piper and macOS providers"

  local test_messages=(
    "Testing speed change"
    "Speed test in progress"
    "Checking audio speed"
    "Speed configuration test"
    "Audio speed test"
  )

  local random_index=$((RANDOM % ${#test_messages[@]}))
  local test_msg="${test_messages[$random_index]}"

  echo ""
  echo "🔊 Testing new speed with: \"$test_msg\""
  "$SCRIPT_DIR/play-tts.sh" "$test_msg" &
}

migrate_legacy_files() {
  if [[ -f "$LEGACY_MAIN_SPEED_FILE" ]] && [[ ! -f "$MAIN_SPEED_FILE" ]]; then
    cp "$LEGACY_MAIN_SPEED_FILE" "$MAIN_SPEED_FILE"
  fi

  if [[ -f "$LEGACY_TARGET_SPEED_FILE" ]] && [[ ! -f "$TARGET_SPEED_FILE" ]]; then
    cp "$LEGACY_TARGET_SPEED_FILE" "$TARGET_SPEED_FILE"
  fi
}

get_speed() {
  migrate_legacy_files

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "   Current Speech Speed Settings"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""

  if [[ -f "$MAIN_SPEED_FILE" ]]; then
    local main_speed=$(grep -v '^#' "$MAIN_SPEED_FILE" 2>/dev/null | grep -v '^$' | tail -1)
    echo "Main voice: ${main_speed}x"
  else
    echo "Main voice: 1.0x (default, normal speed)"
  fi

  if [[ -f "$TARGET_SPEED_FILE" ]]; then
    local target_speed=$(cat "$TARGET_SPEED_FILE" 2>/dev/null)
    echo "Target language: ${target_speed}x"
  else
    echo "Target language: 0.5x (default, slower for learning)"
  fi

  echo ""
  echo "Scale: 0.5x=slower, 1.0x=normal, 2.0x=faster, 3.0x=very fast"
  echo "Works with: Piper TTS and macOS"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
}

case "${1:-}" in
  target)
    set_speed "target" "$2"
    ;;
  get|status)
    get_speed
    ;;
  normal|fast|slow|slower|*x|*.*|+*|-*)
    set_speed "$1"
    ;;
  *)
    echo "Speech Speed Manager"
    echo ""
    echo "Usage:"
    echo "  /agent-vibes:set-speed <speed>         Set main voice speed"
    echo "  /agent-vibes:set-speed target <speed>  Set target language speed"
    echo "  /agent-vibes:set-speed get              Show current speeds"
    echo ""
    echo "Speed values:"
    echo "  0.5x or slow/slower = Half speed (slower)"
    echo "  1x or normal        = Normal speed"
    echo "  2x or fast          = Double speed (faster)"
    echo "  3x or faster        = Triple speed (very fast)"
    echo ""
    echo "Examples:"
    echo "  /agent-vibes:set-speed 2x        # Make voice faster"
    echo "  /agent-vibes:set-speed 0.5x      # Make voice slower"
    echo "  /agent-vibes:set-speed target 0.5x  # Slow down target language for learning"
    echo "  /agent-vibes:set-speed normal    # Reset to normal"
    ;;
esac
