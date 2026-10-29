#!/usr/bin/env bash

set -euo pipefail

export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

INPUT_FILE="${1:-}"
AGENT_NAME="${2:-default}"
OUTPUT_FILE="${3:-}"

CONFIG_FILE="$(cd "$SCRIPT_DIR/.." && pwd)/config/audio-effects.cfg"
BACKGROUNDS_DIR="$(cd "$SCRIPT_DIR/../audio" && pwd)/tracks"
ENABLED_FILE="$(cd "$SCRIPT_DIR/.." && pwd)/config/background-music-enabled.txt"

is_background_music_enabled() {
    if [[ ! -f "$ENABLED_FILE" ]]; then
        return 1
    fi

    local enabled
    enabled=$(cat "$ENABLED_FILE" 2>/dev/null | tr -d '[:space:]')

    [[ "$enabled" == "true" ]]
}

if [[ -z "$INPUT_FILE" ]] || [[ ! -f "$INPUT_FILE" ]]; then
    echo "Error: Input file required and must exist" >&2
    echo "Usage: $0 <input.wav> [agent_name] [output.wav]" >&2
    exit 1
fi

if [[ -z "$OUTPUT_FILE" ]]; then
    OUTPUT_FILE="${INPUT_FILE%.wav}-processed.wav"
fi

if ! command -v sox &> /dev/null; then
    echo "Warning: sox not installed, skipping effects" >&2
    cp "$INPUT_FILE" "$OUTPUT_FILE"
    echo "$OUTPUT_FILE"
    exit 0
fi

get_agent_config() {
    local agent="$1"

    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "default|gain -8||0.0"
        return
    fi

    local config
    config=$(grep -i "^${agent}|" "$CONFIG_FILE" 2>/dev/null | head -1)

    if [[ -z "$config" ]]; then
        config=$(grep "^default|" "$CONFIG_FILE" 2>/dev/null | head -1)
    fi

    if [[ -n "$config" ]]; then
        echo "$config"
    else
        echo "default|gain -8||0.0"
    fi
}

apply_sox_effects() {
    local input="$1"
    local output="$2"
    local effects="$3"

    if [[ -z "$effects" ]]; then
        cp "$input" "$output"
        return 0
    fi

    # shellcheck disable=SC2086
    sox "$input" "$output" $effects 2>/dev/null || {
        echo "Warning: Sox effects failed, using original" >&2
        cp "$input" "$output"
    }
}

POSITION_FILE="$SCRIPT_DIR/../config/background-music-position.txt"

get_background_position() {
    local bg_file="$1"
    local bg_name
    bg_name=$(basename "$bg_file")

    if [[ -f "$POSITION_FILE" ]]; then
        grep "^${bg_name}:" "$POSITION_FILE" 2>/dev/null | cut -d: -f2 | tr -d '[:space:]' || echo "0"
    else
        echo "0"
    fi
}

save_background_position() {
    local bg_file="$1"
    local position="$2"
    local bg_name
    bg_name=$(basename "$bg_file")

    mkdir -p "$(dirname "$POSITION_FILE")"

    if [[ -f "$POSITION_FILE" ]]; then
        grep -v "^${bg_name}:" "$POSITION_FILE" > "${POSITION_FILE}.tmp" 2>/dev/null || true
        mv "${POSITION_FILE}.tmp" "$POSITION_FILE"
    fi
    echo "${bg_name}:${position}" >> "$POSITION_FILE"
}

