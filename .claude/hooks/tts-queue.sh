#!/usr/bin/env bash

set -euo pipefail

if [[ -n "${XDG_RUNTIME_DIR:-}" ]] && [[ -d "$XDG_RUNTIME_DIR" ]]; then
  QUEUE_DIR="$XDG_RUNTIME_DIR/agentvibes-tts-queue"
else
  QUEUE_DIR="/tmp/agentvibes-tts-queue-$USER"
fi

QUEUE_LOCK="$QUEUE_DIR/queue.lock"
WORKER_PID_FILE="$QUEUE_DIR/worker.pid"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ ! -d "$QUEUE_DIR" ]]; then
  mkdir -p "$QUEUE_DIR"
  chmod 700 "$QUEUE_DIR"
fi

add_to_queue() {
  local text="$1"
  local voice="${2:-}"
  local agent="${3:-default}"

  local timestamp=$(date +%s%N)
  local queue_file="$QUEUE_DIR/$timestamp.queue"

  cat > "$queue_file" <<EOF
TEXT_B64=$(echo -n "$text" | base64 -w0)
VOICE_B64=$(echo -n "$voice" | base64 -w0)
AGENT_B64=$(echo -n "$agent" | base64 -w0)
EOF

  start_worker_if_needed
}

start_worker_if_needed() {
  exec 200>"$QUEUE_LOCK"

  if ! flock -x -w 5 200; then
    echo "Warning: Could not acquire queue lock" >&2
    return 1
  fi

  if [[ -f "$WORKER_PID_FILE" ]]; then
    local pid=$(cat "$WORKER_PID_FILE")
    if kill -0 "$pid" 2>/dev/null; then
      flock -u 200
      exec 200>&-
      return 0
    fi
  fi

  "$SCRIPT_DIR/tts-queue-worker.sh" &
  local worker_pid=$!
  echo $worker_pid > "$WORKER_PID_FILE"

  flock -u 200
  exec 200>&-
}

clear_queue() {
  rm -f "$QUEUE_DIR"/*.queue
  echo "✅ Queue cleared"
}

show_queue() {
  local count=$(ls -1 "$QUEUE_DIR"/*.queue 2>/dev/null | wc -l)
  echo "📊 Queue status: $count items pending"

  if [[ -f "$WORKER_PID_FILE" ]]; then
    local pid=$(cat "$WORKER_PID_FILE")
    if kill -0 "$pid" 2>/dev/null; then
      echo "✅ Worker process running (PID: $pid)"
    else
      echo "❌ Worker process not running"
    fi
  else
    echo "❌ Worker process not running"
  fi
}

case "${1:-help}" in
  add)
    add_to_queue "${2:-}" "${3:-}" "${4:-default}"
    ;;
  clear)
    clear_queue
    ;;
  status)
    show_queue
    ;;
  *)
    echo "Usage: tts-queue.sh {add|clear|status}"
    echo ""
    echo "Commands:"
    echo "  add <text> [voice] [agent]  Add TTS request to queue with optional agent for background music"
    echo "  clear                       Clear all pending requests"
    echo "  status                      Show queue status"
    exit 1
    ;;
esac
