#!/usr/bin/env bash

get_provider_config_path() {
  local provider_file

  if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -d "$CLAUDE_PROJECT_DIR/.claude" ]]; then
    provider_file="$CLAUDE_PROJECT_DIR/.claude/tts-provider.txt"
  else
    local current_dir="$PWD"
    while [[ "$current_dir" != "/" ]]; do
      if [[ -d "$current_dir/.claude" ]]; then
        provider_file="$current_dir/.claude/tts-provider.txt"
        break
      fi
      current_dir=$(dirname "$current_dir")
    done

    if [[ -z "$provider_file" ]]; then
      provider_file="$HOME/.claude/tts-provider.txt"
    fi
  fi

  echo "$provider_file"
}

get_active_provider() {
  local provider_file
  provider_file=$(get_provider_config_path)

  if [[ -f "$provider_file" ]]; then
    local provider
    provider=$(cat "$provider_file" | tr -d '[:space:]')
    if [[ -n "$provider" ]]; then
      echo "$provider"
      return 0
    fi
  fi

  echo "piper"
}

set_active_provider() {
  local provider="$1"

  if [[ -z "$provider" ]]; then
    echo "❌ Error: Provider name required"
    echo "Usage: set_active_provider <provider_name>"
    return 1
  fi

  if ! validate_provider "$provider"; then
    echo "❌ Error: Provider '$provider' not found"
    echo "Available providers:"
    list_providers
    return 1
  fi

  local provider_file
  provider_file=$(get_provider_config_path)

  mkdir -p "$(dirname "$provider_file")"

  echo "$provider" > "$provider_file"

  local voice_file
  if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -d "$CLAUDE_PROJECT_DIR/.claude" ]]; then
    voice_file="$CLAUDE_PROJECT_DIR/.claude/tts-voice.txt"
  else
    voice_file="$HOME/.claude/tts-voice.txt"
  fi

  local current_voice=""
  if [[ -f "$voice_file" ]]; then
    current_voice=$(cat "$voice_file" | tr -d '\n\r' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
  fi

  local new_voice
  new_voice=$(migrate_voice_to_provider "$current_voice" "$provider")

  echo "$new_voice" > "$voice_file"

  if [[ -n "$current_voice" ]] && [[ "$current_voice" != "$new_voice" ]]; then
    echo "✓ Active provider set to: $provider"
    echo "🔄 Voice migrated: $current_voice → $new_voice"
  else
    echo "✓ Active provider set to: $provider (voice: $new_voice)"
  fi
}

migrate_voice_to_provider() {
  local current_voice="$1"
  local target_provider="$2"

  local voice_mappings=(
    "en_US-amy-medium:Samantha"
    "en_US-ryan-high:Alex"
    "en_GB-alan-medium:Daniel"
    "en_US-kristin-medium:Victoria"
    "en_US-lessac-medium:Samantha"
    "en_US-joe-medium:Alex"
    "en_US-arctic-medium:Alex"
    "en_US-danny-low:Alex"
  )

  local piper_default="en_US-lessac-medium"
  local macos_default="Samantha"

  if [[ -z "$current_voice" ]]; then
    case "$target_provider" in
      piper) echo "$piper_default" ;;
      macos) echo "$macos_default" ;;
      *) echo "$piper_default" ;;
    esac
    return 0
  fi

  local current_voice_lower
  current_voice_lower=$(echo "$current_voice" | tr '[:upper:]' '[:lower:]')

  for mapping in "${voice_mappings[@]}"; do
    local piper_voice="${mapping%%:*}"
    local macos_voice="${mapping#*:}"

    local piper_voice_lower macos_voice_lower
    piper_voice_lower=$(echo "$piper_voice" | tr '[:upper:]' '[:lower:]')
    macos_voice_lower=$(echo "$macos_voice" | tr '[:upper:]' '[:lower:]')

    case "$target_provider" in
      piper)
        if [[ "$current_voice_lower" == "$macos_voice_lower" ]]; then
          echo "$piper_voice"
          return 0
        fi
        if [[ "$current_voice" =~ ^[a-z]{2}_ ]]; then
          echo "$current_voice"
          return 0
        fi
        ;;
      macos)
        if [[ "$current_voice_lower" == "$piper_voice_lower" ]]; then
          echo "$macos_voice"
          return 0
        fi
        if [[ "$current_voice" =~ ^[A-Z][a-z]+$ ]]; then
          echo "$current_voice"
          return 0
        fi
        ;;
    esac
  done

  case "$target_provider" in
    piper) echo "$piper_default" ;;
    macos) echo "$macos_default" ;;
    *) echo "$piper_default" ;;
  esac
}

list_providers() {
  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

  local providers=()
  shopt -s nullglob
  for file in "$script_dir"/play-tts-*.sh; do
    if [[ -f "$file" ]] && [[ "$file" != *"play-tts.sh" ]]; then
      local basename
      basename=$(basename "$file")
      local provider
      provider="${basename#play-tts-}"
      provider="${provider%.sh}"
      providers+=("$provider")
    fi
  done
  shopt -u nullglob

  if [[ ${#providers[@]} -eq 0 ]]; then
    echo "⚠️ No providers found"
    return 0
  fi

  for provider in "${providers[@]}"; do
    echo "$provider"
  done
}

validate_provider() {
  local provider="$1"

  if [[ -z "$provider" ]]; then
    return 1
  fi

  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  local provider_script="$script_dir/play-tts-${provider}.sh"

  [[ -f "$provider_script" ]]
}

get_provider_script_path() {
  local provider="$1"

  if [[ -z "$provider" ]]; then
    echo "❌ Error: Provider name required" >&2
    return 1
  fi

  local script_dir
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  local provider_script="$script_dir/play-tts-${provider}.sh"

  if [[ ! -f "$provider_script" ]]; then
    echo "❌ Error: Provider '$provider' not found at $provider_script" >&2
    return 1
  fi

  echo "$provider_script"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  case "${1:-}" in
    get)
      get_active_provider
      ;;
    switch|set)
      if [[ -z "${2:-}" ]]; then
        echo "❌ Error: Provider name required"
        echo "Usage: $0 switch <provider>"
        exit 1
      fi
      set_active_provider "$2"
      ;;
    list)
      list_providers
      ;;
    validate)
      if [[ -z "${2:-}" ]]; then
        echo "❌ Error: Provider name required"
        echo "Usage: $0 validate <provider>"
        exit 1
      fi
      validate_provider "$2"
      ;;
    *)
      echo "Usage: $0 {get|switch|list|validate} [provider]"
      echo ""
      echo "Commands:"
      echo "  get              - Show active provider"
      echo "  switch <name>    - Switch to provider"
      echo "  list             - List available providers"
      echo "  validate <name>  - Check if provider exists"
      exit 1
      ;;
  esac
fi
