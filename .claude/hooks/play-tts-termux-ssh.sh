#!/usr/bin/env bash

export LC_ALL=C

TEXT="$1"
VOICE_OVERRIDE="$2"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

get_ssh_host() {
  local ssh_host=""

  if [[ -n "$TERMUX_SSH_HOST" ]]; then
    ssh_host="$TERMUX_SSH_HOST"
  elif [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -f "$CLAUDE_PROJECT_DIR/.claude/termux-ssh-host.txt" ]]; then
    ssh_host=$(cat "$CLAUDE_PROJECT_DIR/.claude/termux-ssh-host.txt" 2>/dev/null | tr -d '\n\r')
  elif [[ -f "$SCRIPT_DIR/../termux-ssh-host.txt" ]]; then
    ssh_host=$(cat "$SCRIPT_DIR/../termux-ssh-host.txt" 2>/dev/null | tr -d '\n\r')
  elif [[ -f "$HOME/.claude/termux-ssh-host.txt" ]]; then
    ssh_host=$(cat "$HOME/.claude/termux-ssh-host.txt" 2>/dev/null | tr -d '\n\r')
  fi

  echo "$ssh_host"
}

check_termux_connection() {
  local host="$1"

  if ssh -o ConnectTimeout=2 -o BatchMode=yes "$host" "echo ok" >/dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

SSH_HOST=$(get_ssh_host)

if [[ -z "$SSH_HOST" ]]; then
  echo "❌ Termux SSH provider not configured" >&2
  echo "   Set SSH host alias in one of:" >&2
  echo "   - Environment: export TERMUX_SSH_HOST='android'" >&2
  echo "   - Global: echo 'android' > ~/.claude/termux-ssh-host.txt" >&2
  echo "   - Project: echo 'android' > .claude/termux-ssh-host.txt" >&2
  echo "" >&2
  echo "   See provider documentation for SSH setup instructions" >&2
  exit 1
fi

if [[ -z "$TEXT" ]]; then
  echo "❌ No text provided for TTS" >&2
  exit 1
fi

if ! check_termux_connection "$SSH_HOST"; then
  echo "⚠️  Cannot connect to SSH host '$SSH_HOST'" >&2
  echo "   Make sure:" >&2
  echo "   - Android device is reachable" >&2
  echo "   - SSH server is running: sshd" >&2
  echo "   - SSH config is correct in ~/.ssh/config" >&2
  exit 1
fi

SAFE_TEXT="${TEXT//\'/\'\\\'\'}"

ssh -o ConnectTimeout=5 "$SSH_HOST" "termux-tts-speak '$SAFE_TEXT'" &

SSH_PID=$!

sleep 0.2

if kill -0 "$SSH_PID" 2>/dev/null; then
  echo "✓ TTS sent to Android device via SSH" >&2
else
  echo "⚠️  SSH command may have failed" >&2
fi

echo ""
