#!/usr/bin/env bash

export LC_ALL=C

TEXT="$1"
VOICE_OVERRIDE="$2"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/piper-voice-manager.sh"
source "$SCRIPT_DIR/language-manager.sh"

DEFAULT_VOICE="en_US-lessac-medium"

VOICE_MODEL=""

CURRENT_LANGUAGE=$(get_language_code)

if [[ -n "$VOICE_OVERRIDE" ]]; then
  VOICE_MODEL="$VOICE_OVERRIDE"
  echo "🎤 Using voice: $VOICE_OVERRIDE (session-specific)"
else
  VOICE_FILE=""

  if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -f "$CLAUDE_PROJECT_DIR/.claude/tts-voice.txt" ]]; then
    VOICE_FILE="$CLAUDE_PROJECT_DIR/.claude/tts-voice.txt"
  elif [[ -f "$SCRIPT_DIR/../tts-voice.txt" ]]; then
    VOICE_FILE="$SCRIPT_DIR/../tts-voice.txt"
  elif [[ -f "$HOME/.claude/tts-voice.txt" ]]; then
    VOICE_FILE="$HOME/.claude/tts-voice.txt"
  fi

  if [[ -n "$VOICE_FILE" ]]; then
    FILE_VOICE=$(cat "$VOICE_FILE" 2>/dev/null)

    VOICE_DIR=$(dirname "$VOICE_FILE")
    MODEL_FILE="$VOICE_DIR/tts-piper-model.txt"
    SPEAKER_ID_FILE="$VOICE_DIR/tts-piper-speaker-id.txt"

    if [[ -f "$MODEL_FILE" ]] && [[ -f "$SPEAKER_ID_FILE" ]]; then
      VOICE_MODEL=$(cat "$MODEL_FILE" 2>/dev/null)
      SPEAKER_ID=$(cat "$SPEAKER_ID_FILE" 2>/dev/null)
      echo "🎭 Using multi-speaker voice: $FILE_VOICE (Model: $VOICE_MODEL, Speaker ID: $SPEAKER_ID)"
    # Check if it's a standard Piper model name or custom voice (just use as-is)
    elif [[ -n "$FILE_VOICE" ]]; then
      VOICE_MODEL="$FILE_VOICE"
    fi
  fi

  if [[ -z "$VOICE_MODEL" ]]; then
    LANG_VOICE=$(get_voice_for_language "$CURRENT_LANGUAGE" "piper" 2>/dev/null)

    if [[ -n "$LANG_VOICE" ]]; then
      VOICE_MODEL="$LANG_VOICE"
      echo "🌍 Using $CURRENT_LANGUAGE voice: $LANG_VOICE (Piper)"
    else
      VOICE_MODEL="$DEFAULT_VOICE"
    fi
  fi
fi

if [[ -z "$TEXT" ]]; then
  echo "Usage: $0 \"text to speak\" [voice_model_name]"
  exit 1
fi

if ! command -v piper &> /dev/null; then
  echo "❌ Error: Piper TTS not installed"
  echo "Install with: pipx install piper-tts"
  echo "Or run: .claude/hooks/piper-installer.sh"
  exit 2
fi

if [[ "${AGENTVIBES_TEST_MODE:-false}" != "true" ]] && ! verify_voice "$VOICE_MODEL"; then
  echo "📥 Voice model not found: $VOICE_MODEL"
  echo "   File size: ~25MB"
  echo "   Preview: https://huggingface.co/rhasspy/piper-voices"
  echo ""
  read -p "   Download this voice model? [y/N]: " -n 1 -r
  echo

  if [[ $REPLY =~ ^[Yy]$ ]]; then
    if ! download_voice "$VOICE_MODEL"; then
      echo "❌ Failed to download voice model"
      echo "Fix: Download manually or choose different voice"
      exit 3
    fi
  else
    echo "❌ Voice download cancelled"
    exit 3
  fi
fi

if [[ "${AGENTVIBES_TEST_MODE:-false}" == "true" ]]; then
  VOICE_PATH="/tmp/mock-voice-${VOICE_MODEL}.onnx"
else
  VOICE_PATH=$(get_voice_path "$VOICE_MODEL")
  if [[ $? -ne 0 ]]; then
    echo "❌ Voice model path not found: $VOICE_MODEL"
    exit 3
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
TEMP_FILE="$AUDIO_DIR/tts-$(date +%s).wav"

