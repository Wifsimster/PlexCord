#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

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

LANGUAGE_FILE="$CLAUDE_DIR/tts-language.txt"
mkdir -p "$CLAUDE_DIR"

source "$SCRIPT_DIR/provider-manager.sh" 2>/dev/null || true


_get_piper_voice() {
    local lang="$1"
    case "$lang" in
        spanish) echo "es_ES-davefx-medium" ;;
        french) echo "fr_FR-siwis-medium" ;;
        german) echo "de_DE-thorsten-medium" ;;
        italian) echo "it_IT-riccardo-x_low" ;;
        portuguese) echo "pt_BR-faber-medium" ;;
        chinese) echo "zh_CN-huayan-medium" ;;
        japanese) echo "ja_JP-hikari-medium" ;;
        korean) echo "ko_KR-eunyoung-medium" ;;
        russian) echo "ru_RU-dmitri-medium" ;;
        polish) echo "pl_PL-darkman-medium" ;;
        dutch) echo "nl_NL-rdh-medium" ;;
        turkish) echo "tr_TR-dfki-medium" ;;
        arabic) echo "ar_JO-kareem-medium" ;;
        hindi) echo "hi_IN-amitabh-medium" ;;
        swedish) echo "sv_SE-nst-medium" ;;
        danish) echo "da_DK-talesyntese-medium" ;;
        norwegian) echo "no_NO-talesyntese-medium" ;;
        finnish) echo "fi_FI-harri-medium" ;;
        czech) echo "cs_CZ-jirka-medium" ;;
        romanian) echo "ro_RO-mihai-medium" ;;
        ukrainian) echo "uk_UA-lada-x_low" ;;
        greek) echo "el_GR-rapunzelina-low" ;;
        bulgarian) echo "bg_BG-valentin-medium" ;;
        croatian) echo "hr_HR-gorana-medium" ;;
        slovak) echo "sk_SK-lili-medium" ;;
        *) echo "" ;;
    esac
}

_get_language_voice() {
    _get_piper_voice "$1"
}

_is_language_supported() {
    local lang="$1"
    local voice
    voice=$(_get_piper_voice "$lang")
    [[ -n "$voice" ]]
}

SUPPORTED_LANGUAGES="spanish, french, german, italian, portuguese, chinese, japanese, korean, polish, dutch, turkish, russian, arabic, hindi, swedish, danish, norwegian, finnish, czech, romanian, ukrainian, greek, bulgarian, croatian, slovak"

set_language() {
    local lang="$1"

    lang=$(echo "$lang" | tr '[:upper:]' '[:lower:]')

    if [[ "$lang" == "reset" ]] || [[ "$lang" == "english" ]] || [[ "$lang" == "en" ]]; then
        if [[ -f "$LANGUAGE_FILE" ]]; then
            rm "$LANGUAGE_FILE"
            echo "✓ Language reset to English (default)"
        else
            echo "Already using English (default)"
        fi
        return 0
    fi

    if ! _is_language_supported "$lang"; then
        echo "❌ Language '$lang' not supported"
        echo ""
        echo "Supported languages:"
        echo "$SUPPORTED_LANGUAGES"
        return 1
    fi

    echo "$lang" > "$LANGUAGE_FILE"

    local provider=""
    if [[ -f "$CLAUDE_DIR/tts-provider.txt" ]]; then
        provider=$(cat "$CLAUDE_DIR/tts-provider.txt")
    elif [[ -f "$HOME/.claude/tts-provider.txt" ]]; then
        provider=$(cat "$HOME/.claude/tts-provider.txt")
    else
        provider="piper"
    fi

    local recommended_voice
    recommended_voice=$(get_voice_for_language "$lang" "$provider")

    if [[ -z "$recommended_voice" ]]; then
        recommended_voice=$(_get_language_voice "$lang")
    fi

    echo "✓ Language set to: $lang"
    echo "📢 Recommended voice for $provider TTS: $recommended_voice"
    echo ""
    echo "TTS will now speak in $lang."
    echo "Switch voice with: /agent-vibes:switch \"$recommended_voice\""
}

