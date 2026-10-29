#!/usr/bin/env bash

export LC_ALL=C

TARGET_AUDIO_FILE="${CLAUDE_PROJECT_DIR:-.}/.claude/last-target-audio.txt"

if [ ! -f "$TARGET_AUDIO_FILE" ]; then
  echo "❌ No target language audio found."
  echo "   Language learning mode may not be active."
  echo "   Activate with: /agent-vibes:learn"
  exit 1
fi

LAST_AUDIO=$(cat "$TARGET_AUDIO_FILE")

if [ ! -f "$LAST_AUDIO" ]; then
  echo "❌ Audio file not found: $LAST_AUDIO"
  echo "   The file may have been deleted or moved."
  exit 1
fi

echo "🔁 Replaying target language audio..."

LOCK_FILE="/tmp/agentvibes-audio.lock"

for i in {1..60}; do
  if [ ! -f "$LOCK_FILE" ]; then
    break
  fi
  sleep 0.5
done

touch "$LOCK_FILE"

DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$LAST_AUDIO" 2>/dev/null)
DURATION=${DURATION%.*}
DURATION=${DURATION:-1}

(paplay "$LAST_AUDIO" || aplay "$LAST_AUDIO" || mpg123 "$LAST_AUDIO" || mpv "$LAST_AUDIO") >/dev/null 2>&1 &
PLAYER_PID=$!

(sleep $DURATION; rm -f "$LOCK_FILE") &
disown

echo "✅ Replay complete: $(basename "$LAST_AUDIO")"
