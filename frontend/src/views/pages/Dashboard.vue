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
    <div class="grid grid-cols-12 gap-6">
        <!-- Error Banners - Positioned at top -->
        <div v-if="errors.length > 0" class="col-span-12 space-y-3">
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

        <!-- Header -->
        <div class="col-span-12">
            <h1 class="text-2xl font-bold text-surface-900 dark:text-surface-0">Dashboard</h1>
            <p class="text-muted-color">PlexCord status at a glance</p>
        </div>

        <!-- Discord Preview Widget - Full width on mobile, half on desktop -->
        <div class="col-span-12 lg:col-span-6">
            <DiscordPreview />
        </div>

        <!-- Connection Status Widget -->
        <div class="col-span-12 lg:col-span-6">
            <ConnectionStatus />
        </div>
    </div>
</template>
