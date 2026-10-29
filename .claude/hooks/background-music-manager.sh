#!/usr/bin/env bash

set -euo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

CONFIG_DIR="$SCRIPT_DIR/../config"
ENABLED_FILE="$CONFIG_DIR/background-music-enabled.txt"
VOLUME_FILE="$CONFIG_DIR/background-music-volume.txt"

DEFAULT_VOLUME="0.34"

mkdir -p "$CONFIG_DIR"

is_enabled() {
    if [[ -f "$ENABLED_FILE" ]]; then
        local status
        status=$(cat "$ENABLED_FILE" 2>/dev/null | tr -d '[:space:]')
        [[ "$status" == "true" || "$status" == "on" || "$status" == "1" ]]
    else
        return 1
    fi
}

get_volume() {
    if [[ -f "$VOLUME_FILE" ]]; then
        cat "$VOLUME_FILE" 2>/dev/null | tr -d '[:space:]'
    else
        echo "$DEFAULT_VOLUME"
    fi
}

set_enabled() {
    local value="$1"
    echo "$value" > "$ENABLED_FILE"
}

set_volume() {
    local value="$1"

    if ! [[ "$value" =~ ^[0-9]*\.?[0-9]+$ ]]; then
        echo "Error: Volume must be a number between 0.0 and 1.0" >&2
        return 1
    fi

    if command -v bc &> /dev/null; then
        if (( $(echo "$value < 0" | bc -l) )) || (( $(echo "$value > 1" | bc -l) )); then
            echo "Error: Volume must be between 0.0 and 1.0" >&2
            return 1
        fi
    fi

    echo "$value" > "$VOLUME_FILE"
}

list_tracks() {
    local bg_dir="$SCRIPT_DIR/../audio/tracks"

    if [[ ! -d "$bg_dir" ]]; then
        echo "❌ No tracks folder found at $bg_dir"
        return 1
    fi

    echo "🎵 Available Background Music Tracks"
    echo "===================================="
    echo ""

    local count=0
    while IFS= read -r -d '' file; do
        local basename
        basename=$(basename "$file")
        count=$((count + 1))
        printf "%2d. %s\n" "$count" "$basename"
    done < <(find "$bg_dir" -type f \( -name "*.mp3" -o -name "*.wav" -o -name "*.ogg" \) -print0 2>/dev/null | sort -z)

    if [[ $count -eq 0 ]]; then
        echo "No audio files found in tracks folder"
    else
        echo ""
        echo "Total: $count track(s)"
    fi
}

get_default_track() {
    local audio_effects_cfg="$SCRIPT_DIR/../config/audio-effects.cfg"

    if [[ ! -f "$audio_effects_cfg" ]]; then
        echo ""
        return
    fi

    grep "^default|" "$audio_effects_cfg" 2>/dev/null | cut -d'|' -f3 || echo ""
}

set_default_track() {
    local track="$1"
    local audio_effects_cfg="$SCRIPT_DIR/../config/audio-effects.cfg"
    local bg_dir="$SCRIPT_DIR/../audio/tracks"

    if [[ -z "$track" ]]; then
        echo "❌ Error: No track name provided"
        echo "Usage: $0 set-default TRACK_NAME"
        return 1
    fi

    if [[ ! -f "$bg_dir/$track" ]]; then
        echo "❌ Error: Track '$track' not found in tracks folder"
        echo "Run '$0 list' to see available tracks"
        return 1
    fi

    if [[ ! -f "$audio_effects_cfg" ]]; then
        echo "❌ Error: audio-effects.cfg not found at $audio_effects_cfg"
        return 1
    fi

    local temp_file
    temp_file=$(mktemp)

    if grep -q "^default|" "$audio_effects_cfg"; then
        awk -F'|' -v track="$track" '
            /^default\|/ {
                print $1 "|" $2 "|" track "|" $4
                next
            }
            { print }
        ' "$audio_effects_cfg" > "$temp_file"

        mv "$temp_file" "$audio_effects_cfg"

        if ! is_enabled; then
            set_enabled "true"
            echo "✅ Default background music set to: $track"
            echo "🎵 Background music enabled automatically"
        else
            echo "✅ Default background music set to: $track"
        fi
    else
        echo "❌ Error: No default entry found in audio-effects.cfg"
        return 1
    fi
}

show_status() {
    echo "🎵 Background Music Status"
    echo "=========================="

    if is_enabled; then
        echo "Status: ✅ ENABLED"
    else
        echo "Status: 🔇 DISABLED"
    fi

    echo "Volume: $(get_volume) ($(echo "scale=0; $(get_volume) * 100 / 1" | bc -l 2>/dev/null || echo "?")%)"

    local default_track
    default_track=$(get_default_track)
    if [[ -n "$default_track" ]]; then
        echo "Default Track: $default_track"
    else
        echo "Default Track: (none)"
    fi

    local bg_dir="$SCRIPT_DIR/../audio/tracks"
    if [[ -d "$bg_dir" ]]; then
        local count
        count=$(find "$bg_dir" -type f \( -name "*.mp3" -o -name "*.wav" -o -name "*.ogg" \) 2>/dev/null | wc -l)
        echo "Tracks: $count audio file(s) in tracks folder"
    else
        echo "Tracks: No tracks folder found"
    fi

    echo ""
    echo "Dependencies:"
    if command -v sox &> /dev/null; then
        echo "  sox: ✅ installed"
    else
        echo "  sox: ❌ not installed (needed for effects)"
    fi
    if command -v ffmpeg &> /dev/null; then
        echo "  ffmpeg: ✅ installed"
    else
        echo "  ffmpeg: ❌ not installed (needed for mixing)"
    fi
}

