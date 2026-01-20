<script setup>
import { computed, onMounted, onUnmounted } from 'vue';
import { useConnectionStore } from '@/stores/connection';
import Button from 'primevue/button';

const connectionStore = useConnectionStore();

// Initialize on mount
onMounted(() => {
    connectionStore.initialize();
});

// Cleanup on unmount
onUnmounted(() => {
    connectionStore.cleanup();
});

// Computed properties
const plexStatus = computed(() => connectionStore.plex);
const discordStatus = computed(() => connectionStore.discord);
const plexLastConnected = computed(() => connectionStore.plexLastConnectedRelative);
const discordLastConnected = computed(() => connectionStore.discordLastConnectedRelative);
const isPlexRetrying = computed(() => connectionStore.isPlexRetrying);
const isDiscordRetrying = computed(() => connectionStore.isDiscordRetrying);

// Status indicator classes
const getStatusText = (connected, inError = false) => {
    if (inError) return 'Error';
    if (connected) return 'Connected';
    return 'Disconnected';
};

// Retry handlers
const handleRetryPlex = () => {
    connectionStore.retryPlex();
};

const handleRetryDiscord = () => {
    connectionStore.retryDiscord();
};
</script>

<template>
    <div>
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
            <h2 class="text-2xl font-bold tracking-tight text-surface-900 dark:text-surface-50">
                Connection Status
            </h2>
            <Button
                icon="pi pi-refresh"
                severity="secondary"
                text
                rounded
                class="hover:bg-surface-100 dark:hover:bg-surface-700 transition-colors"
                @click="connectionStore.refreshStatus()"
                v-tooltip.left="'Refresh status'"
            />
        </div>

        <div class="grid grid-cols-2 gap-6 w-full">
            <!-- Plex Status Card -->
            <div class="border-2 rounded-lg p-6 bg-surface-50 dark:bg-surface-900"
                :class="[
                    plexStatus.inErrorState ? 'border-red-500' :
                    plexStatus.connected ? 'border-green-500' :
                    'border-yellow-500'
                ]"
            >
                <div class="flex flex-col h-full">
                    <!-- Icon -->
                    <div class="mb-4">
                        <i class="pi pi-server text-3xl text-surface-700 dark:text-surface-300"></i>
                    </div>

                    <!-- Service name -->
                    <h3 class="text-xl font-semibold mb-3 text-surface-900 dark:text-surface-50">
                        Plex
                    </h3>

                    <!-- Status -->
                    <div class="mb-3">
                        <span class="text-base font-medium text-surface-900 dark:text-surface-50">
                            {{ getStatusText(plexStatus.connected, plexStatus.inErrorState) }}
                        </span>
                    </div>

                    <!-- User info -->
                    <div class="flex items-center gap-2 mb-2">
                        <i class="pi pi-user text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            <span v-if="plexStatus.userName">{{ plexStatus.userName }}</span>
                            <span v-else class="italic">No user connected</span>
                        </span>
                    </div>

                    <!-- Last connected -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-clock text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Last: {{ plexLastConnected }}
                        </span>
                    </div>

                    <!-- Retry button -->
                    <div v-if="!plexStatus.connected || plexStatus.inErrorState" class="mt-4">
                        <Button
                            label="Retry"
                            icon="pi pi-refresh"
                            severity="secondary"
                            size="small"
                            :loading="isPlexRetrying || connectionStore.loading.plex"
                            @click="handleRetryPlex"
                        />
                    </div>

                    <!-- Retry state info -->
                    <div
                        v-if="plexStatus.retryState?.isRetrying"
                        class="mt-3 pt-3 border-t border-surface-200 dark:border-surface-700 flex items-center gap-2 text-sm text-surface-600 dark:text-surface-400"
                    >
                        <i class="pi pi-spin pi-spinner"></i>
                        <span>
                            Retry #{{ plexStatus.retryState.attemptNumber }} - 
                            Next in {{ Math.ceil(plexStatus.retryState.nextRetryIn / 1000000000) }}s
                        </span>
                    </div>
                </div>
            </div>

            <!-- Discord Status Card -->
            <div class="border-2 rounded-lg p-6 bg-surface-50 dark:bg-surface-900"
                :class="[
                    discordStatus.error ? 'border-red-500' :
                    discordStatus.connected ? 'border-green-500' :
                    'border-yellow-500'
                ]"
            >
                <div class="flex flex-col h-full">
                    <!-- Icon -->
                    <div class="mb-4">
                        <i class="pi pi-discord text-3xl text-surface-700 dark:text-surface-300"></i>
                    </div>

                    <!-- Service name -->
                    <h3 class="text-xl font-semibold mb-3 text-surface-900 dark:text-surface-50">
                        Discord
                    </h3>

                    <!-- Status -->
                    <div class="mb-3">
                        <span class="text-base font-medium text-surface-900 dark:text-surface-50">
                            {{ getStatusText(discordStatus.connected, !!discordStatus.error) }}
                        </span>
                    </div>

                    <!-- Rich presence info -->
                    <div class="flex items-center gap-2 mb-2">
                        <i class="pi pi-bolt text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Rich Presence {{ discordStatus.connected ? 'Active' : 'Inactive' }}
                        </span>
                    </div>

                    <!-- Last connected -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-clock text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Last: {{ discordLastConnected }}
                        </span>
                    </div>

                    <!-- Connect button -->
                    <div v-if="!discordStatus.connected" class="mt-4">
                        <Button
                            label="Connect"
                            icon="pi pi-link"
                            severity="secondary"
                            size="small"
                            :loading="isDiscordRetrying || connectionStore.loading.discord"
                            @click="handleRetryDiscord"
                        />
                    </div>

                    <!-- Retry state info -->
                    <div
                        v-if="discordStatus.retryState?.isRetrying"
                        class="mt-3 pt-3 border-t border-surface-200 dark:border-surface-700 flex items-center gap-2 text-sm text-surface-600 dark:text-surface-400"
                    >
                        <i class="pi pi-spin pi-spinner"></i>
                        <span>
                            Retry #{{ discordStatus.retryState.attemptNumber }} - 
                            Next in {{ Math.ceil(discordStatus.retryState.nextRetryIn / 1000000000) }}s
                        </span>
                    </div>

                    <!-- Error message -->
                    <div
                        v-if="discordStatus.error"
                        class="mt-3 pt-3 border-t border-red-200 dark:border-red-900/30 flex items-start gap-2 text-sm"
                    >
                        <i class="pi pi-exclamation-triangle text-red-500 mt-0.5"></i>
                        <span class="text-red-700 dark:text-red-400">
                            {{ discordStatus.error.message || discordStatus.error }}
                        </span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
