<script setup>
// Single logo lockup (spec §5.0.5): the Plexamp polyline glyph recolored
// to the Plex brand pigment + the "PlexCord" wordmark. Kills the duplicated
// inline SVGs in AppTopbar/SetupWizard.
defineProps({
    /** Optional uppercase suffix after the wordmark (e.g. 'SETUP'). */
    suffix: { type: String, default: '' }
});
</script>

<template>
    <span class="pc-brandmark">
        <svg class="pc-brandmark-glyph" width="18" height="18" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <polyline
                fill="none"
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.5"
                points="4.5 24 23.444 24 12.808 9.342 16.883 9.342 27.519 24 16.883 38.658 20.957 38.658 31.594 24 20.957 9.342 25.032 9.342 35.668 24 25.032 38.658 29.107 38.658 39.743 24 43.5 24"
            />
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
    color: var(--pc-plex);
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
