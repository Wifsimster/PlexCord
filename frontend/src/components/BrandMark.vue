<script setup>
import BrandSymbol from '@/components/BrandSymbol.vue';

defineProps({
    suffix: { type: String, default: '' },
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
