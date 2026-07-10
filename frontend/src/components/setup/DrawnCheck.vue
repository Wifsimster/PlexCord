<script setup>
// M19 — success check draw (spec §4): SVG check stroke-dashoffset draw
// 320ms --pc-ease-out + optional circle fill fade 180ms delayed 120ms.
// Reduced motion: the global policy collapses the animation, so the
// check simply appears (per the M19 fallback).
defineProps({
    /** Rendered size in px (square). */
    size: { type: Number, default: 14 },
    /** Draw the success-dim circle behind the check. */
    circle: { type: Boolean, default: false }
});
</script>

<template>
    <svg class="pc-drawn-check" :width="size" :height="size" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <circle v-if="circle" cx="12" cy="12" r="11" class="check-circle" />
        <path d="M6.5 12.5l3.8 3.8L17.5 8" class="check-path" />
    </svg>
</template>

<style scoped>
.pc-drawn-check {
    display: inline-block;
    flex: none;
    vertical-align: middle;
}
.check-path {
    stroke: var(--pc-success);
    stroke-width: 2.5;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-dasharray: 18;
    stroke-dashoffset: 18;
    animation: pc-check-draw 320ms var(--pc-ease-out) forwards;
}
.check-circle {
    fill: var(--pc-success-dim);
    opacity: 0;
    animation: pc-check-fill 180ms var(--pc-ease-out) 120ms forwards;
}
@keyframes pc-check-draw {
    to {
        stroke-dashoffset: 0;
    }
}
@keyframes pc-check-fill {
    to {
        opacity: 1;
    }
}
</style>
