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
    <div>
        <!-- Header -->
        <div class="flex items-center justify-between mb-6">
            <h2 class="text-2xl font-bold tracking-tight text-surface-900 dark:text-surface-50">Connection Status</h2>
            <Button icon="pi pi-refresh" severity="secondary" text rounded class="hover:bg-surface-100 dark:hover:bg-surface-700 transition-colors" @click="refresh()" v-tooltip.left="'Refresh status'" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6 w-full">
            <PlexStatusCard />
            <DiscordStatusCard />
        </div>
    </div>
</template>
