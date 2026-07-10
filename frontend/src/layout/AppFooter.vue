<script setup>
import { useVersion } from '@/composables/useVersion';

const { display, buildDate } = useVersion();
</script>

<template>
    <footer class="layout-footer">
        <span class="pc-chip-mono" :title="buildDate ? $t('footer.buildDate', { date: buildDate }) : undefined">{{ display }}</span>
    </footer>
</template>

<style scoped>
/* One quiet caption line (spec §5.1) — the only version display. */
.layout-footer {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px 0 12px;
}
</style>
