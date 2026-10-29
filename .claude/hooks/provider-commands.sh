#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/provider-manager.sh"
source "$SCRIPT_DIR/language-manager.sh"

COMMAND="${1:-help}"

is_language_supported() {
  local language="$1"
  local provider="$2"

  if [[ "$language" == "english" ]] || [[ "$language" == "en" ]]; then
    return 0
  fi

  case "$provider" in
    piper)
      return 1
      ;;
    macos)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

provider_list() {
  local current_provider
  current_provider=$(get_active_provider)

  local is_macos=false
  if [[ "$(uname -s)" == "Darwin" ]]; then
    is_macos=true
  fi

  echo "┌────────────────────────────────────────────────────────────┐"
  echo "│ Available TTS Providers                                    │"
  echo "├────────────────────────────────────────────────────────────┤"

  if [[ "$is_macos" == true ]]; then
    if [[ "$current_provider" == "macos" ]]; then
      echo "│ ✓ macOS Say     Built-in, free     ⭐⭐⭐⭐      [ACTIVE]    │"
    else
      echo "│   macOS Say     Built-in, free     ⭐⭐⭐⭐   [RECOMMENDED] │"
    fi
    echo "│   Cost: Free (built-in)                                   │"
    echo "│   Platform: macOS only                                    │"
    echo "│   Offline: Yes                                            │"
    echo "│                                                            │"
  fi

  if [[ "$current_provider" == "piper" ]]; then
    echo "│ ✓ Piper TTS     Free, offline      ⭐⭐⭐⭐       [ACTIVE]    │"
  else
    echo "│   Piper TTS     Free, offline      ⭐⭐⭐⭐                  │"
  fi
  echo "│   Cost: Free forever                                       │"
  echo "│   Platform: WSL, Linux only                                │"
  echo "│   Offline: Yes                                             │"

  if [[ "$is_macos" != true ]]; then
    echo "│                                                            │"
    if [[ "$current_provider" == "macos" ]]; then
      echo "│ ✓ macOS Say     Built-in, free     ⭐⭐⭐⭐      [ACTIVE]    │"
    else
      echo "│   macOS Say     Built-in, free     ⭐⭐⭐⭐   (macOS only)  │"
    fi
    echo "│   Cost: Free (built-in)                                   │"
    echo "│   Platform: macOS only                                    │"
    echo "│   Offline: Yes                                            │"
  fi

  echo "└────────────────────────────────────────────────────────────┘"
  echo ""
  echo "Learn more: agentvibes.org/providers"
}