get_language() {
    if [[ -f "$LANGUAGE_FILE" ]]; then
        local lang=$(cat "$LANGUAGE_FILE")

        local provider=""
        if [[ -f "$CLAUDE_DIR/tts-provider.txt" ]]; then
            provider=$(cat "$CLAUDE_DIR/tts-provider.txt")
        elif [[ -f "$HOME/.claude/tts-provider.txt" ]]; then
            provider=$(cat "$HOME/.claude/tts-provider.txt")
        else
            provider="piper"
        fi

        local recommended_voice
        recommended_voice=$(get_voice_for_language "$lang" "$provider")

        if [[ -z "$recommended_voice" ]]; then
            recommended_voice=$(_get_language_voice "$lang")
        fi

        echo "Current language: $lang"
        echo "Recommended voice ($provider): $recommended_voice"
    else
        echo "Current language: english (default)"
        echo "No multilingual voice required"
    fi
}

get_language_code() {
    if [[ -f "$LANGUAGE_FILE" ]]; then
        cat "$LANGUAGE_FILE"
    else
        echo "english"
    fi
}

is_voice_multilingual() {
    local voice="$1"

    local multilingual_voices=("Antoni" "Rachel" "Domi" "Bella" "Charlotte" "Matilda")

    for mv in "${multilingual_voices[@]}"; do
        if [[ "$voice" == "$mv" ]]; then
            return 0
        fi
    done

    return 1
}

get_best_voice_for_language() {
    local lang=$(get_language_code)

    if [[ "$lang" == "english" ]]; then
        echo ""
        return
    fi

    _get_language_voice "$lang"
}

get_voice_for_language() {
    local language="$1"
    local provider="${2:-}"

    language=$(echo "$language" | tr '[:upper:]' '[:lower:]')

    if [[ -z "$provider" ]]; then
        if command -v get_active_provider &>/dev/null; then
            provider=$(get_active_provider 2>/dev/null)
        else
            local search_dir="$PWD"
            local found=false

            while [[ "$search_dir" != "/" ]]; do
                if [[ -f "$search_dir/.claude/tts-provider.txt" ]]; then
                    provider=$(cat "$search_dir/.claude/tts-provider.txt")
                    found=true
                    break
                fi
                search_dir=$(dirname "$search_dir")
            done

            if [[ "$found" = false ]]; then
                if [[ -f "$HOME/.claude/tts-provider.txt" ]]; then
                    provider=$(cat "$HOME/.claude/tts-provider.txt")
                else
                    provider="piper"
                fi
            fi
        fi
    fi

    case "$provider" in
        piper)
            _get_piper_voice "$language"
            ;;
        macos)
            echo ""
            ;;
        *)
            _get_piper_voice "$language"
            ;;
    esac
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    case "${1:-}" in
        set)
            if [[ -z "$2" ]]; then
                echo "Usage: language-manager.sh set <language>"
                exit 1
            fi
            set_language "$2"
            ;;
        get)
            get_language
            ;;
        code)
            get_language_code
            ;;
        check-voice)
            if [[ -z "$2" ]]; then
                echo "Usage: language-manager.sh check-voice <voice-name>"
                exit 1
            fi
            if is_voice_multilingual "$2"; then
                echo "yes"
            else
                echo "no"
            fi
            ;;
        best-voice)
            get_best_voice_for_language
            ;;
        voice-for-language)
            if [[ -z "$2" ]]; then
                echo "Usage: language-manager.sh voice-for-language <language> [provider]"
                exit 1
            fi
            get_voice_for_language "$2" "$3"
            ;;
        list)
            echo "Supported languages and recommended voices:"
            echo ""
            for lang in spanish french german italian portuguese chinese japanese korean russian polish dutch turkish arabic hindi swedish danish norwegian finnish czech romanian ukrainian greek bulgarian croatian slovak; do
                printf "%-15s → %s\n" "$lang" "$(_get_language_voice "$lang")"
            done | sort
            ;;
        *)
            echo "AgentVibes Language Manager"
            echo ""
            echo "Usage:"
            echo "  language-manager.sh set <language>                    Set language"
            echo "  language-manager.sh get                               Get current language"
            echo "  language-manager.sh code                              Get language code only"
            echo "  language-manager.sh check-voice <name>                Check if voice is multilingual"
            echo "  language-manager.sh best-voice                        Get best voice for current language"
            echo "  language-manager.sh voice-for-language <lang> [prov] Get voice for language & provider"
            echo "  language-manager.sh list                              List all supported languages"
            exit 1
            ;;
    esac
fi
