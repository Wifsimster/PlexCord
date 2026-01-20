<script setup>
import { ref, onMounted } from 'vue';
import DiscordPreview from '@/components/setup/DiscordPreview.vue';
import ConnectionStatus from '@/components/ConnectionStatus.vue';
import ErrorBanner from '@/components/ErrorBanner.vue';
import { GetVersion } from '../../../wailsjs/go/main/App';
import { useConnectionStatus } from '@/composables/useConnectionStatus';

const { errors, plex, discord } = useConnectionStatus();

// Version info
const version = ref('');

onMounted(async () => {
    try {
        const versionInfo = await GetVersion();
        version.value = versionInfo.version;
    } catch (error) {
        console.error('Failed to get version:', error);
    }
});

// Error banner handling
const handleDismissError = (source) => {
    if (source === 'plex') {
        plex.error.value = null;
    } else {
        discord.error.value = null;
    }
};

const handleRetry = (source) => {
    if (source === 'plex') {
        plex.retry();
    } else {
        discord.retry();
    }
};
</script>

<template>
    <div class="min-h-screen bg-surface-50 dark:bg-surface-950 p-6 md:p-8">
        <div class="max-w-7xl mx-auto">
            <!-- Header -->
            <div class="mb-8 flex flex-col md:flex-row md:items-end justify-between gap-4">
                <div>
                    <h1 class="text-3xl font-bold text-surface-900 dark:text-surface-0 tracking-tight">Dashboard</h1>
                    <p class="text-surface-500 dark:text-surface-400 mt-1">Manage your PlexCord integration</p>
                </div>
                <div v-if="version" class="text-sm text-surface-400 dark:text-surface-500 font-mono">v{{ version }}</div>
            </div>

            <!-- Error Banners -->
            <div v-if="errors.length > 0" class="mb-8 space-y-4">
                <ErrorBanner
                    v-for="error in errors"
                    :key="error.source"
                    :error-info="error"
                    :retry-state="error.source === 'plex' ? plex.retryState : discord.retryState"
                    :source="error.source"
                    :is-retrying="error.source === 'plex' ? plex.isRetrying : discord.isRetrying"
                    @dismiss="handleDismissError(error.source)"
                    @retry="handleRetry(error.source)"
                />
            </div>

            <!-- Main Grid -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
                <!-- Left Column: Live Preview (5 columns) -->
                <div class="lg:col-span-5 flex flex-col">
                    <div class="bg-surface-0 dark:bg-surface-900 rounded-2xl shadow-sm border border-surface-200 dark:border-surface-800 p-6 flex-grow">
                        <div class="flex items-center gap-3 mb-6">
                            <div class="w-10 h-10 rounded-full bg-indigo-50 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-500">
                                <i class="pi pi-eye text-xl"></i>
                            </div>
                            <div>
                                <h2 class="text-lg font-bold text-surface-900 dark:text-surface-0">Live Preview</h2>
                                <p class="text-sm text-surface-500 dark:text-surface-400">What others see on Discord</p>
                            </div>
                        </div>

                        <div class="flex items-center justify-center py-8">
                            <DiscordPreview />
                        </div>
                    </div>
                </div>

                <!-- Right Column: Status & Controls (7 columns) -->
                <div class="lg:col-span-7 flex flex-col">
                    <div class="bg-surface-0 dark:bg-surface-900 rounded-2xl shadow-sm border border-surface-200 dark:border-surface-800 p-6 h-full">
                        <ConnectionStatus />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
