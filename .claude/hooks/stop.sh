#!/usr/bin/env bash

export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

MODE_FILE="$PROJECT_ROOT/.agentvibes/config/mode.txt"

AGENTVIBES_HOOKS="$PROJECT_ROOT/.agentvibes/hooks"
if [[ ! -d "$AGENTVIBES_HOOKS" ]]; then
  exit 0
fi

CURRENT_MODE="full"
if [[ -f "$MODE_FILE" ]]; then
  CURRENT_MODE=$(cat "$MODE_FILE" 2>/dev/null | tr -d '[:space:]')
fi

if [[ "$CURRENT_MODE" == "lite" ]]; then
  if [[ -f "$AGENTVIBES_HOOKS/stop-lite.sh" ]]; then
    bash "$AGENTVIBES_HOOKS/stop-lite.sh" "$@"
  fi
fi
