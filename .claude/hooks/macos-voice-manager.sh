#!/usr/bin/env bash

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "❌ Error: macOS voice manager only works on macOS"
  exit 1
fi

list_macos_voices() {
  say -v ? 2>/dev/null | awk '{print $1}'
}

list_macos_voices_detailed() {
  say -v ? 2>/dev/null | while read -r line; do
    local voice=$(echo "$line" | awk '{print $1}')
    local lang=$(echo "$line" | awk '{print $2}')
    echo "$voice $lang"
  done
}

list_english_voices() {
  say -v ? 2>/dev/null | grep -i "en_" | awk '{print $1}'
}

get_voice_language() {
  local voice="$1"
  say -v ? 2>/dev/null | grep -i "^${voice} " | awk '{print $2}'
}

validate_voice() {
  local voice="$1"
  say -v ? 2>/dev/null | grep -qi "^${voice} "
}

get_recommended_voices() {
  local recommended=("Samantha" "Alex" "Daniel" "Karen" "Moira" "Tessa" "Fiona" "Veena" "Victoria")

  for voice in "${recommended[@]}"; do
    if validate_voice "$voice"; then
      echo "$voice"
    fi
  done
}

get_voice_info() {
  local voice="$1"

  if ! validate_voice "$voice"; then
    echo "❌ Voice not found: $voice"
    return 1
  fi

  local full_info=$(say -v ? 2>/dev/null | grep -i "^${voice} ")
  local lang=$(echo "$full_info" | awk '{print $2}')

  echo "Voice: $voice"
  echo "Language: $lang"
  echo "Provider: macOS Say (built-in)"

  case "$voice" in
    Alex) echo "Description: American English male (enhanced)" ;;
    Samantha) echo "Description: American English female (enhanced)" ;;
    Victoria) echo "Description: American English female" ;;
    Daniel) echo "Description: British English male (enhanced)" ;;
    Karen) echo "Description: Australian English female (enhanced)" ;;
    Moira) echo "Description: Irish English female (enhanced)" ;;
    Tessa) echo "Description: South African English female (enhanced)" ;;
    Fiona) echo "Description: Scottish English female (enhanced)" ;;
    Veena) echo "Description: Indian English female (enhanced)" ;;
  esac
}

count_voices() {
  say -v ? 2>/dev/null | wc -l | tr -d ' '
}

count_english_voices() {
  say -v ? 2>/dev/null | grep -i "en_" | wc -l | tr -d ' '
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  case "${1:-}" in
    list)
      echo "🎤 Available macOS Voices:"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      list_macos_voices_detailed | while read -r voice lang; do
        printf "  %-15s %s\n" "$voice" "$lang"
      done
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo "Total: $(count_voices) voices"
      ;;
    list-english)
      echo "🎤 English macOS Voices:"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      list_english_voices
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      echo "Total: $(count_english_voices) English voices"
      ;;
    recommended)
      echo "🌟 Recommended macOS Voices:"
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      get_recommended_voices
      echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
      ;;
    info)
      if [[ -z "$2" ]]; then
        echo "Usage: $0 info <voice_name>"
        exit 1
      fi
      get_voice_info "$2"
      ;;
    validate)
      if [[ -z "$2" ]]; then
        echo "Usage: $0 validate <voice_name>"
        exit 1
      fi
      if validate_voice "$2"; then
        echo "✅ Voice '$2' is available"
        exit 0
      else
        echo "❌ Voice '$2' is not available"
        exit 1
      fi
      ;;
    *)
      echo "Usage: $0 {list|list-english|recommended|info|validate} [voice_name]"
      echo ""
      echo "Commands:"
      echo "  list           - List all available voices"
      echo "  list-english   - List English voices only"
      echo "  recommended    - List recommended high-quality voices"
      echo "  info <voice>   - Get details about a voice"
      echo "  validate <voice> - Check if voice exists"
      exit 1
      ;;
  esac
fi
