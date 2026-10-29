#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/piper-voice-manager.sh"

AUTO_YES=false
if [[ "$1" == "--yes" ]] || [[ "$1" == "-y" ]]; then
  AUTO_YES=true
fi

COMMON_VOICES=(
  "en_US-ryan-high"
  "en_US-amy-medium"
  "en_US-hfc_female-medium"
  "en_US-lessac-medium"
  "en_US-danny-low"
  "en_US-bryce-medium"
  "en_US-kathleen-low"
  "en_US-kusal-medium"
  "en_US-kristin-medium"
  "en_US-libritts_r-high"
  "en_US-libritts-high"
  "16Speakers"
)

echo "🎙️  Piper Voice Model Downloader"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "This will download the most commonly used Piper voice models."
echo "Each voice is approximately 25MB."
echo ""

if ! command -v piper &> /dev/null; then
  echo "❌ Error: Piper TTS not installed"
  echo "Install with: pipx install piper-tts"
  exit 1
fi

VOICE_DIR=$(get_voice_storage_dir)

echo "📂 Storage location: $VOICE_DIR"
echo ""

ALREADY_DOWNLOADED=0
ALREADY_DOWNLOADED_LIST=()
NEED_DOWNLOAD=()

for voice in "${COMMON_VOICES[@]}"; do
  if verify_voice "$voice" 2>/dev/null; then
    ((ALREADY_DOWNLOADED++))
    ALREADY_DOWNLOADED_LIST+=("$voice")
  else
    NEED_DOWNLOAD+=("$voice")
  fi
done

echo "📊 Status:"
echo "   Already downloaded: $ALREADY_DOWNLOADED voice(s)"
echo "   Need to download: ${#NEED_DOWNLOAD[@]} voice(s)"
echo ""

if [[ $ALREADY_DOWNLOADED -gt 0 ]]; then
  echo "✅ Already downloaded (skipped):"
  for voice in "${ALREADY_DOWNLOADED_LIST[@]}"; do
    echo "   ✓ $voice"
  done
  echo ""
fi

if [[ ${#NEED_DOWNLOAD[@]} -eq 0 ]]; then
  echo "🎉 All common voices ready to use!"
  exit 0
fi

echo "Voices to download:"
for voice in "${NEED_DOWNLOAD[@]}"; do
  echo "  • $voice (~25MB)"
done
echo ""

if [[ "$AUTO_YES" == "false" ]]; then
  read -p "Download ${#NEED_DOWNLOAD[@]} voice model(s)? [Y/n]: " -n 1 -r
  echo

  if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ -n $REPLY ]]; then
    echo "❌ Download cancelled"
    exit 0
  fi
else
  echo "Auto-downloading ${#NEED_DOWNLOAD[@]} voice model(s)..."
  echo ""
fi

DOWNLOADED=0
FAILED=0

for voice in "${NEED_DOWNLOAD[@]}"; do
  echo ""
  echo "📥 Downloading: $voice..."

  if download_voice "$voice"; then
    ((DOWNLOADED++))
    local voice_path="$VOICE_DIR/${voice}.onnx"
    local file_size=$(du -h "$voice_path" 2>/dev/null | cut -f1)
    echo "   ✓ Downloaded: $voice"
    echo "   📁 Path: $voice_path"
    echo "   📦 Size: $file_size"
  else
    ((FAILED++))
    echo "   ✗ Failed: $voice"
  fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Download Summary:"
echo ""
echo "Installed voices:"
for voice in "${ALREADY_DOWNLOADED_LIST[@]}"; do
  local voice_path="$VOICE_DIR/${voice}.onnx"
  local file_size=$(du -h "$voice_path" 2>/dev/null | cut -f1)
  echo "   ✓ $voice ($file_size)"
  echo "     $voice_path"
done

if [[ $DOWNLOADED -gt 0 ]]; then
  echo ""
  echo "Just downloaded:"
  for voice in "${NEED_DOWNLOAD[@]}"; do
    local voice_path="$VOICE_DIR/${voice}.onnx"
    if [[ -f "$voice_path" ]]; then
      local file_size=$(du -h "$voice_path" 2>/dev/null | cut -f1)
      echo "   ✓ $voice ($file_size)"
      echo "     $voice_path"
    fi
  done
fi

if [[ $FAILED -gt 0 ]]; then
  echo ""
  echo "Failed downloads:"
  for voice in "${NEED_DOWNLOAD[@]}"; do
    local voice_path="$VOICE_DIR/${voice}.onnx"
    if [[ ! -f "$voice_path" ]]; then
      echo "   ✗ $voice"
    fi
  done
fi

echo ""
echo "Total: $((ALREADY_DOWNLOADED + DOWNLOADED)) voices available"
echo ""

if [[ $DOWNLOADED -gt 0 ]]; then
  echo "✨ Ready to use Piper TTS with downloaded voices!"
  echo ""
  echo "Try it:"
  echo "  /agent-vibes:provider switch piper"
  echo "  /agent-vibes:preview"
fi

exit 0
