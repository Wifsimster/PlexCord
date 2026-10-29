#!/usr/bin/env bash

PIPER_VOICES_BASE_URL="https://huggingface.co/rhasspy/piper-voices/resolve/main"

get_voice_storage_dir() {
  local voice_dir

  if [[ -n "$PIPER_VOICES_DIR" ]]; then
    voice_dir="$PIPER_VOICES_DIR"
  else
    local config_file
    if [[ -n "$CLAUDE_PROJECT_DIR" ]] && [[ -f "$CLAUDE_PROJECT_DIR/.claude/piper-voices-dir.txt" ]]; then
      config_file="$CLAUDE_PROJECT_DIR/.claude/piper-voices-dir.txt"
    else
      local current_dir="$PWD"
      while [[ "$current_dir" != "/" ]]; do
        if [[ -f "$current_dir/.claude/piper-voices-dir.txt" ]]; then
          config_file="$current_dir/.claude/piper-voices-dir.txt"
          break
        fi
        current_dir=$(dirname "$current_dir")
      done

      if [[ -z "$config_file" ]] && [[ -f "$HOME/.claude/piper-voices-dir.txt" ]]; then
        config_file="$HOME/.claude/piper-voices-dir.txt"
      fi
    fi

    if [[ -n "$config_file" ]]; then
      voice_dir=$(cat "$config_file" | tr -d '[:space:]')
    fi
  fi

  if [[ -z "$voice_dir" ]]; then
    voice_dir="$HOME/.claude/piper-voices"
  fi

  mkdir -p "$voice_dir"
  echo "$voice_dir"
}

verify_voice() {
  local voice_name="$1"
  local voice_dir
  voice_dir=$(get_voice_storage_dir)

  local onnx_file="$voice_dir/${voice_name}.onnx"
  local json_file="$voice_dir/${voice_name}.onnx.json"

  [[ -f "$onnx_file" ]] && [[ -f "$json_file" ]]
}

get_voice_path() {
  local voice_name="$1"
  local voice_dir
  voice_dir=$(get_voice_storage_dir)

  local onnx_file="$voice_dir/${voice_name}.onnx"

  if [[ ! -f "$onnx_file" ]]; then
    echo "❌ Voice model not found: $voice_name" >&2
    return 1
  fi

  echo "$onnx_file"
}

parse_voice_components() {
  local voice_name="$1"

  local lang_locale="${voice_name%%-*}"
  local speaker_quality="${voice_name#*-}"

  LANG="${lang_locale%%_*}"
  LOCALE="${lang_locale#*_}"
  SPEAKER="${speaker_quality%%-*}"
  QUALITY="${speaker_quality#*-}"
}

download_voice() {
  local voice_name="$1"
  local lang_code="${2:-}"

  local voice_dir
  voice_dir=$(get_voice_storage_dir)

  if verify_voice "$voice_name"; then
    echo "✅ Voice already downloaded: $voice_name"
    return 0
  fi

  parse_voice_components "$voice_name"

  local model_path="${LANG}/${LANG}_${LOCALE}/${SPEAKER}/${QUALITY}/${voice_name}"
  local onnx_url="${PIPER_VOICES_BASE_URL}/${model_path}.onnx"
  local json_url="${PIPER_VOICES_BASE_URL}/${model_path}.onnx.json"

  echo "📥 Downloading Piper voice: $voice_name"
  echo "   Source: HuggingFace (rhasspy/piper-voices)"
  echo "   Size: ~25MB"
  echo ""

  echo "   Downloading model file..."
  if ! curl -L --progress-bar --connect-timeout 30 --max-time 300 -o "$voice_dir/${voice_name}.onnx" "$onnx_url"; then
    echo "❌ Failed to download voice model"
    rm -f "$voice_dir/${voice_name}.onnx"
    return 1
  fi

  echo "   Downloading config file..."
  if ! curl -L -s --connect-timeout 30 --max-time 60 -o "$voice_dir/${voice_name}.onnx.json" "$json_url"; then
    echo "❌ Failed to download voice config"
    rm -f "$voice_dir/${voice_name}.onnx" "$voice_dir/${voice_name}.onnx.json"
    return 1
  fi

  if [[ ! -s "$voice_dir/${voice_name}.onnx" ]]; then
    echo "❌ Downloaded file is empty or corrupt"
    rm -f "$voice_dir/${voice_name}.onnx" "$voice_dir/${voice_name}.onnx.json"
    return 1
  fi

  echo "✅ Voice downloaded successfully: $voice_name"
  echo "   Location: $voice_dir/${voice_name}.onnx"
}

list_downloaded_voices() {
  local voice_dir
  voice_dir=$(get_voice_storage_dir)

  echo "📦 Downloaded Piper Voices:"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  local count=0
  shopt -s nullglob
  for onnx_file in "$voice_dir"/*.onnx; do
    if [[ -f "$onnx_file" ]]; then
      local voice_name
      voice_name=$(basename "$onnx_file" .onnx)
      local file_size
      file_size=$(du -h "$onnx_file" | cut -f1)
      echo "  • $voice_name ($file_size)"
      ((count++))
    fi
  done
  shopt -u nullglob

  if [[ $count -eq 0 ]]; then
    echo "  (No voices downloaded yet)"
  fi

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "Total: $count voices"
}