provider_switch() {
  local new_provider="$1"
  local force_mode=false

  if [[ "$2" == "--force" ]] || [[ "$2" == "--yes" ]] || [[ "$2" == "-y" ]]; then
    force_mode=true
  fi

  if [[ ! -t 0 ]] || [[ -n "$CLAUDE_PROJECT_DIR" ]] || [[ -n "$MCP_SERVER" ]]; then
    force_mode=true
  fi

  if [[ -z "$new_provider" ]]; then
    echo "❌ Error: Provider name required"
    echo "Usage: /agent-vibes:provider switch <provider> [--force]"
    echo "Available: piper, macos"
    return 1
  fi

  if ! validate_provider "$new_provider"; then
    echo "❌ Invalid provider: $new_provider"
    echo ""
    echo "Available providers:"
    list_providers
    return 1
  fi

  local current_provider
  current_provider=$(get_active_provider)

  if [[ "$current_provider" == "$new_provider" ]]; then
    echo "✓ Already using $new_provider"
    return 0
  fi

  if [[ "$new_provider" == "piper" ]]; then
    if ! grep -qi microsoft /proc/version 2>/dev/null && [[ "$(uname -s)" != "Linux" ]]; then
      echo "❌ Piper is only supported on WSL and Linux"
      echo "Your platform: $(uname -s)"
      echo "See: agentvibes.org/platform-support"
      return 1
    fi

    if ! command -v piper &> /dev/null; then
      echo "❌ Piper TTS is not installed"
      echo ""
      echo "Install with: pipx install piper-tts"
      echo "Or run: .claude/hooks/piper-installer.sh"
      echo ""
      echo "Visit: agentvibes.org/install-piper"
      return 1
    fi
  fi

  if [[ "$new_provider" == "macos" ]]; then
    if [[ "$(uname -s)" != "Darwin" ]]; then
      echo "❌ macOS Say provider is only supported on macOS"
      echo "Your platform: $(uname -s)"
      echo ""
      echo "Alternative providers:"
      echo "  • Piper TTS (free, offline) - for Linux/WSL"
      echo ""
      echo "See: agentvibes.org/platform-support"
      return 1
    fi
  fi

  local current_language
  current_language=$(get_language_code)

  if [[ "$current_language" != "english" ]]; then
    if ! is_language_supported "$current_language" "$new_provider" 2>/dev/null; then
      echo "⚠️  Language Compatibility Warning"
      echo ""
      echo "Current language: $current_language"
      echo "Target provider:  $new_provider"
      echo ""
      echo "❌ Language '$current_language' is not natively supported by $new_provider"
      echo "   Will fall back to English when using $new_provider"
      echo ""
      echo "Options:"
      echo "  1. Continue anyway (will use English)"
      echo "  2. Switch language to English"
      echo "  3. Cancel provider switch"
      echo ""

      if [[ "$force_mode" == true ]]; then
        echo "⏩ Force mode: Continuing with fallback to English..."
      else
        read -p "Choose option [1-3]: " -n 1 -r
        echo

        case $REPLY in
          1)
            echo "⏩ Continuing with fallback to English..."
            ;;
          2)
            echo "🔄 Switching language to English..."
            "$SCRIPT_DIR/language-manager.sh" set english
            ;;
          3)
            echo "❌ Provider switch cancelled"
            return 1
            ;;
          *)
            echo "❌ Invalid option, cancelling"
            return 1
            ;;
        esac
      fi
    fi
  fi

  if [[ "$force_mode" != true ]]; then
    echo ""
    echo "⚠️  Switch to $(echo $new_provider | tr '[:lower:]' '[:upper:]')?"
    echo ""
    echo "Current: $current_provider"
    echo "New:     $new_provider"
    if [[ "$current_language" != "english" ]]; then
      echo "Language: $current_language"
    fi
    echo ""
    read -p "Continue? [y/N]: " -n 1 -r
    echo

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "❌ Switch cancelled"
      return 1
    fi
  else
    echo "⏩ Force mode: Switching to $new_provider..."
  fi

  set_active_provider "$new_provider"

  local target_lang_file=""
  local target_voice_file=""

  if [[ -d "$SCRIPT_DIR/../.." ]]; then
    local project_dir="$SCRIPT_DIR/../.."
    if [[ -f "$project_dir/.claude/tts-target-language.txt" ]]; then
      target_lang_file="$project_dir/.claude/tts-target-language.txt"
      target_voice_file="$project_dir/.claude/tts-target-voice.txt"
    fi
  fi

  if [[ -z "$target_lang_file" ]]; then
    if [[ -f "$HOME/.claude/tts-target-language.txt" ]]; then
      target_lang_file="$HOME/.claude/tts-target-language.txt"
      target_voice_file="$HOME/.claude/tts-target-voice.txt"
    fi
  fi

  if [[ -n "$target_lang_file" ]] && [[ -f "$target_lang_file" ]]; then
    local target_lang
    target_lang=$(cat "$target_lang_file")

    if [[ -n "$target_lang" ]]; then
      local new_target_voice
      new_target_voice=$(get_voice_for_language "$target_lang" "$new_provider")

      if [[ -n "$new_target_voice" ]]; then
        echo "$new_target_voice" > "$target_voice_file"
        echo ""
        echo "🔄 Updated target language voice:"
        echo "   Language: $target_lang"
        echo "   Voice: $new_target_voice (for $new_provider)"
      fi
    fi
  fi

  echo ""
  echo "🔊 Testing provider..."
  "$SCRIPT_DIR/play-tts.sh" "Provider switched to $new_provider successfully" 2>/dev/null

  echo ""
  echo "✓ Provider switch complete!"
  echo "Visit agentvibes.org for tips and tricks"
}

