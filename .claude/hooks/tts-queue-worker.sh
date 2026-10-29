#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${XDG_RUNTIME_DIR:-}" ]] && [[ -d "$XDG_RUNTIME_DIR" ]]; then
  QUEUE_DIR="$XDG_RUNTIME_DIR/agentvibes-tts-queue"
else
  QUEUE_DIR="/tmp/agentvibes-tts-queue-$USER"
fi

if [[ ! -d "$QUEUE_DIR" ]]; then
  echo "Error: Queue directory does not exist: $QUEUE_DIR" >&2
  exit 1
fi

if [[ "$(stat -c '%u' "$QUEUE_DIR" 2>/dev/null || stat -f '%u' "$QUEUE_DIR" 2>/dev/null)" != "$(id -u)" ]]; then
  echo "Error: Queue directory not owned by current user" >&2
  exit 1
fi

WORKER_PID_FILE="$QUEUE_DIR/worker.pid"
IDLE_TIMEOUT=5

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

SPEAKER_DELAY=4

if [[ -f ".claude/tts-speaker-delay.txt" ]]; then
  CUSTOM_DELAY=$(cat .claude/tts-speaker-delay.txt 2>/dev/null | tr -d '[:space:]')
  if [[ "$CUSTOM_DELAY" =~ ^[0-9]+$ ]]; then
    SPEAKER_DELAY=$CUSTOM_DELAY
  fi
elif [[ -f "$HOME/.claude/tts-speaker-delay.txt" ]]; then
  CUSTOM_DELAY=$(cat "$HOME/.claude/tts-speaker-delay.txt" 2>/dev/null | tr -d '[:space:]')
  if [[ "$CUSTOM_DELAY" =~ ^[0-9]+$ ]]; then
    SPEAKER_DELAY=$CUSTOM_DELAY
  fi
fi

trap 'rm -f "$WORKER_PID_FILE"' EXIT

process_queue() {
  local idle_count=0

  while true; do
    local queue_item=$(ls -1 "$QUEUE_DIR"/*.queue 2>/dev/null | sort | head -1)

    if [[ -z "$queue_item" ]]; then
      idle_count=$((idle_count + 1))

      if [[ $idle_count -ge $IDLE_TIMEOUT ]]; then
        exit 0
      fi

      sleep 1
      continue
    fi

    idle_count=0

    source "$queue_item"

    TEXT=$(echo -n "$TEXT_B64" | base64 -d)
    VOICE=$(echo -n "$VOICE_B64" | base64 -d)
    AGENT=$(echo -n "${AGENT_B64:-}" | base64 -d 2>/dev/null || echo "default")

    if [[ -f "$SCRIPT_DIR/play-tts-enhanced.sh" ]] && [[ "$AGENT" != "default" ]] && [[ -n "$AGENT" ]]; then
      bash "$SCRIPT_DIR/play-tts-enhanced.sh" "$TEXT" "$AGENT" "$VOICE" || true
    else
      if [[ -n "${VOICE:-}" ]]; then
        bash "$SCRIPT_DIR/play-tts.sh" "$TEXT" "$VOICE" || true
      else
        bash "$SCRIPT_DIR/play-tts.sh" "$TEXT" || true
      fi
    fi

    sleep $SPEAKER_DELAY

    rm -f "$queue_item"
  done
}

process_queue
