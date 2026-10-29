#!/usr/bin/env bash

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "❌ Error: macOS provider only works on macOS"
  echo "   Current platform: $(uname -s)"
  echo ""
  echo "   Switch to a different provider:"
  echo "   /agent-vibes:provider switch piper"
  exit 1
fi

TEXT="$1"
VOICE_OVERRIDE="$2"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

DEFAULT_VOICE="Samantha"

show_common_voices() {
  echo "  Alex - American English male (enhanced)"
  echo "  Daniel - British English male (enhanced)"
  echo "  Fiona - Scottish English female (enhanced)"
  echo "  Karen - Australian English female (enhanced)"
  echo "  Moira - Irish English female (enhanced)"
  echo "  Samantha - American English female (enhanced)"
  echo "  Tessa - South African English female (enhanced)"
  echo "  Veena - Indian English female (enhanced)"
  echo "  Victoria - American English female"
}

get_voice_file_path() {
  local voice_file=""

  if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -f "$CLAUDE_PROJECT_DIR/.claude/tts-voice.txt" ]]; then
    voice_file="$CLAUDE_PROJECT_DIR/.claude/tts-voice.txt"
  elif [[ -f "$SCRIPT_DIR/../tts-voice.txt" ]]; then
    voice_file="$SCRIPT_DIR/../tts-voice.txt"
  elif [[ -f "$HOME/.claude/tts-voice.txt" ]]; then
    voice_file="$HOME/.claude/tts-voice.txt"
  fi

  echo "$voice_file"
}

VOICE_NAME=""

if [[ -n "$VOICE_OVERRIDE" ]]; then
  VOICE_NAME="$VOICE_OVERRIDE"
  echo "🎤 Using voice: $VOICE_OVERRIDE (session-specific)"
else
  VOICE_FILE=$(get_voice_file_path)

  if [[ -n "$VOICE_FILE" ]] && [[ -f "$VOICE_FILE" ]]; then
    FILE_VOICE=$(cat "$VOICE_FILE" 2>/dev/null | tr -d '\n\r')
    if [[ -n "$FILE_VOICE" ]]; then
      VOICE_NAME="$FILE_VOICE"
    fi
  fi

  if [[ -z "$VOICE_NAME" ]]; then
    VOICE_NAME="$DEFAULT_VOICE"
  fi
fi

if [[ -z "$TEXT" ]]; then
  echo "Usage: $0 \"text to speak\" [voice_name]"
  echo ""
  echo "Common voices (run 'say -v ?' for full list):"
  show_common_voices
  exit 1
fi

validate_voice() {
  local voice="$1"
  say -v ? 2>/dev/null | grep -qi "^${voice} "
}

if ! validate_voice "$VOICE_NAME"; then
  echo "⚠️  Voice '$VOICE_NAME' not found on this system"
  echo "   Falling back to default: $DEFAULT_VOICE"
  VOICE_NAME="$DEFAULT_VOICE"

  if ! validate_voice "$VOICE_NAME"; then
    VOICE_NAME=$(say -v ? 2>/dev/null | grep -i "en_" | head -1 | awk '{print $1}')
    if [[ -z "$VOICE_NAME" ]]; then
      echo "❌ No English voices found on this system"
      exit 2
    fi
    echo "   Using first available English voice: $VOICE_NAME"
  fi
fi

if [[ -n "$CLAUDE_PROJECT_DIR" ]]; then
  AUDIO_DIR="$CLAUDE_PROJECT_DIR/.claude/audio"
else
  CURRENT_DIR="$PWD"
  while [[ "$CURRENT_DIR" != "/" ]]; do
    if [[ -d "$CURRENT_DIR/.claude" ]]; then
      AUDIO_DIR="$CURRENT_DIR/.claude/audio"
      break
    fi
    CURRENT_DIR=$(dirname "$CURRENT_DIR")
  done
  if [[ -z "$AUDIO_DIR" ]]; then
    AUDIO_DIR="$HOME/.claude/audio"
  fi
