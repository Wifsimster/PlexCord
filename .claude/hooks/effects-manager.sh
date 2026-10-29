#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$SCRIPT_DIR/../config/audio-effects.cfg"

declare -A REVERB_LEVELS=(
    ["off"]=""
    ["light"]="reverb 20 50 50"
    ["medium"]="reverb 40 50 70"
    ["heavy"]="reverb 70 50 100"
    ["cathedral"]="reverb 90 30 100"
)

get_agent_effects() {
    local agent_name="$1"

    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo ""
        return
    fi

    local config_line=$(grep "^${agent_name}|" "$CONFIG_FILE" 2>/dev/null || echo "")

    if [[ -z "$config_line" ]]; then
        config_line=$(grep "^default|" "$CONFIG_FILE" 2>/dev/null || echo "")
    fi

    if [[ -n "$config_line" ]]; then
        echo "$config_line" | cut -d'|' -f2
    else
        echo ""
    fi
}

set_reverb() {
    local level="$1"
    local agent_name="${2:-default}"
    local apply_all="${3:-}"

    if [[ ! -v REVERB_LEVELS[$level] ]]; then
        echo "❌ Invalid reverb level: $level"
        echo "Valid levels: off, light, medium, heavy, cathedral"
        return 1
    fi

    local reverb_effect="${REVERB_LEVELS[$level]}"

    if [[ "$apply_all" == "--all" ]]; then
        echo "🎛️ Setting reverb to '$level' for all agents..."

        local temp_file="${CONFIG_FILE}.tmp"

        while IFS='|' read -r agent sox_effects bg_file bg_vol || [[ -n "$agent" ]]; do
            if [[ "$agent" =~ ^# ]] || [[ -z "$agent" ]]; then
                echo "$agent|$sox_effects|$bg_file|$bg_vol" >> "$temp_file"
                continue
            fi

            local new_effects=$(echo "$sox_effects" | sed -E 's/reverb [0-9]+ [0-9]+ [0-9]+//g' | sed 's/  */ /g' | sed 's/^ //;s/ $//')

            if [[ -n "$reverb_effect" ]]; then
                if [[ -n "$new_effects" ]]; then
                    new_effects="$reverb_effect $new_effects"
                else
                    new_effects="$reverb_effect"
                fi
            fi

            echo "${agent}|${new_effects}|${bg_file}|${bg_vol}" >> "$temp_file"
        done < "$CONFIG_FILE"

        mv "$temp_file" "$CONFIG_FILE"

        echo "✅ Reverb set to '$level' for all agents"
    else
        echo "🎛️ Setting reverb to '$level' for '$agent_name'..."

        if ! grep -q "^${agent_name}|" "$CONFIG_FILE" 2>/dev/null; then
            echo "${agent_name}|${reverb_effect}|agentvibes_soft_flamenco_loop.mp3|0.30" >> "$CONFIG_FILE"
            echo "✅ Created new config for '$agent_name' with reverb '$level'"
            return
        fi

        local temp_file="${CONFIG_FILE}.tmp"

        while IFS='|' read -r agent sox_effects bg_file bg_vol || [[ -n "$agent" ]]; do
            if [[ "$agent" =~ ^# ]] || [[ -z "$agent" ]]; then
                echo "$agent|$sox_effects|$bg_file|$bg_vol" >> "$temp_file"
                continue
            fi

            if [[ "$agent" == "$agent_name" ]]; then
                local new_effects=$(echo "$sox_effects" | sed -E 's/reverb [0-9]+ [0-9]+ [0-9]+//g' | sed 's/  */ /g' | sed 's/^ //;s/ $//')

                if [[ -n "$reverb_effect" ]]; then
                    if [[ -n "$new_effects" ]]; then
                        new_effects="$reverb_effect $new_effects"
                    else
                        new_effects="$reverb_effect"
                    fi
                fi

                echo "${agent}|${new_effects}|${bg_file}|${bg_vol}" >> "$temp_file"
            else
                echo "${agent}|${sox_effects}|${bg_file}|${bg_vol}" >> "$temp_file"
            fi
        done < "$CONFIG_FILE"

        mv "$temp_file" "$CONFIG_FILE"

        echo "✅ Reverb set to '$level' for '$agent_name'"
    fi
}

list_effects() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "❌ Config file not found: $CONFIG_FILE"
        return 1
    fi

    echo "📊 Current Audio Effects Configuration:"
    echo ""

    while IFS='|' read -r agent sox_effects bg_file bg_vol || [[ -n "$agent" ]]; do
        if [[ "$agent" =~ ^# ]] || [[ -z "$agent" ]]; then
            continue
        fi

        local reverb_level="off"
        if [[ "$sox_effects" =~ reverb\ ([0-9]+)\ ([0-9]+)\ ([0-9]+) ]]; then
            local reverb_val="${BASH_REMATCH[1]}"
            if [[ "$reverb_val" -le 20 ]]; then
                reverb_level="light"
            elif [[ "$reverb_val" -le 40 ]]; then
                reverb_level="medium"
            elif [[ "$reverb_val" -le 70 ]]; then
                reverb_level="heavy"
            else
                reverb_level="cathedral"
            fi
        fi

        echo "   $agent: reverb=$reverb_level"
        if [[ -n "$sox_effects" ]]; then
            echo "      Effects: $sox_effects"
        fi
    done < "$CONFIG_FILE"
}

get_reverb_level() {
    local agent_name="$1"
    local effects=$(get_agent_effects "$agent_name")

    if [[ -z "$effects" ]] || [[ ! "$effects" =~ reverb ]]; then
        echo "off"
        return
    fi

    if [[ "$effects" =~ reverb\ ([0-9]+)\ ([0-9]+)\ ([0-9]+) ]]; then
        local reverb_val="${BASH_REMATCH[1]}"
        if [[ "$reverb_val" -le 20 ]]; then
            echo "light"
        elif [[ "$reverb_val" -le 40 ]]; then
            echo "medium"
        elif [[ "$reverb_val" -le 70 ]]; then
            echo "heavy"
        else
            echo "cathedral"
        fi
    else
        echo "off"
    fi
}

case "${1:-help}" in
    set-reverb)
        if [[ -z "${2:-}" ]]; then
            echo "Usage: effects-manager.sh set-reverb <level> [agent-name] [--all]"
            echo "Levels: off, light, medium, heavy, cathedral"
            exit 1
        fi
        set_reverb "$2" "${3:-default}" "${4:-}"
        ;;
    get-reverb)
        get_reverb_level "${2:-default}"
        ;;
    get-effects)
        get_agent_effects "${2:-default}"
        ;;
    list)
        list_effects
        ;;
    *)
        echo "Usage: effects-manager.sh {set-reverb|get-reverb|get-effects|list}"
        echo ""
        echo "Commands:"
        echo "  set-reverb <level> [agent] [--all]  Set reverb level"
        echo "  get-reverb [agent]                  Get current reverb level"
        echo "  get-effects [agent]                 Get all effects for agent"
        echo "  list                                List all agent effects"
        echo ""
        echo "Reverb Levels: off, light, medium, heavy, cathedral"
        exit 1
        ;;
esac
