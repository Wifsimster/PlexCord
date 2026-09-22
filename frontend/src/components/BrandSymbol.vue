<script setup>
import { computed, useId } from 'vue';

// The PlexCord symbol — "the live badge" (docs/brand.md). A ring (a record,
// or a profile picture) in ink, with a presence badge on its corner in the
// tally. Drawn at two optical sizes, picked by size, not taste:
//   regular (≥ 32px): thinner ring, play triangle knocked out of the badge
//   small   (< 32px): thicker ring, bigger solid badge — survives at 16px
// Geometry is canonical (64-unit box) and mirrored in build/brand/.
const props = defineProps({
    size: { type: Number, default: 20 },
    /** Force an optical size; default picks by `size`. */
    variant: { type: String, default: '' },
    /** Unlit: ring only, no badge — the mark "off air". */
    unlit: { type: Boolean, default: false }
});

const isSmall = computed(() => (props.variant ? props.variant === 'small' : props.size < 32));
const uid = useId();
const gapId = `pc-sym-gap-${uid}`;
const playId = `pc-sym-play-${uid}`;
</script>

<template>
    <svg class="pc-symbol" :width="size" :height="size" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
        <template v-if="isSmall">
            <mask :id="gapId">
                <rect width="64" height="64" fill="#fff" />
                <circle v-if="!unlit" cx="46" cy="46" r="19" fill="#000" />
            </mask>
            <path class="pc-symbol-ring" :mask="`url(#${gapId})`" fill-rule="evenodd" d="M28 2a26 26 0 1 1 0 52a26 26 0 1 1 0-52zM28 19a9 9 0 1 0 0 18a9 9 0 1 0 0-18z" />
            <circle v-if="!unlit" class="pc-symbol-badge" cx="46" cy="46" r="14" />
        </template>
        <template v-else>
            <mask :id="gapId">
                <rect width="64" height="64" fill="#fff" />
                <circle v-if="!unlit" cx="47" cy="47" r="17.5" fill="#000" />
            </mask>
            <mask :id="playId">
                <rect width="64" height="64" fill="#fff" />
                <path d="M43.2 40.6 L53.4 47 L43.2 53.4 Z" fill="#000" stroke="#000" stroke-width="2" stroke-linejoin="round" />
            </mask>
            <path class="pc-symbol-ring" :mask="`url(#${gapId})`" fill-rule="evenodd" d="M29 4a25 25 0 1 1 0 50a25 25 0 1 1 0-50zM29 19a10 10 0 1 0 0 20a10 10 0 1 0 0-20z" />
            <circle v-if="!unlit" class="pc-symbol-badge" :mask="`url(#${playId})`" cx="47" cy="47" r="13" />
        </template>
    </svg>
</template>

<style scoped>
.pc-symbol {
    flex: none;
}
.pc-symbol-ring {
    fill: var(--pc-symbol-ink, var(--pc-text));
}
.pc-symbol-badge {
    fill: var(--pc-tally);
}
</style>
