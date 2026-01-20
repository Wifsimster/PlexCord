<script setup>
import { useSetupStore } from '@/stores/setup';

const setupStore = useSetupStore();
</script>

<template>
    <div class="setup-step">
        <div class="step-content">
            <h2 class="text-3xl font-bold mb-4">Connect to Discord</h2>
            <p class="text-lg mb-6">
                Discord configuration will be implemented in future stories.
            </p>

            <div class="placeholder-notice">
                <i class="pi pi-info-circle mr-2"></i>
                <span>Placeholder for Story 3.x: Discord Rich Presence Configuration</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.setup-step {
    max-width: 600px;
    margin: 0 auto;
}

.step-content {
    padding: 2rem 0;
}

.placeholder-notice {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
    display: flex;
    align-items: center;
    color: var(--text-color-secondary);
}
</style>
