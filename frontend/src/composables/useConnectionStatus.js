import { computed, onMounted, onUnmounted } from 'vue';
import { storeToRefs } from 'pinia';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';

/**
 * Composable for managing combined connection status
 * Provides unified access to Plex and Discord connection states
 * with lifecycle management and computed properties.
 *
 * @returns {Object} Connection status and control methods
 */
export function useConnectionStatus() {
    const plexStore = usePlexConnectionStore();
    const discordStore = useDiscordConnectionStore();

    // Get reactive refs from stores
    const { connected: plexConnected, loading: plexLoading, error: plexError, lastConnectedRelative: plexLastConnected, isRetrying: plexRetrying, statusLabel: plexStatus } = storeToRefs(plexStore);

    const { connected: discordConnected, loading: discordLoading, error: discordError, lastConnectedRelative: discordLastConnected, isRetrying: discordRetrying, statusLabel: discordStatus } = storeToRefs(discordStore);

    // Computed properties combining both stores
    const allConnected = computed(() => plexConnected.value && discordConnected.value);

    const anyConnected = computed(() => plexConnected.value || discordConnected.value);

    const hasErrors = computed(() => plexStore.hasError || discordStore.hasError);

    const isLoading = computed(() => plexLoading.value || discordLoading.value);

    const connectionHealth = computed(() => {
        if (plexConnected.value && discordConnected.value) return 'healthy';
        if (plexStore.hasError || discordStore.hasError) return 'error';
        if (!plexConnected.value || !discordConnected.value) return 'partial';
        return 'unknown';
    });

    // Combined error list for display
    const errors = computed(() => {
        const errorList = [];
        if (plexError.value) {
            errorList.push({ source: 'plex', ...plexError.value });
        }
        if (discordError.value) {
            errorList.push({ source: 'discord', ...discordError.value });
        }
        return errorList;
    });

    // Initialize both stores
    const initialize = async () => {
        await Promise.all([plexStore.initialize(), discordStore.initialize()]);
    };

    // Cleanup both stores
    const cleanup = () => {
        plexStore.cleanup();
        discordStore.cleanup();
    };

    // Lifecycle hooks
    onMounted(() => initialize());
    onUnmounted(() => cleanup());

    return {
        // Plex state
        plex: {
            connected: plexConnected,
            loading: plexLoading,
            error: plexError,
            lastConnected: plexLastConnected,
            isRetrying: plexRetrying,
            status: plexStatus,
            retry: () => plexStore.retry()
        },

        // Discord state
        discord: {
            connected: discordConnected,
            loading: discordLoading,
            error: discordError,
            lastConnected: discordLastConnected,
            isRetrying: discordRetrying,
            status: discordStatus,
            retry: () => discordStore.retry(),
            connect: (clientId) => discordStore.connect(clientId)
        },

        // Combined state
        allConnected,
        anyConnected,
        hasErrors,
        isLoading,
        connectionHealth,
        errors,

        // Control methods
        initialize,
        cleanup,
        refresh: async () => {
            await Promise.all([plexStore.refreshStatus(), discordStore.refreshStatus()]);
        }
    };
}
