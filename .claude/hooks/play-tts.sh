#!/usr/bin/env bash

export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

GLOBAL_MUTE_FILE="$HOME/.agentvibes-muted"
PROJECT_MUTE_FILE="$PROJECT_ROOT/.claude/agentvibes-muted"
PROJECT_UNMUTE_FILE="$PROJECT_ROOT/.claude/agentvibes-unmuted"

if [[ -f "$PROJECT_UNMUTE_FILE" ]]; then
  :
elif [[ -f "$PROJECT_MUTE_FILE" ]]; then
  if [[ -f "$GLOBAL_MUTE_FILE" ]]; then
    echo "🔇 TTS muted (project + global)"
  else
    echo "🔇 TTS muted (project)"
  fi
  exit 0
elif [[ -f "$GLOBAL_MUTE_FILE" ]]; then
  echo "🔇 TTS muted (global)"
  exit 0
fi

TEXT="$1"
VOICE_OVERRIDE="$2"

if [[ -z "$TEXT" ]]; then
  echo "Error: No text provided" >&2
  exit 1
fi

if [[ -n "$VOICE_OVERRIDE" ]] && [[ "$VOICE_OVERRIDE" =~ [';|&$`<>(){}'] ]]; then
  echo "Error: Invalid characters in voice parameter" >&2
  exit 1
fi

TEXT="${TEXT//\\!/!}"
TEXT="${TEXT//\\\$/\$}"
TEXT="${TEXT//\\?/?}"
TEXT="${TEXT//\\,/,}"
TEXT="${TEXT//\\./.}"
TEXT="${TEXT//\\\\/\\}"

source "$SCRIPT_DIR/provider-manager.sh"

ACTIVE_PROVIDER=$(get_active_provider)

"$SCRIPT_DIR/github-star-reminder.sh" 2>/dev/null || true

detect_voice_provider() {
  local voice="$1"
  if [[ "$voice" == *"_"*"-"* ]]; then
    echo "piper"
  else
    echo "$ACTIVE_PROVIDER"
  fi
}

if [[ -n "$VOICE_OVERRIDE" ]]; then
  DETECTED_PROVIDER=$(detect_voice_provider "$VOICE_OVERRIDE")
  if [[ "$DETECTED_PROVIDER" != "$ACTIVE_PROVIDER" ]]; then
    ACTIVE_PROVIDER="$DETECTED_PROVIDER"
  fi
fi

speak_text() {
  local text="$1"
  local voice="${2:-}"
  local provider="${3:-$ACTIVE_PROVIDER}"

  case "$provider" in
    piper)
      "$SCRIPT_DIR/play-tts-piper.sh" "$text" "$voice"
      ;;
    macos)
      "$SCRIPT_DIR/play-tts-macos.sh" "$text" "$voice"
      ;;
    termux-ssh)
      "$SCRIPT_DIR/play-tts-termux-ssh.sh" "$text" "$voice"
      ;;
    *)
      echo "❌ Unknown provider: $provider" >&2
      return 1
      ;;
  esac
}

handle_learning_mode() {
  source "$SCRIPT_DIR/learn-manager.sh" 2>/dev/null || return 1

  if ! is_learn_mode_enabled 2>/dev/null; then
    return 1
  fi

  local target_lang
  target_lang=$(get_target_language 2>/dev/null || echo "")
  local target_voice
  target_voice=$(get_target_voice 2>/dev/null || echo "")

  if [[ -z "$target_lang" ]] || [[ -z "$target_voice" ]]; then
    return 1
  fi

  speak_text "$TEXT" "$VOICE_OVERRIDE" "$ACTIVE_PROVIDER"

  local translated
  translated=$(python3 "$SCRIPT_DIR/translator.py" "$TEXT" "$target_lang" 2>/dev/null) || translated="$TEXT"

  sleep 0.5

  local target_provider
  target_provider=$(detect_voice_provider "$target_voice")
  speak_text "$translated" "$target_voice" "$target_provider"

  return 0
}

handle_translation_mode() {
  source "$SCRIPT_DIR/translate-manager.sh" 2>/dev/null || return 1

  if ! is_translation_enabled 2>/dev/null; then
    return 1
  fi

  local translate_to
  translate_to=$(get_translate_to 2>/dev/null || echo "")

  if [[ -z "$translate_to" ]] || [[ "$translate_to" == "english" ]]; then
    return 1
  fi

  local translated
  translated=$(python3 "$SCRIPT_DIR/translator.py" "$TEXT" "$translate_to" 2>/dev/null) || translated="$TEXT"

  local voice_to_use="$VOICE_OVERRIDE"
  if [[ -z "$voice_to_use" ]]; then
    source "$SCRIPT_DIR/language-manager.sh" 2>/dev/null || true
    voice_to_use=$(get_voice_for_language "$translate_to" "$ACTIVE_PROVIDER" 2>/dev/null || echo "")
  fi

  local provider_to_use="$ACTIVE_PROVIDER"
  if [[ -n "$voice_to_use" ]]; then
    provider_to_use=$(detect_voice_provider "$voice_to_use")
  fi

  speak_text "$translated" "$voice_to_use" "$provider_to_use"
  return 0
}

if handle_learning_mode; then
  exit 0
fi

if handle_translation_mode; then
  exit 0
fi

case "$ACTIVE_PROVIDER" in
  piper)
    exec "$SCRIPT_DIR/play-tts-piper.sh" "$TEXT" "$VOICE_OVERRIDE"
    ;;
  macos)
    exec "$SCRIPT_DIR/play-tts-macos.sh" "$TEXT" "$VOICE_OVERRIDE"
    ;;
  termux-ssh)
    exec "$SCRIPT_DIR/play-tts-termux-ssh.sh" "$TEXT" "$VOICE_OVERRIDE"
    ;;
  *)
    echo "❌ Unknown provider: $ACTIVE_PROVIDER"
    echo "   Run: /agent-vibes:provider list"
    exit 1
    ;;
esac