get_speech_rate() {
  local target_config=""
  local main_config=""

  if [[ -f "$SCRIPT_DIR/../config/tts-target-speech-rate.txt" ]]; then
    target_config="$SCRIPT_DIR/../config/tts-target-speech-rate.txt"
  elif [[ -f "$HOME/.claude/config/tts-target-speech-rate.txt" ]]; then
    target_config="$HOME/.claude/config/tts-target-speech-rate.txt"
  elif [[ -f "$SCRIPT_DIR/../config/piper-target-speech-rate.txt" ]]; then
    target_config="$SCRIPT_DIR/../config/piper-target-speech-rate.txt"
  elif [[ -f "$HOME/.claude/config/piper-target-speech-rate.txt" ]]; then
    target_config="$HOME/.claude/config/piper-target-speech-rate.txt"
  fi

  if [[ -f "$SCRIPT_DIR/../config/tts-speech-rate.txt" ]]; then
    main_config="$SCRIPT_DIR/../config/tts-speech-rate.txt"
  elif [[ -f "$HOME/.claude/config/tts-speech-rate.txt" ]]; then
    main_config="$HOME/.claude/config/tts-speech-rate.txt"
  elif [[ -f "$SCRIPT_DIR/../config/piper-speech-rate.txt" ]]; then
    main_config="$SCRIPT_DIR/../config/piper-speech-rate.txt"
  elif [[ -f "$HOME/.claude/config/piper-speech-rate.txt" ]]; then
    main_config="$HOME/.claude/config/piper-speech-rate.txt"
  fi

  if [[ "$CURRENT_LANGUAGE" != "english" ]] && [[ -n "$target_config" ]]; then
    local user_speed=$(cat "$target_config" 2>/dev/null)
    echo "scale=2; 1.0 / $user_speed" | bc -l 2>/dev/null || echo "1.0"
    return
  fi

  if [[ -n "$main_config" ]]; then
    local user_speed=$(grep -v '^#' "$main_config" 2>/dev/null | grep -v '^$' | tail -1)
    echo "scale=2; 1.0 / $user_speed" | bc -l 2>/dev/null || echo "1.0"
    return
  fi

  if [[ "$CURRENT_LANGUAGE" != "english" ]]; then
    echo "2.0"
  else
    echo "1.0"
  fi
}

SPEECH_RATE=$(get_speech_rate)

if [[ -n "$SPEAKER_ID" ]]; then
  echo "$TEXT" | piper --model "$VOICE_PATH" --speaker "$SPEAKER_ID" --length-scale "$SPEECH_RATE" --sentence-silence 2.0 --output_file "$TEMP_FILE" 2>/dev/null
else
  echo "$TEXT" | piper --model "$VOICE_PATH" --length-scale "$SPEECH_RATE" --sentence-silence 2.0 --output_file "$TEMP_FILE" 2>/dev/null
fi

if [[ ! -f "$TEMP_FILE" ]] || [[ ! -s "$TEMP_FILE" ]]; then
  echo "❌ Failed to synthesize speech with Piper"
  echo "Voice model: $VOICE_MODEL"
  echo "Check that voice model is valid"
  exit 4
fi

if [[ -z "${AGENTVIBES_RDP_MODE:-}" ]]; then
  if [[ -n "${SSH_CLIENT:-}" ]] || [[ -n "${SSH_TTY:-}" ]] || [[ "${DISPLAY:-}" =~ ^localhost:.* ]]; then
    export AGENTVIBES_RDP_MODE=true
    echo "🌐 Remote session detected - enabling audio compression"
  fi
fi

if [[ "${AGENTVIBES_RDP_MODE:-false}" == "true" ]] && command -v ffmpeg &> /dev/null; then
  COMPRESSED_FILE="$AUDIO_DIR/tts-compressed-$(date +%s).wav"
  ffmpeg -i "$TEMP_FILE" -ac 1 -ar 22050 -b:a 64k -y "$COMPRESSED_FILE" 2>/dev/null

  if [[ -f "$COMPRESSED_FILE" ]]; then
    rm -f "$TEMP_FILE"
    TEMP_FILE="$COMPRESSED_FILE"
  fi
fi

if command -v ffmpeg &> /dev/null; then
  PADDED_FILE="$AUDIO_DIR/tts-padded-$(date +%s).wav"
  ffmpeg -f lavfi -i anullsrc=r=44100:cl=stereo:d=0.2 -i "$TEMP_FILE" \
    -filter_complex "[0:a][1:a]concat=n=2:v=0:a=1[out]" \
    -map "[out]" -y "$PADDED_FILE" 2>/dev/null

  if [[ -f "$PADDED_FILE" ]]; then
    rm -f "$TEMP_FILE"
    TEMP_FILE="$PADDED_FILE"
  fi
