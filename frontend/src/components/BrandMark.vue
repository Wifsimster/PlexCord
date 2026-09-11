<script setup>
// Single logo lockup (spec §5.0.5): the PlexCord symbol + the "PlexCord"
// wordmark. Kills the duplicated inline SVGs in AppTopbar/SetupWizard.
//
// The symbol is the signal path compressed to three masses — two ends and the
// relay between them (see docs/brand.md). It renders in --pc-text, not in a
// brand pigment: gold and blurple identify the two ENDPOINTS of the path, so
// spending either on the product's own mark would break the color contract
// (§1.2) and make PlexCord look like the thing it connects to.
//
// This is the small optical size of the mark (shorter, thicker masses), which
// is the one that survives at the 18px the topbar renders it at; the regular
// geometry lives in build/brand/plexcord-mark.svg.
defineProps({
    /** Optional uppercase suffix after the wordmark (e.g. 'SETUP'). */
    suffix: { type: String, default: '' }
});
</script>

<template>
    <span class="pc-brandmark">
        <svg class="pc-brandmark-glyph" width="18" height="18" viewBox="0 0 64 64" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <rect x="2" y="26.5" width="13" height="11" rx="5.5" fill="currentColor" />
            <circle cx="32" cy="32" r="10.5" fill="currentColor" />
            <rect x="49" y="26.5" width="13" height="11" rx="5.5" fill="currentColor" />
        </svg>
        <span class="pc-brandmark-name">PlexCord</span>
        <span v-if="suffix" class="pc-brandmark-suffix">{{ suffix }}</span>
    </span>
</template>

<style scoped>
.pc-brandmark {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    line-height: 1;
}
.pc-brandmark-glyph {
    color: var(--pc-text);
    flex: none;
}
.pc-brandmark-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--pc-text);
    letter-spacing: -0.01em;
}
.pc-brandmark-suffix {
    font-size: var(--pc-text-micro);
    font-weight: 600;
    letter-spacing: 0.07em;
    text-transform: uppercase;
    color: var(--pc-text-muted);
}
</style>
