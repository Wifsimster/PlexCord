#!/usr/bin/env bash

set -euo pipefail

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔄 AgentVibes Configuration Migration${NC}"
echo ""
echo "Migrating from .claude/config/ and .claude/plugins/ to .agentvibes/"
echo ""

if [[ -n "${CLAUDE_PROJECT_DIR:-}" ]]; then
    PROJECT_ROOT="$CLAUDE_PROJECT_DIR"
else
    PROJECT_ROOT="$(pwd)"
fi

cd "$PROJECT_ROOT"

MIGRATED=false

echo -e "${BLUE}📁 Creating .agentvibes/ directory structure...${NC}"
mkdir -p .agentvibes/config
echo -e "${GREEN}✓ Directories created${NC}"
echo ""

echo -e "${BLUE}🔍 Checking for AgentVibes config in .claude/config/...${NC}"

if [[ -f ".claude/config/agentvibes.json" ]]; then
    echo -e "${YELLOW}  Found: agentvibes.json${NC}"
    mv .claude/config/agentvibes.json .agentvibes/config/
    echo -e "${GREEN}  ✓ Moved to .agentvibes/config/${NC}"
    MIGRATED=true
fi

if [[ -f ".claude/config/personality-voice-defaults.default.json" ]]; then
    echo -e "${YELLOW}  Found: personality-voice-defaults.default.json${NC}"
    mv .claude/config/personality-voice-defaults.default.json .agentvibes/config/
    echo -e "${GREEN}  ✓ Moved to .agentvibes/config/${NC}"
    MIGRATED=true
fi

if [[ -f ".claude/config/personality-voice-defaults.json" ]]; then
    echo -e "${YELLOW}  Found: personality-voice-defaults.json${NC}"
    mv .claude/config/personality-voice-defaults.json .agentvibes/config/
    echo -e "${GREEN}  ✓ Moved to .agentvibes/config/${NC}"
    MIGRATED=true
fi

if [[ -f ".claude/config/README-personality-defaults.md" ]]; then
    echo -e "${YELLOW}  Found: README-personality-defaults.md${NC}"
    mv .claude/config/README-personality-defaults.md .agentvibes/config/
    echo -e "${GREEN}  ✓ Moved to .agentvibes/config/${NC}"
    MIGRATED=true
fi

echo ""

echo -e "${BLUE}🧹 Cleaning up...${NC}"

if [[ -d ".claude/plugins" ]] && [[ -z "$(ls -A .claude/plugins 2>/dev/null)" ]]; then
    rmdir .claude/plugins
    echo -e "${GREEN}✓ Removed empty .claude/plugins/ directory${NC}"
fi

echo ""

if [[ "$MIGRATED" == "true" ]]; then
    echo -e "${GREEN}✅ Migration complete!${NC}"
    echo ""
    echo "Your AgentVibes configuration has been moved to:"
    echo "  .agentvibes/config/  - AgentVibes settings"
    echo ""
    echo "Old locations are no longer used:"
    echo "  .claude/plugins/     - (removed if empty)"
    echo "  .claude/config/      - (AgentVibes files removed)"
    echo ""
    echo -e "${BLUE}ℹ️  Note: .claude/config/ still exists for runtime state files${NC}"
    echo "   (like tts-speech-rate.txt - these belong to Claude Code)"
else
    echo -e "${GREEN}✓ No migration needed${NC}"
    echo ""
    echo "All configuration is already in .agentvibes/"
fi

echo ""
echo -e "${GREEN}🎉 Ready to use AgentVibes!${NC}"