provider_info() {
  local provider_name="$1"

  if [[ -z "$provider_name" ]]; then
    echo "❌ Error: Provider name required"
    echo "Usage: /agent-vibes:provider info <provider>"
    return 1
  fi

  case "$provider_name" in
    piper)
      echo "┌────────────────────────────────────────────────────────────┐"
      echo "│ Piper TTS - Free Offline Provider                          │"
      echo "├────────────────────────────────────────────────────────────┤"
      echo "│ Quality:     ⭐⭐⭐⭐  (Very good)                            │"
      echo "│ Cost:        Free forever                                  │"
      echo "│ Platform:    WSL, Linux only                               │"
      echo "│ Offline:     Yes (fully local)                             │"
      echo "│                                                            │"
      echo "│ Trade-offs:                                                │"
      echo "│ + Completely free, no API costs                           │"
      echo "│ + Works offline, no internet needed                       │"
      echo "│ + Fast synthesis (local processing)                       │"
      echo "│ - WSL/Linux only (no macOS/Windows)                       │"
      echo "│                                                            │"
      echo "│ Best for: Budget-conscious, offline use, privacy          │"
      echo "└────────────────────────────────────────────────────────────┘"
      echo ""
      echo "Full comparison: agentvibes.org/providers"
      ;;

    macos)
      echo "┌────────────────────────────────────────────────────────────┐"
      echo "│ macOS Say - Native macOS TTS Provider                      │"
      echo "├────────────────────────────────────────────────────────────┤"
      echo "│ Quality:     ⭐⭐⭐⭐  (Very good, Siri-quality on Mojave+)   │"
      echo "│ Cost:        Free (built-in)                               │"
      echo "│ Platform:    macOS only (all versions since 10.0)          │"
      echo "│ Offline:     Yes (fully local)                             │"
      echo "│                                                            │"
      echo "│ Trade-offs:                                                │"
      echo "│ + Zero setup - built into every Mac                       │"
      echo "│ + 100+ voices in 40+ languages                            │"
      echo "│ + Completely free, no API costs                           │"
      echo "│ + Works offline, no internet needed                       │"
      echo "│ - macOS only (no Windows/Linux)                           │"
      echo "│                                                            │"
      echo "│ Popular voices: Samantha, Alex, Daniel, Victoria          │"
      echo "│                                                            │"
      echo "│ Best for: Mac users wanting free, zero-setup TTS          │"
      echo "└────────────────────────────────────────────────────────────┘"
      echo ""
      echo "Full comparison: agentvibes.org/providers"
      ;;

    *)
      echo "❌ Unknown provider: $provider_name"
      echo "Available: piper, macos"
      ;;
  esac
}

provider_test() {
  local current_provider
  current_provider=$(get_active_provider)

  echo "🔊 Testing provider: $current_provider"
  echo ""

  "$SCRIPT_DIR/play-tts.sh" "Provider test successful. Audio is working correctly with $current_provider."

  echo ""
  echo "✓ Test complete"
}

provider_get() {
  local current_provider
  current_provider=$(get_active_provider)

  echo "🎤 Current Provider: $current_provider"
  echo ""

  case "$current_provider" in
    piper)
      echo "Quality: ⭐⭐⭐⭐"
      echo "Cost: Free forever"
      echo "Offline: Yes"
      ;;
    macos)
      echo "Quality: ⭐⭐⭐⭐"
      echo "Cost: Free (built-in)"
      echo "Offline: Yes"
      ;;
  esac

  echo ""
  echo "Use /agent-vibes:provider info $current_provider for details"
}

