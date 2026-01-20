<script setup>
import { useSetupStore } from '@/stores/setup';

const setupStore = useSetupStore();
</script>

<template>
    <div class="setup-step">
        <div class="step-content">
            <h2 class="text-3xl font-bold mb-4">Welcome to PlexCord</h2>
            <p class="text-lg mb-6">
                PlexCord bridges your Plex Media Server with Discord Rich Presence, letting your friends see what you're watching or listening to.
            </p>

            <div class="feature-list mb-8">
                <h3 class="text-xl font-semibold mb-3">What you'll need:</h3>
                <ul class="list-disc list-inside space-y-2">
                    <li>A Plex account and media server</li>
                    <li>Discord desktop application</li>
                    <li>Just a few minutes to set everything up</li>
                </ul>
            </div>

            <p class="text-muted-color mb-4">
                This wizard will guide you through the setup process step by step.
            </p>
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

.feature-list {
    background: var(--surface-ground);
    padding: 1.5rem;
    border-radius: 8px;
}
</style>
