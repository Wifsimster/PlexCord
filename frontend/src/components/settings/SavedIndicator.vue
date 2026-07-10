<script setup>
/**
 * Inline "✓ Saved" autosave indicator (spec §5.3 save model, motion M15).
 * Check scales 0→1 with the snap easing, the parent holds it visible for
 * ~1600ms, then it fades out over 240ms. Replaces toasts for routine saves
 * (F10/F36). Reduced motion: appears/disappears instantly (global policy).
 */
defineProps({
    visible: { type: Boolean, default: false },
    label: { type: String, default: '' }
});
</script>

<template>
    <Transition name="pc-saved">
        <span v-if="visible" class="pc-saved" role="status">
            <i class="pi pi-check" aria-hidden="true"></i>
            <span>{{ label || $t('common.saved') }}</span>
        </span>
    </Transition>
</template>

<style scoped>
.pc-saved {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: var(--pc-text-caption);
    line-height: 1.45;
    color: var(--pc-success);
    white-space: nowrap;
}
.pc-saved .pi {
    font-size: 11px;
}

/* M15 — check scales 0→1 (snap, 180ms); fade-out 240ms */
.pc-saved-enter-active {
    transition:
        transform 180ms var(--pc-ease-snap),
        opacity 180ms var(--pc-ease-out);
}
.pc-saved-enter-from {
    transform: scale(0);
    opacity: 0;
}
.pc-saved-leave-active {
    transition: opacity 240ms var(--pc-ease-in);
}
.pc-saved-leave-to {
    opacity: 0;
}
</style>
