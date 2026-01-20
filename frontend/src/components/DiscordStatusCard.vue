<script setup>
import { computed } from 'vue';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import Card from 'primevue/card';
import Badge from 'primevue/badge';
import Button from 'primevue/button';

const discordStore = useDiscordConnectionStore();

// Computed border color
const borderColor = computed(() => {
    if (discordStore.hasError) return 'border-red-500';
    if (discordStore.connected) return 'border-green-500';
    return 'border-yellow-500';
});

// Computed badge severity
const statusSeverity = computed(() => {
    if (discordStore.hasError) return 'danger';
    if (discordStore.connected) return 'success';
    return 'warn';
});

// Computed pt config
const cardPt = computed(() => ({
    root: `border-2 ${borderColor.value} bg-surface-50 dark:bg-surface-900`,
    body: 'p-6',
    content: 'p-0'
}));
</script>

<template>
    <Card :pt="cardPt">
        <template #content>
            <div class="flex flex-col h-full">
                <!-- Icon -->
                <div class="mb-4">
                    <i class="pi pi-discord text-3xl text-surface-700 dark:text-surface-300"></i>
                </div>

                <!-- Service name and status -->
                <div class="flex items-center justify-between mb-3">
                    <h3 class="text-xl font-semibold text-surface-900 dark:text-surface-50">Discord</h3>
                    <Badge :value="discordStore.statusLabel" :severity="statusSeverity" size="small" />
                </div>

                <!-- Connection details -->
                <div class="space-y-2 mb-4">
                    <!-- Rich presence info -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-bolt text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Rich Presence {{ discordStore.connected ? 'Active' : 'Inactive' }}
                        </span>
                    </div>

                    <!-- Last connected -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-clock text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Last: {{ discordStore.lastConnectedRelative }}
                        </span>
                    </div>
                </div>

                <!-- Connect button -->
                <div v-if="!discordStore.connected" class="mt-auto">
                    <Button
                        label="Connect"
                        icon="pi pi-link"
                        severity="secondary"
                        size="small"
                        :loading="discordStore.loading || discordStore.isRetrying"
                        @click="discordStore.retry()"
                        class="w-full"
                    />
                </div>
            </div>
        </template>
    </Card>
</template>
