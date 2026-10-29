#!/usr/bin/env bash

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    set -euo pipefail
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -d "$PWD/.claude" ]]; then
    PROJECT_DIR="$PWD"
else
    PROJECT_DIR="$SCRIPT_DIR/../.."
fi

TRANSLATE_FILE="$PROJECT_DIR/.claude/tts-translate-to.txt"
GLOBAL_TRANSLATE_FILE="$HOME/.claude/tts-translate-to.txt"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

SUPPORTED_LANGUAGES="spanish french german italian portuguese chinese japanese korean russian polish dutch turkish arabic hindi swedish danish norwegian finnish czech romanian ukrainian greek bulgarian croatian slovak"

get_translate_to() {
    if [[ -f "$TRANSLATE_FILE" ]]; then
        local manual
        manual=$(cat "$TRANSLATE_FILE")
        if [[ "$manual" != "off" ]]; then
            echo "$manual"
            return
        elif [[ "$manual" == "off" ]]; then
            echo ""
            return
        fi
    elif [[ -f "$GLOBAL_TRANSLATE_FILE" ]]; then
        local manual
        manual=$(cat "$GLOBAL_TRANSLATE_FILE")
        if [[ "$manual" != "off" ]]; then
            echo "$manual"
            return
        elif [[ "$manual" == "off" ]]; then
            echo ""
            return
        fi
    fi

    echo ""
}

is_translation_enabled() {
    local translate_to
    translate_to=$(get_translate_to)
    [[ -n "$translate_to" ]] && [[ "$translate_to" != "english" ]]
}

translate_text() {
    local text="$1"
    local target="${2:-}"

    if [[ -z "$target" ]]; then
        target=$(get_translate_to)
    fi

    if [[ -z "$target" ]] || [[ "$target" == "english" ]]; then
        echo "$text"
        return
    fi

    local translated
    translated=$(python3 "$SCRIPT_DIR/translator.py" "$text" "$target" 2>/dev/null) || translated="$text"

    echo "$translated"
}

set_translate() {
    local lang="$1"

    if [[ -z "$lang" ]]; then
        echo -e "${YELLOW}Usage: translate-manager.sh set <language|off>${NC}"
        exit 1
    fi

    lang=$(echo "$lang" | tr '[:upper:]' '[:lower:]')

    mkdir -p "$PROJECT_DIR/.claude"

    if [[ "$lang" == "off" ]]; then
        echo "off" > "$TRANSLATE_FILE"
        echo -e "${GREEN}✓${NC} Translation: ${YELLOW}DISABLED${NC}"
        echo "  TTS will speak in English only"
        return
    fi

    local valid=false
    for supported in $SUPPORTED_LANGUAGES; do
        if [[ "$lang" == "$supported" ]]; then
            valid=true
            break
        fi
    done

    if [[ "$valid" != "true" ]]; then
        echo -e "${RED}❌${NC} Language '$lang' not supported"
        echo ""
        echo "Supported languages:"
        echo "$SUPPORTED_LANGUAGES" | tr ' ' '\n' | column
        exit 1
    fi

    echo "$lang" > "$TRANSLATE_FILE"
    echo -e "${GREEN}✓${NC} Translation set to: ${BLUE}$lang${NC}"
    echo "  All TTS will be translated to $lang before speaking"

    source "$SCRIPT_DIR/language-manager.sh" 2>/dev/null || true
    if command -v get_voice_for_language &>/dev/null; then
        local provider
        provider=$(get_active_provider 2>/dev/null || echo "piper")
        local voice
        voice=$(get_voice_for_language "$lang" "$provider" 2>/dev/null || echo "")
        if [[ -n "$voice" ]]; then
            echo -e "  ${BLUE}💡${NC} Recommended voice: ${YELLOW}$voice${NC}"
            echo -e "     Switch with: /agent-vibes:switch $voice"
        fi
    fi
}

show_status() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}   Translation Settings${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    local setting=""
    if [[ -f "$TRANSLATE_FILE" ]]; then
        setting=$(cat "$TRANSLATE_FILE")
    elif [[ -f "$GLOBAL_TRANSLATE_FILE" ]]; then
        setting=$(cat "$GLOBAL_TRANSLATE_FILE")
    fi

    local effective
    effective=$(get_translate_to)

    echo -e "  ${BLUE}Setting:${NC}    ${setting:-"(not set)"}"
    echo -e "  ${BLUE}Effective:${NC}  $(if [[ -n "$effective" ]]; then echo -e "${GREEN}$effective${NC}"; else echo -e "${YELLOW}No translation${NC}"; fi)"
    echo ""

    if [[ -n "$effective" ]]; then
        echo -e "  ${GREEN}✓${NC} TTS will be translated to ${BLUE}$effective${NC}"
    else
        echo -e "  ${YELLOW}ℹ${NC}  TTS will speak in English"
    fi

    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "  ${BLUE}Commands:${NC}"
    echo -e "    /agent-vibes:translate set <lang>  Set translation language"
    echo -e "    /agent-vibes:translate off         Disable translation"
    echo ""
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
case "${1:-}" in
    get-translate-to)
        get_translate_to
        ;;
    is-enabled)
        if is_translation_enabled; then
            echo "ON"
            exit 0
        else
            echo "OFF"
            exit 1
        fi
        ;;
    translate)
        if [[ -z "${2:-}" ]]; then
            echo "Usage: translate-manager.sh translate <text> [target_lang]" >&2
            exit 1
        fi
        translate_text "$2" "${3:-}"
        ;;
    set)
        set_translate "${2:-}"
        ;;
    off)
        set_translate "off"
        ;;
    status)
        show_status
        ;;
    *)
        show_status
        ;;
esac
fi
