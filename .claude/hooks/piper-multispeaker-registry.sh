#!/usr/bin/env bash

_to_lower() {
  echo "$1" | tr '[:upper:]' '[:lower:]'
}

MULTISPEAKER_VOICES=(
  "Cori_Samuel:16Speakers:0:US English Female"
  "Kara_Shallenberg:16Speakers:1:US English Female"
  "Kristin_Hughes:16Speakers:2:US English Female"
  "Maria_Kasper:16Speakers:3:US English Female"
  "Mike_Pelton:16Speakers:4:US English Male"
  "Mark_Nelson:16Speakers:5:US English Male"
  "Michael_Scherer:16Speakers:6:US English Male"
  "James_K_White:16Speakers:7:US English Male"
  "Rose_Ibex:16Speakers:8:US English Female"
  "progressingamerica:16Speakers:9:US English Male"
  "Steve_C:16Speakers:10:US English Male"
  "Owlivia:16Speakers:11:US English Female"

  "Paul_Hampton:16Speakers:12:UK English Male"
  "Jennifer_Dorr:16Speakers:13:UK English Female"
  "Emily_Cripps:16Speakers:14:UK English Female"
  "Martin_Clifton:16Speakers:15:UK English Male"
)

get_multispeaker_info() {
  local speaker_name="$1"
  for entry in "${MULTISPEAKER_VOICES[@]}"; do
    name="${entry%%:*}"
    rest="${entry#*:}"
    model="${rest%%:*}"
    rest="${rest#*:}"
    speaker_id="${rest%%:*}"

    if [[ "$(_to_lower "$name")" == "$(_to_lower "$speaker_name")" ]]; then
      echo "$model:$speaker_id"
      return 0
    fi
  done
  return 1
}

list_multispeaker_voices() {
  echo "🎭 Multi-Speaker Voices (16Speakers Model):"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  local current_model=""
  for entry in "${MULTISPEAKER_VOICES[@]}"; do
    name="${entry%%:*}"
    rest="${entry#*:}"
    model="${rest%%:*}"
    rest="${rest#*:}"
    speaker_id="${rest%%:*}"
    description="${rest#*:}"

    if [[ "$model" != "$current_model" ]]; then
      if [[ -n "$current_model" ]]; then
        echo ""
      fi
      echo "  Model: $model.onnx"
      current_model="$model"
    fi

    echo "    • $name (ID: $speaker_id) - $description"
  done

  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo ""
  echo "Usage: /agent-vibes:switch Cori_Samuel"
  echo "       /agent-vibes:switch Rose_Ibex"
}

get_multispeaker_description() {
  local speaker_name="$1"
  for entry in "${MULTISPEAKER_VOICES[@]}"; do
    name="${entry%%:*}"
    rest="${entry#*:}"
    rest="${rest#*:}"
    rest="${rest#*:}"
    description="${rest}"

    if [[ "$(_to_lower "$name")" == "$(_to_lower "$speaker_name")" ]]; then
      echo "$description"
      return 0
    fi
  done
  return 1
}
