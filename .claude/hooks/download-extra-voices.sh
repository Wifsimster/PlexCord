#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/piper-voice-manager.sh"

AUTO_YES=false
if [[ "$1" == "--yes" ]] || [[ "$1" == "-y" ]]; then
  AUTO_YES=true
fi

HUGGINGFACE_REPO="agentvibes/piper-custom-voices"
HUGGINGFACE_BASE_URL="https://huggingface.co/${HUGGINGFACE_REPO}/resolve/main"

EXTRA_VOICES=(
  "kristin:Kristin (US English female, Public Domain, 64MB)"
  "jenny:Jenny (UK English female with Irish accent, CC BY, 64MB)"
  "16Speakers:Tracy/16Speakers (Multi-speaker: 12 US + 4 UK voices, Public Domain, 77MB)"
)

echo "🎙️  AgentVibes Extra Voice Downloader"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "This will download high-quality custom Piper voices from HuggingFace."
echo ""
echo "📦 Voices available:"
for voice_info in "${EXTRA_VOICES[@]}"; do
  voice_name="${voice_info%%:*}"
  voice_desc="${voice_info#*:}"
  echo "  • $voice_desc"
done
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

for voice_info in "${EXTRA_VOICES[@]}"; do
  voice_name="${voice_info%%:*}"
  voice_desc="${voice_info#*:}"

  if [[ -f "$VOICE_DIR/${voice_name}.onnx" ]] && [[ -f "$VOICE_DIR/${voice_name}.onnx.json" ]]; then
    ((ALREADY_DOWNLOADED++))
    ALREADY_DOWNLOADED_LIST+=("$voice_desc")
  else
    NEED_DOWNLOAD+=("$voice_info")
  fi
done

echo "📊 Status:"
echo "   Already downloaded: $ALREADY_DOWNLOADED voice(s)"
echo "   Need to download: ${#NEED_DOWNLOAD[@]} voice(s)"
echo ""

if [[ $ALREADY_DOWNLOADED -gt 0 ]]; then
  echo "✅ Already downloaded (skipped):"
  for voice_desc in "${ALREADY_DOWNLOADED_LIST[@]}"; do
    echo "   ✓ $voice_desc"
  done
  echo ""
fi

if [[ ${#NEED_DOWNLOAD[@]} -eq 0 ]]; then
  echo "🎉 All extra voices already downloaded!"
  exit 0
fi

echo "Voices to download:"
for voice_info in "${NEED_DOWNLOAD[@]}"; do
  voice_desc="${voice_info#*:}"
  echo "  • $voice_desc"
done
echo ""

TOTAL_SIZE_MB=0
for voice_info in "${NEED_DOWNLOAD[@]}"; do
  voice_desc="${voice_info#*:}"
  if [[ "$voice_desc" =~ ([0-9]+)MB ]]; then
    TOTAL_SIZE_MB=$((TOTAL_SIZE_MB + ${BASH_REMATCH[1]}))
  fi
done

echo "💾 Total download size: ~${TOTAL_SIZE_MB}MB"
echo ""

if [[ "$AUTO_YES" == "false" ]]; then
  read -p "Download ${#NEED_DOWNLOAD[@]} extra voice(s)? [Y/n]: " -n 1 -r
  echo

  if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ -n $REPLY ]]; then
    echo "❌ Download cancelled"
    exit 0
  fi
else
  echo "Auto-downloading ${#NEED_DOWNLOAD[@]} extra voice(s)..."
  echo ""
fi

mkdir -p "$VOICE_DIR"

download_voice_file() {
  local url="$1"
  local output_path="$2"
  local file_name="$3"

  echo "  📥 Downloading $file_name..."

  if curl -L --progress-bar "$url" -o "$output_path" 2>&1; then
    echo "  ✅ Downloaded: $file_name"
    return 0
  else
    echo "  ❌ Failed to download: $file_name"
    return 1
  fi
}

DOWNLOADED=0
FAILED=0

for voice_info in "${NEED_DOWNLOAD[@]}"; do
  voice_name="${voice_info%%:*}"
  voice_desc="${voice_info#*:}"

  echo ""
  echo "📥 Downloading: ${voice_desc%%,*}..."
  echo ""

  onnx_url="${HUGGINGFACE_BASE_URL}/${voice_name}.onnx"
  onnx_path="${VOICE_DIR}/${voice_name}.onnx"

  json_url="${HUGGINGFACE_BASE_URL}/${voice_name}.onnx.json"
  json_path="${VOICE_DIR}/${voice_name}.onnx.json"

  success=true

  if ! download_voice_file "$onnx_url" "$onnx_path" "${voice_name}.onnx"; then
    success=false
  fi

  if ! download_voice_file "$json_url" "$json_path" "${voice_name}.onnx.json"; then
    success=false
  fi

  if [[ "$success" == "true" ]]; then
    ((DOWNLOADED++))
    echo ""
    echo "✅ Successfully downloaded: ${voice_desc%%,*}"
  else
    ((FAILED++))
    echo ""
    echo "❌ Failed to download: ${voice_desc%%,*}"
    rm -f "$onnx_path" "$json_path"
  fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Download Summary:"
echo "   ✅ Successfully downloaded: $DOWNLOADED"
echo "   ❌ Failed: $FAILED"
echo "   📦 Total extra voices available: $((ALREADY_DOWNLOADED + DOWNLOADED))"
echo ""

if [[ $DOWNLOADED -gt 0 ]]; then
  echo "✨ Extra voices ready to use!"
  echo ""
  echo "Try them:"
  echo "  /agent-vibes:provider switch piper"
  echo "  /agent-vibes:switch kristin"
  echo "  /agent-vibes:switch jenny"
  echo "  /agent-vibes:switch 16Speakers"
fi

if [[ $DOWNLOADED -gt 0 ]] || [[ $ALREADY_DOWNLOADED -gt 0 ]]; then
  exit 0
else
  exit 1
fi
