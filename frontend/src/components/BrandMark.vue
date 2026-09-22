<script setup>
import BrandSymbol from '@/components/BrandSymbol.vue';

// Single logo lockup (spec §5.0.5): the PlexCord symbol + the "PlexCord"
// wordmark, used by the topbar and the wizard rail. See BrandSymbol.vue and
// docs/brand.md for the symbol itself ("the live badge").
defineProps({
    /** Optional uppercase suffix after the wordmark (e.g. 'SETUP'). */
    suffix: { type: String, default: '' },
    /** Symbol size in px (the small optical size below 32). */
    size: { type: Number, default: 20 }
});
</script>

<template>
    <span class="pc-brandmark">
        <BrandSymbol :size="size" />
        <span class="pc-brandmark-name">PlexCord</span>
        <span v-if="suffix" class="pc-brandmark-suffix">{{ suffix }}</span>
    </span>
</template>

<style scoped>
.pc-brandmark {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    line-height: 1;
}
.pc-brandmark-name {
    font-family: var(--pc-font-display);
    font-size: 16px;
    font-weight: 700;
    color: var(--pc-text);
    letter-spacing: -0.025em;
}
.pc-brandmark-suffix {
    font-size: var(--pc-text-micro);
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--pc-text-muted);
}
</style>