set_agent_track() {
    local agent="$1"
    local track="$2"
    local config_file="$SCRIPT_DIR/../config/audio-effects.cfg"

    if [[ -z "$agent" ]] || [[ -z "$track" ]]; then
        echo "❌ Error: Agent name and track required"
        echo "Usage: $0 set-agent AGENT_NAME TRACK_NAME"
        exit 1
    fi

    local bg_dir="$SCRIPT_DIR/../audio/tracks"
    if [[ ! -f "$bg_dir/$track" ]]; then
        echo "❌ Error: Track not found: $track"
        echo "Run '$0 list' to see available tracks"
        exit 1
    fi

    if grep -q "^${agent}|" "$config_file" 2>/dev/null; then
        local temp_file
        temp_file=$(mktemp)
        awk -F'|' -v agent="$agent" -v track="$track" '
            $1 == agent {
                print $1 "|" $2 "|" track "|" $4
                next
            }
            { print }
        ' "$config_file" > "$temp_file"
        mv "$temp_file" "$config_file"
        echo "✅ Updated background music for $agent: $track"
    else
        echo "${agent}||${track}|0.30" >> "$config_file"
        echo "✅ Added background music for $agent: $track"
    fi

    if ! is_enabled; then
        set_enabled "true"
        echo "🎵 Background music enabled automatically"
    fi
}

set_all_agents_track() {
    local track="$1"
    local config_file="$SCRIPT_DIR/../config/audio-effects.cfg"

    if [[ -z "$track" ]]; then
        echo "❌ Error: Track name required"
        echo "Usage: $0 set-all TRACK_NAME"
        exit 1
    fi

    local bg_dir="$SCRIPT_DIR/../audio/tracks"
    if [[ ! -f "$bg_dir/$track" ]]; then
        echo "❌ Error: Track not found: $track"
        echo "Run '$0 list' to see available tracks"
        exit 1
    fi

    local temp_file
    temp_file=$(mktemp)
    local count=0

    awk -F'|' -v track="$track" '
        # Skip comments and empty lines
        /^#/ || /^[[:space:]]*$/ { print; next }

        # Skip only _party_mode (but UPDATE default!)
        $1 == "_party_mode" { print; next }

        # Update all agent lines including default
        {
            print $1 "|" $2 "|" track "|" $4
        }
    ' "$config_file" > "$temp_file"

    count=$(grep -v '^#' "$config_file" | grep -v '^[[:space:]]*$' | grep -v '^_party_mode|' | wc -l)

    mv "$temp_file" "$config_file"
    echo "✅ Updated background music for $count agents: $track"

    if ! is_enabled; then
        set_enabled "true"
        echo "🎵 Background music enabled automatically"
    fi
}

case "${1:-status}" in
    status|"")
        show_status
        ;;
    on|enable|true)
        set_enabled "true"
        echo "🎵 Background music ENABLED at $(get_volume) volume"
        ;;
    off|disable|false)
        set_enabled "false"
        echo "🔇 Background music DISABLED"
        ;;
    volume)
        if [[ -z "${2:-}" ]]; then
            echo "Current volume: $(get_volume)"
        else
            set_volume "$2"
            echo "🎵 Background music volume set to $2"
        fi
        ;;
    list)
        list_tracks
        ;;
    set-default)
        if [[ -z "${2:-}" ]]; then
            echo "❌ Error: No track name provided"
            echo "Usage: $0 set-default TRACK_NAME"
            echo "Run '$0 list' to see available tracks"
            exit 1
        fi
        set_default_track "$2"
        ;;
    set-agent)
        if [[ -z "${2:-}" ]] || [[ -z "${3:-}" ]]; then
            echo "❌ Error: Agent name and track required"
            echo "Usage: $0 set-agent AGENT_NAME TRACK_NAME"
            echo "Run '$0 list' to see available tracks"
            exit 1
        fi
        set_agent_track "$2" "$3"
        ;;
    set-all)
        if [[ -z "${2:-}" ]]; then
            echo "❌ Error: Track name required"
            echo "Usage: $0 set-all TRACK_NAME"
            echo "Run '$0 list' to see available tracks"
            exit 1
        fi
        set_all_agents_track "$2"
        ;;
    get-enabled)
        if is_enabled; then
            echo "true"
        else
            echo "false"
        fi
        ;;
    get-volume)
        get_volume
        ;;
    *)
        echo "Usage: $0 {status|on|off|volume [X]|list|set-default TRACK|set-agent AGENT TRACK|set-all TRACK|get-enabled|get-volume}"
        exit 1
        ;;
esac
