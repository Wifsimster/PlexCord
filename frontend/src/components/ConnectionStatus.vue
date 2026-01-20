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
const getStatusClass = (connected, inError = false) => {
    if (inError) return 'bg-red-500';
    if (connected) return 'bg-green-500';
    return 'bg-yellow-500';
};

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
    <div class="card">
        <div class="flex items-center justify-between mb-4">
            <span class="text-xl font-semibold">Connection Status</span>
            <Button
                icon="pi pi-refresh"
                severity="secondary"
                text
                rounded
                @click="connectionStore.refreshStatus()"
                v-tooltip.left="'Refresh status'"
            />
        </div>

        <div class="space-y-4">
            <!-- Plex Status -->
            <div class="flex items-center justify-between p-3 rounded-lg bg-surface-100 dark:bg-surface-800">
                <div class="flex items-center gap-3">
                    <!-- Status indicator -->
                    <div
                        :class="[
                            'w-3 h-3 rounded-full',
                            getStatusClass(plexStatus.connected, plexStatus.inErrorState)
                        ]"
                    ></div>

                    <!-- Service info -->
                    <div>
                        <div class="font-medium flex items-center gap-2">
                            <i class="pi pi-server text-orange-500"></i>
                            Plex
                        </div>
                        <div class="text-sm text-muted-color">
                            {{ getStatusText(plexStatus.connected, plexStatus.inErrorState) }}
                            <span v-if="plexStatus.userName" class="ml-1">
                                - {{ plexStatus.userName }}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- Right side: retry button or last connected -->
                <div class="text-right">
                    <Button
                        v-if="!plexStatus.connected || plexStatus.inErrorState"
                        label="Retry"
                        icon="pi pi-refresh"
                        severity="secondary"
                        size="small"
                        :loading="isPlexRetrying || connectionStore.loading.plex"
                        @click="handleRetryPlex"
                    />
                    <div v-else class="text-xs text-muted-color">
                        Last: {{ plexLastConnected }}
                    </div>
                </div>
            </div>

            <!-- Retry state info for Plex -->
            <div
                v-if="plexStatus.retryState?.isRetrying"
                class="ml-6 text-sm text-muted-color flex items-center gap-2"
            >
                <i class="pi pi-spin pi-spinner"></i>
                <span>
                    Retry #{{ plexStatus.retryState.attemptNumber }} -
                    Next in {{ Math.ceil(plexStatus.retryState.nextRetryIn / 1000000000) }}s
                </span>
            </div>

            <!-- Discord Status -->
            <div class="flex items-center justify-between p-3 rounded-lg bg-surface-100 dark:bg-surface-800">
                <div class="flex items-center gap-3">
                    <!-- Status indicator -->
                    <div
                        :class="[
                            'w-3 h-3 rounded-full',
                            getStatusClass(discordStatus.connected, !!discordStatus.error)
                        ]"
                    ></div>

                    <!-- Service info -->
                    <div>
                        <div class="font-medium flex items-center gap-2">
                            <i class="pi pi-discord text-indigo-500"></i>
                            Discord
                        </div>
                        <div class="text-sm text-muted-color">
                            {{ getStatusText(discordStatus.connected, !!discordStatus.error) }}
                        </div>
                    </div>
                </div>

                <!-- Right side: retry button or last connected -->
                <div class="text-right">
                    <Button
                        v-if="!discordStatus.connected"
                        label="Connect"
                        icon="pi pi-link"
                        severity="secondary"
                        size="small"
                        :loading="isDiscordRetrying || connectionStore.loading.discord"
                        @click="handleRetryDiscord"
                    />
                    <div v-else class="text-xs text-muted-color">
                        Last: {{ discordLastConnected }}
                    </div>
                </div>
            </div>

            <!-- Retry state info for Discord -->
            <div
                v-if="discordStatus.retryState?.isRetrying"
                class="ml-6 text-sm text-muted-color flex items-center gap-2"
            >
                <i class="pi pi-spin pi-spinner"></i>
                <span>
                    Retry #{{ discordStatus.retryState.attemptNumber }} -
                    Next in {{ Math.ceil(discordStatus.retryState.nextRetryIn / 1000000000) }}s
                </span>
            </div>

            <!-- Error message if any -->
            <div
                v-if="discordStatus.error"
                class="p-3 rounded-lg bg-red-100 dark:bg-red-900/20 text-red-700 dark:text-red-400 text-sm"
            >
                <i class="pi pi-exclamation-triangle mr-2"></i>
                {{ discordStatus.error.message || discordStatus.error }}
            </div>
        </div>
    </div>
</template>
