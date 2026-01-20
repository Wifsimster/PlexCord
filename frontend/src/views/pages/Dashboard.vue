<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import DiscordPreview from '@/components/setup/DiscordPreview.vue';
import ConnectionStatus from '@/components/ConnectionStatus.vue';
import ErrorBanner from '@/components/ErrorBanner.vue';
import Button from 'primevue/button';
import { GetVersion } from '../../../wailsjs/go/main/App';
import { useConnectionStore } from '@/stores/connection';

const router = useRouter();
const connectionStore = useConnectionStore();

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

// Navigation
const goToSettings = () => {
    router.push('/settings');
};

// Error banner computed properties
const activeErrors = computed(() => connectionStore.activeErrors);

// Get retry state for a specific source
const getRetryState = (source) => {
    return source === 'plex'
        ? connectionStore.plex.retryState
        : connectionStore.discord.retryState;
};

// Check if retry is in progress for a source
const isRetrying = (source) => {
    return source === 'plex'
        ? connectionStore.loading.plex || connectionStore.isPlexRetrying
        : connectionStore.loading.discord || connectionStore.isDiscordRetrying;
};

// Handle error dismissal
const handleDismissError = (source) => {
    connectionStore.dismissError(source);
};

// Handle retry request
const handleRetry = (source) => {
    if (source === 'plex') {
        connectionStore.retryPlex();
    } else {
        connectionStore.retryDiscord();
    }
};
</script>

<template>
    <div class="grid grid-cols-12 gap-6">
        <!-- Error Banners - Positioned at top -->
        <div v-if="activeErrors.length > 0" class="col-span-12 space-y-3">
            <ErrorBanner
                v-for="error in activeErrors"
                :key="error.source"
                :error-info="error.errorInfo"
                :retry-state="getRetryState(error.source)"
                :source="error.source"
                :is-retrying="isRetrying(error.source)"
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