mix_background() {
    local voice="$1"
    local background="$2"
    local volume="$3"
    local output="$4"

    if [[ -z "$background" ]] || [[ ! -f "$background" ]]; then
        cp "$voice" "$output"
        return 0
    fi

    if ! command -v ffmpeg &> /dev/null; then
        echo "Warning: ffmpeg not installed, skipping background mix" >&2
        cp "$voice" "$output"
        return 0
    fi

    local duration
    duration=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$voice" 2>/dev/null)

    if [[ -z "$duration" ]]; then
        cp "$voice" "$output"
        return 0
    fi

    local bg_duration
    bg_duration=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$background" 2>/dev/null)
    bg_duration=${bg_duration:-0}

    local start_pos
    start_pos=$(get_background_position "$background")

    if command -v bc &> /dev/null; then
        if ! [[ "$start_pos" =~ ^[0-9]+\.?[0-9]*$ ]] || (( $(echo "$start_pos < 0.001" | bc -l) )); then
            start_pos="0"
        fi
    else
        if ! [[ "$start_pos" =~ ^[0-9]+\.?[0-9]*$ ]]; then
            start_pos="0"
        fi
    fi

    if command -v bc &> /dev/null && [[ -n "$bg_duration" ]]; then
        if (( $(echo "$start_pos >= $bg_duration" | bc -l) )); then
            start_pos=$(echo "$start_pos % $bg_duration" | bc -l)
        fi
    fi

    local total_duration
    if command -v bc &> /dev/null; then
        total_duration=$(echo "$duration + 2" | bc -l)
    else
        total_duration=$(awk "BEGIN {print $duration + 2}")
    fi

    local new_pos
    if command -v bc &> /dev/null; then
        new_pos=$(echo "$start_pos + $total_duration" | bc -l)
        if [[ -n "$bg_duration" ]] && (( $(echo "$new_pos >= $bg_duration" | bc -l) )); then
            new_pos=$(echo "$new_pos % $bg_duration" | bc -l)
        fi
    else
        new_pos="0"
    fi

    local bg_fade_out_start
    if command -v bc &> /dev/null; then
        bg_fade_out_start=$(echo "$duration" | bc -l)
    else
        bg_fade_out_start="$duration"
    fi

    if [[ -z "${AGENTVIBES_RDP_MODE:-}" ]]; then
        if [[ -n "${SSH_CLIENT:-}" ]] || [[ -n "${SSH_TTY:-}" ]] || [[ "${DISPLAY:-}" =~ ^localhost:.* ]]; then
            export AGENTVIBES_RDP_MODE=true
        fi
    fi

    local audio_settings=""
    if [[ "${AGENTVIBES_RDP_MODE:-false}" == "true" ]]; then
        audio_settings="-ac 1 -ar 22050 -b:a 64k"
    fi

    local voice_delay_ms="2000"
    local voice_delay_sec="2.0"
    local bg_fade_out_adjusted
    if command -v bc &> /dev/null; then
        bg_fade_out_adjusted=$(echo "$duration + $voice_delay_sec" | bc -l)
    else
        bg_fade_out_adjusted=$(echo "$duration + 2" | bc)
    fi

    ffmpeg -y -i "$voice" -ss "$start_pos" -stream_loop -1 -i "$background" \
        -filter_complex "[1:a]volume=${volume},afade=t=in:st=0:d=0.3,afade=t=out:st=${bg_fade_out_adjusted}:d=2[bg];[0:a]adelay=${voice_delay_ms}|${voice_delay_ms}[v];[v][bg]amix=inputs=2:duration=longest[out]" \
        -map "[out]" $audio_settings -t "$total_duration" "$output" 2>/dev/null || {
        echo "Warning: Background mixing failed, using voice only" >&2
        cp "$voice" "$output"
        return
    }

    save_background_position "$background" "$new_pos"
}

main() {
    echo "🎛️ Processing audio for agent: $AGENT_NAME" >&2

    local config
    config=$(get_agent_config "$AGENT_NAME")

    IFS='|' read -r _ sox_effects background_file bg_volume <<< "$config"

    local temp_effects
    local temp_final
    temp_effects="/tmp/agentvibes-effects-$$.wav"
    temp_final="/tmp/agentvibes-final-$$.wav"

    trap 'rm -f /tmp/agentvibes-effects-'"$$"'.wav /tmp/agentvibes-final-'"$$"'.wav' EXIT

    if [[ -n "$sox_effects" ]]; then
        echo "  → Applying effects: $sox_effects" >&2
        apply_sox_effects "$INPUT_FILE" "$temp_effects" "$sox_effects"
    else
        cp "$INPUT_FILE" "$temp_effects"
    fi

    local background_path=""
    if [[ -n "$background_file" ]]; then
        background_path="$BACKGROUNDS_DIR/$background_file"
    fi

    local used_background=""
    if is_background_music_enabled && [[ -n "$background_path" ]] && [[ -f "$background_path" ]] && [[ "${bg_volume:-0}" != "0" ]] && [[ "${bg_volume:-0}" != "0.0" ]]; then
        echo "  → Mixing background: $background_file at ${bg_volume} volume" >&2
        mix_background "$temp_effects" "$background_path" "$bg_volume" "$temp_final"
        used_background="$background_path"
    else
        cp "$temp_effects" "$temp_final"
    fi

    mv "$temp_final" "$OUTPUT_FILE"

    echo "$OUTPUT_FILE|$used_background"
}

main