fi

mkdir -p "$AUDIO_DIR"

TIMESTAMP=$(date +%s)
TEMP_FILE="$AUDIO_DIR/tts-${TIMESTAMP}.aiff"
FINAL_FILE="$AUDIO_DIR/tts-padded-${TIMESTAMP}.wav"

get_speech_rate() {
  local rate_config=""

  if [[ -f "$SCRIPT_DIR/../config/tts-speech-rate.txt" ]]; then
    rate_config="$SCRIPT_DIR/../config/tts-speech-rate.txt"
  elif [[ -f "$HOME/.claude/config/tts-speech-rate.txt" ]]; then
    rate_config="$HOME/.claude/config/tts-speech-rate.txt"
  fi

  if [[ -n "$rate_config" ]]; then
    local user_speed=$(cat "$rate_config" 2>/dev/null | grep -v '^#' | grep -v '^$' | tail -1)
    echo "scale=0; 200 * $user_speed / 1" | bc -l 2>/dev/null || echo "200"
    return
  fi

  echo "200"
}

SPEECH_RATE=$(get_speech_rate)

echo "$TEXT" | say -v "$VOICE_NAME" -r "$SPEECH_RATE" -o "$TEMP_FILE" 2>/dev/null

if [[ ! -f "$TEMP_FILE" ]] || [[ ! -s "$TEMP_FILE" ]]; then
  echo "❌ Failed to synthesize speech with macOS say command"
  echo "Voice: $VOICE_NAME"
  exit 3
fi

if command -v ffmpeg &> /dev/null; then
  ffmpeg -f lavfi -i anullsrc=r=44100:cl=stereo:d=0.2 -i "$TEMP_FILE" \
    -filter_complex "[0:a][1:a]concat=n=2:v=0:a=1[out]" \
    -map "[out]" -y "$FINAL_FILE" 2>/dev/null

  if [[ -f "$FINAL_FILE" ]]; then
    rm -f "$TEMP_FILE"
    TEMP_FILE="$FINAL_FILE"
  fi
else
  FINAL_FILE="$AUDIO_DIR/tts-padded-${TIMESTAMP}.aiff"
  mv "$TEMP_FILE" "$FINAL_FILE"
  TEMP_FILE="$FINAL_FILE"
fi

LOCK_FILE="/tmp/agentvibes-audio.lock"

for i in {1..60}; do
  if [ ! -f "$LOCK_FILE" ]; then
    break
  fi
  sleep 0.5
done

touch "$LOCK_FILE"

if command -v ffprobe &> /dev/null; then
  DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$TEMP_FILE" 2>/dev/null)
  DURATION=${DURATION%.*}
else
  WORD_COUNT=$(echo "$TEXT" | wc -w)
  DURATION=$(( (WORD_COUNT * 60 / 150) + 1 ))
fi
DURATION=${DURATION:-2}

if [[ "${AGENTVIBES_TEST_MODE:-false}" != "true" ]] && [[ "${AGENTVIBES_NO_PLAYBACK:-false}" != "true" ]]; then
  if [[ -n "$SSH_CONNECTION" ]] && [[ -n "$PULSE_SERVER" ]]; then
    if command -v /opt/homebrew/bin/paplay &> /dev/null; then
      /opt/homebrew/bin/paplay "$TEMP_FILE" >/dev/null 2>&1 &
      PLAYER_PID=$!
      echo "🔊 Playing via PulseAudio tunnel"
    else
      echo "⚠️  paplay not found - install pulseaudio for SSH audio"
      afplay "$TEMP_FILE" >/dev/null 2>&1 &
      PLAYER_PID=$!
    fi
  else
    afplay "$TEMP_FILE" >/dev/null 2>&1 &
    PLAYER_PID=$!
  fi
fi

(sleep $DURATION; rm -f "$LOCK_FILE") &
disown

echo "🎵 Saved to: $TEMP_FILE"
echo "🎤 Voice used: $VOICE_NAME (macOS Say)"
