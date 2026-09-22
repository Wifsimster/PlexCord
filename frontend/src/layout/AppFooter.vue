<script setup>
import { useVersion } from '@/composables/useVersion';

const { display, buildDate } = useVersion();
</script>

<template>
    <footer class="layout-footer">
        <span class="footer-version" :title="buildDate ? $t('footer.buildDate', { date: buildDate }) : undefined">{{ display }}</span>
    </footer>
</template>

<style scoped>
/* One quiet caption line (spec §5.1) — the only version display. */
.layout-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px 0 14px;
}
.footer-version {
    font-family: var(--pc-font-mono);
    font-size: 11.5px;
    color: var(--pc-text-faint);
    font-variant-numeric: tabular-nums;
    letter-spacing: 0.02em;
}
</style>
