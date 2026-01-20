<script setup>
import { onMounted, onUnmounted } from 'vue';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import PlexStatusCard from '@/components/PlexStatusCard.vue';
import DiscordStatusCard from '@/components/DiscordStatusCard.vue';
import Button from 'primevue/button';

const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();

// Initialize stores
onMounted(async () => {
    await Promise.all([plexStore.initialize(), discordStore.initialize()]);
});

// Cleanup on unmount
onUnmounted(() => {
    plexStore.cleanup();
    discordStore.cleanup();
});

// Refresh both connections
const refresh = async () => {
    await Promise.all([plexStore.refreshStatus(), discordStore.refreshStatus()]);
};
</script>

<template>
    <div class="h-full flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-full bg-green-50 dark:bg-green-900/30 flex items-center justify-center text-green-500">
                    <i class="pi pi-server text-xl"></i>
                </div>
                <div>
                    <h2 class="text-lg font-bold text-surface-900 dark:text-surface-0">System Status</h2>
                    <p class="text-sm text-surface-500 dark:text-surface-400">Connection health</p>
                </div>
            </div>
            <Button icon="pi pi-refresh" severity="secondary" text rounded class="hover:bg-surface-100 dark:hover:bg-surface-700 transition-colors" @click="refresh()" v-tooltip.left="'Refresh status'" />
        </div>

        <div class="grid grid-cols-1 xl:grid-cols-2 gap-6 w-full flex-grow">
            <PlexStatusCard class="h-full" />
            <DiscordStatusCard class="h-full" />
        </div>
    </div>
</template>
