#!/usr/bin/env bash

set -e

NON_INTERACTIVE=false
if [[ "$1" == "--non-interactive" ]]; then
  NON_INTERACTIVE=true
fi

echo "🎤 Piper TTS Installer"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

PLATFORM="$(uname -s)"
ARCH="$(uname -m)"

if [[ "$PLATFORM" == "Darwin" ]]; then
  IS_MACOS=true
  echo "🍎 Detected macOS"
elif grep -qi microsoft /proc/version 2>/dev/null || [[ "$PLATFORM" == "Linux" ]]; then
  IS_MACOS=false
  echo "🐧 Detected Linux/WSL"
else
  echo "❌ Unsupported platform: $PLATFORM"
  echo ""
  echo "   For Windows, use macOS provider instead:"
  echo "   /agent-vibes:provider switch macos"
  exit 1
fi

if command -v piper &> /dev/null; then
  echo "✅ Piper TTS is already installed!"
  echo "   Location: $(which piper)"
  echo ""
  echo "   Download voices with: .claude/hooks/piper-download-voices.sh"
  exit 0
fi

echo "📦 Installing Piper TTS..."
echo ""

if [[ "$IS_MACOS" == true ]]; then
  echo "🍎 Installing Piper TTS from precompiled binaries for macOS..."
  echo ""

  if [[ "$ARCH" == "arm64" ]]; then
    echo "Detected Apple Silicon (M1/M2/M3)"
    PIPER_URL="https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_macos_aarch64.tar.gz"
  elif [[ "$ARCH" == "x86_64" ]]; then
    echo "Detected Intel Mac"
    PIPER_URL="https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_macos_x64.tar.gz"
  else
    echo "❌ Unsupported macOS architecture: $ARCH"
    exit 1
  fi

  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"

  echo "📥 Downloading Piper from: $PIPER_URL"
  TEMP_DIR=$(mktemp -d)
  cd "$TEMP_DIR"

  if curl -L "$PIPER_URL" | tar -xz; then
    echo "✅ Downloaded and extracted successfully"

    if [[ -d "piper" ]]; then
      cp -r piper/* "$INSTALL_DIR/"
      chmod +x "$INSTALL_DIR/piper"
      chmod +x "$INSTALL_DIR/piper_phonemize"

      echo "✅ Installed Piper to: $INSTALL_DIR/piper"

      if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo "⚠️  Add $INSTALL_DIR to your PATH:"
        echo ""
        if [[ "$SHELL" == *"zsh"* ]]; then
          echo "   echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
          echo "   source ~/.zshrc"
        else
          echo "   echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bash_profile"
          echo "   source ~/.bash_profile"
        fi
      fi
    else
      echo "❌ Failed to extract Piper binaries"
      cd -
      rm -rf "$TEMP_DIR"
      exit 1
    fi
  else
    echo "❌ Failed to download Piper"
    cd -
    rm -rf "$TEMP_DIR"
    exit 1
  fi

  cd -
  rm -rf "$TEMP_DIR"

# Linux/WSL: Use pipx
else
  if ! command -v pipx &> /dev/null; then
    echo "⚠️  pipx not found. Installing pipx first..."
    echo ""

    if command -v apt-get &> /dev/null; then
      sudo apt-get update
      sudo apt-get install -y pipx
    elif command -v brew &> /dev/null; then
      brew install pipx
    elif command -v dnf &> /dev/null; then
      sudo dnf install -y pipx
    elif command -v pacman &> /dev/null; then
      sudo pacman -S python-pipx
    else
      echo "❌ Unable to install pipx automatically."
      echo ""
      echo "   Please install pipx manually:"
      echo "   https://pipx.pypa.io/stable/installation/"
      exit 1
    fi

    pipx ensurepath
    echo ""
  fi

  echo "📥 Installing Piper TTS via pipx..."
  pipx install piper-tts

  if ! command -v piper &> /dev/null; then
    echo ""
    echo "❌ Installation completed but piper command not found in PATH"
    echo ""
    echo "   Try running: pipx ensurepath"
    echo "   Then restart your terminal"
    exit 1
  fi
fi

echo ""
echo "✅ Piper TTS installed successfully!"
echo ""

PIPER_VERSION=$($INSTALL_DIR/piper --version 2>&1 || echo "unknown")
echo "   Version: $PIPER_VERSION"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLAUDE_DIR="$(dirname "$SCRIPT_DIR")"

VOICES_DIR=""
if [[ -f "$CLAUDE_DIR/piper-voices-dir.txt" ]]; then
  VOICES_DIR=$(cat "$CLAUDE_DIR/piper-voices-dir.txt")
elif [[ -f "$HOME/.claude/piper-voices-dir.txt" ]]; then
  VOICES_DIR=$(cat "$HOME/.claude/piper-voices-dir.txt")
else
  VOICES_DIR="$HOME/.claude/piper-voices"
fi

echo "📁 Voice storage location: $VOICES_DIR"
echo ""

DOWNLOAD_VOICES=true
if [[ "$NON_INTERACTIVE" == "false" ]]; then
  read -p "Would you like to download voice models now? [Y/n] " -n 1 -r
  echo ""

  if [[ ! $REPLY =~ ^[Yy]$ ]] && [[ -n $REPLY ]]; then
    DOWNLOAD_VOICES=false
  fi
else
  echo "📥 Auto-downloading recommended voices (non-interactive mode)..."
  echo ""
fi

if [[ "$DOWNLOAD_VOICES" == "true" ]]; then
  echo ""
  echo "📥 Downloading recommended voices..."
  echo ""

  if [[ -f "$SCRIPT_DIR/piper-download-voices.sh" ]]; then
    "$SCRIPT_DIR/piper-download-voices.sh" --yes
  else
    mkdir -p "$VOICES_DIR"

    echo "Downloading en_US-lessac-medium (recommended)..."
    curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx" \
      -o "$VOICES_DIR/en_US-lessac-medium.onnx"
    curl -L "https://huggingface.co/rhasspy/piper-voices/resolve/main/en/en_US/lessac/medium/en_US-lessac-medium.onnx.json" \
      -o "$VOICES_DIR/en_US-lessac-medium.onnx.json"

    echo "✅ Voice downloaded!"
  fi
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🎉 Piper TTS Setup Complete!"
echo ""
echo "Next steps:"
echo "  1. Download more voices: .claude/hooks/piper-download-voices.sh"
echo "  2. List available voices: /agent-vibes:list"
echo "  3. Test it out: /agent-vibes:preview"
echo ""
echo "Enjoy your free, offline TTS! 🎤"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