fi

BACKGROUND_MUSIC=""
if [[ -f "$SCRIPT_DIR/audio-processor.sh" ]]; then
  PROCESSED_FILE="$AUDIO_DIR/tts-processed-$(date +%s).wav"
  PROCESSOR_OUTPUT=$("$SCRIPT_DIR/audio-processor.sh" "$TEMP_FILE" "default" "$PROCESSED_FILE" 2>/dev/null) || {
    echo "Warning: Audio processing failed, using unprocessed audio" >&2
    PROCESSED_FILE="$TEMP_FILE"
    PROCESSOR_OUTPUT="$TEMP_FILE|"
  }

  PROCESSED_FILE="${PROCESSOR_OUTPUT%%|*}"
  BACKGROUND_MUSIC="${PROCESSOR_OUTPUT##*|}"

  if [[ -f "$PROCESSED_FILE" ]] && [[ "$PROCESSED_FILE" != "$TEMP_FILE" ]]; then
    rm -f "$TEMP_FILE"
    TEMP_FILE="$PROCESSED_FILE"
  fi
fi

LOCK_FILE="/tmp/agentvibes-audio.lock"

for i in {1..4}; do
  if [ ! -f "$LOCK_FILE" ]; then
    break
  fi
  sleep 0.5
done

if [ -f "$LOCK_FILE" ]; then
  echo "⏭️  Skipping TTS (previous audio still playing)" >&2
  exit 0
fi

if [[ "$CURRENT_LANGUAGE" != "english" ]]; then
  TARGET_AUDIO_FILE="${CLAUDE_PROJECT_DIR:-.}/.claude/last-target-audio.txt"
  echo "$TEMP_FILE" > "$TARGET_AUDIO_FILE"
fi

touch "$LOCK_FILE"

DURATION=$(ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 "$TEMP_FILE" 2>/dev/null)
DURATION=${DURATION%.*}
DURATION=${DURATION:-1}

if [[ "${AGENTVIBES_TEST_MODE:-false}" != "true" ]] && [[ "${AGENTVIBES_NO_PLAYBACK:-false}" != "true" ]]; then
  if [[ "$(uname -s)" == "Darwin" ]]; then
    afplay "$TEMP_FILE" >/dev/null 2>&1 &
    PLAYER_PID=$!
  else
    (paplay "$TEMP_FILE" || mpv "$TEMP_FILE" || aplay "$TEMP_FILE") >/dev/null 2>&1 &
    PLAYER_PID=$!
  fi
fi

(sleep $DURATION; rm -f "$LOCK_FILE") &
disown

echo "🎵 Saved to: $TEMP_FILE"
if [[ -n "$BACKGROUND_MUSIC" ]]; then
  echo "🎶 Background music: $BACKGROUND_MUSIC"
fi
echo "🎤 Voice used: $VOICE_MODEL (Piper TTS)"

GLOBAL_MUTE_FILE="$HOME/.agentvibes-muted"
PROJECT_MUTE_FILE="$PROJECT_ROOT/.claude/agentvibes-muted"
PROJECT_UNMUTE_FILE="$PROJECT_ROOT/.claude/agentvibes-unmuted"
BACKGROUND_ENABLED_FILE="$PROJECT_ROOT/.claude/config/background-music-enabled.txt"

if [[ -f "$PROJECT_UNMUTE_FILE" ]] && [[ -f "$GLOBAL_MUTE_FILE" ]]; then
  echo "🔊 Status: Unmuted (project overrides global mute)"
elif [[ -f "$PROJECT_MUTE_FILE" ]]; then
  echo "🔇 Status: Muted (project)"
elif [[ -f "$GLOBAL_MUTE_FILE" ]]; then
  echo "🔇 Status: Would be muted (global) - but this project is speaking"
fi

if [[ -z "$BACKGROUND_MUSIC" ]]; then
  if [[ -f "$BACKGROUND_ENABLED_FILE" ]] && grep -q "true" "$BACKGROUND_ENABLED_FILE" 2>/dev/null; then
    echo "🎵 Background music: Enabled but not playing (check config)"
  else
    echo "🎵 Background music: Disabled"
  fi
fi