provider_preview() {
  local current_provider
  current_provider=$(get_active_provider)

  echo "🎤 Voice Preview ($current_provider)"
  echo ""

  case "$current_provider" in
    piper)
      source "$SCRIPT_DIR/piper-voice-manager.sh"

      local voice_arg="$1"

      if [[ -n "$voice_arg" ]]; then

        if [[ "$voice_arg" =~ ^[a-z]{2}_[A-Z]{2}- ]]; then
          if verify_voice "$voice_arg"; then
            echo "🎤 Previewing Piper voice: $voice_arg"
            echo ""
            "$SCRIPT_DIR/play-tts.sh" "Hello, this is the $voice_arg voice. How do you like it?" "$voice_arg"
          else
            echo "❌ Voice model not found: $voice_arg"
            echo ""
            echo "💡 Piper voice names look like: en_US-lessac-medium"
            echo "   Run /agent-vibes:list to see available Piper voices"
          fi
        else
          echo "❌ Invalid voice format: '$voice_arg'"
          echo ""
          echo "💡 Piper voice names look like: en_US-lessac-medium"
          echo "   Run /agent-vibes:list to see available Piper voices"
          echo ""
          echo "Popular Piper voices to try:"
          echo "  • en_US-lessac-medium  (clear, professional)"
          echo "  • en_US-amy-medium     (warm, friendly)"
          echo "  • en_US-joe-medium     (casual, natural)"
        fi
        return
      fi

      echo "🎤 Piper Preview of 3 people"
      echo ""

      local sample_voices=(
        "en_US-lessac-medium:Lessac"
        "en_US-amy-medium:Amy"
        "en_US-joe-medium:Joe"
      )

      for voice_entry in "${sample_voices[@]}"; do
        local voice_name="${voice_entry%%:*}"
        local display_name="${voice_entry##*:}"

        echo "🔊 ${display_name}..."
        "$SCRIPT_DIR/play-tts.sh" "Hi, my name is ${display_name}" "$voice_name"

        sleep 3
      done

      echo ""
      echo "✓ Preview complete"
      echo "💡 Use /agent-vibes:list to see all available Piper voices"
      ;;
    macos)
      if [[ "$(uname -s)" != "Darwin" ]]; then
        echo "❌ macOS voices only available on macOS"
        echo "Your platform: $(uname -s)"
        return 1
      fi

      local voice_arg="$1"

      if [[ -n "$voice_arg" ]]; then
        if say -v ? 2>/dev/null | grep -qi "^${voice_arg} "; then
          echo "🎤 Previewing macOS voice: $voice_arg"
          echo ""
          "$SCRIPT_DIR/play-tts.sh" "Hello, this is ${voice_arg}. How do you like my voice?" "$voice_arg"
        else
          echo "❌ Voice not found: $voice_arg"
          echo ""
          echo "Available macOS voices (showing English):"
          say -v ? 2>/dev/null | grep -i "en_" | head -10
          echo ""
          echo "Run /agent-vibes:list to see all available voices"
        fi
        return
      fi

      echo "🎤 macOS Preview of 3 voices"
      echo ""

      local sample_voices=("Samantha" "Alex" "Daniel")

      for voice in "${sample_voices[@]}"; do
        if say -v ? 2>/dev/null | grep -qi "^${voice} "; then
          echo "🔊 ${voice}..."
          "$SCRIPT_DIR/play-tts.sh" "Hi, my name is ${voice}" "$voice"
          sleep 3
        fi
      done

      echo ""
      echo "✓ Preview complete"
      echo "💡 Use /agent-vibes:list to see all available macOS voices"
      ;;

    *)
      echo "❌ Unknown provider: $current_provider"
      ;;
  esac
}

provider_help() {
  echo "Provider Management Commands"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Usage:"
  echo "  /agent-vibes:provider list              # Show all providers"
  echo "  /agent-vibes:provider switch <name>     # Switch provider"
  echo "  /agent-vibes:provider info <name>       # Provider details"
  echo "  /agent-vibes:provider test              # Test current provider"
  echo "  /agent-vibes:provider get               # Show active provider"
  echo ""
  echo "Examples:"
  echo "  /agent-vibes:provider switch piper"
  echo "  /agent-vibes:provider info piper"
  echo ""
  echo "Learn more: agentvibes.org/docs/providers"
}

case "$COMMAND" in
  list)
    provider_list
    ;;
  switch)
    provider_switch "$2" "$3"
    ;;
  info)
    provider_info "$2"
    ;;
  test)
    provider_test
    ;;
  get)
    provider_get
    ;;
  preview)
    shift
    provider_preview "$@"
    ;;
  help|*)
    provider_help
    ;;
esac
