<script setup>
import { computed } from 'vue';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import Card from 'primevue/card';
import Badge from 'primevue/badge';
import Button from 'primevue/button';

const plexStore = usePlexConnectionStore();

// Computed border color
const borderColor = computed(() => {
    if (plexStore.hasError) return 'border-red-500';
    if (plexStore.connected && plexStore.polling) return 'border-green-500';
    return 'border-yellow-500';
});

// Computed badge severity
const statusSeverity = computed(() => {
    if (plexStore.hasError) return 'danger';
    if (plexStore.connected && plexStore.polling) return 'success';
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
                    <i class="pi pi-server text-3xl text-surface-700 dark:text-surface-300"></i>
                </div>

                <!-- Service name and status -->
                <div class="flex items-center justify-between mb-3">
                    <h3 class="text-xl font-semibold text-surface-900 dark:text-surface-50">Plex</h3>
                    <Badge :value="plexStore.statusLabel" :severity="statusSeverity" size="small" />
                </div>

                <!-- Connection details -->
                <div class="space-y-2 mb-4">
                    <!-- User info -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-user text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            User connected
                        </span>
                    </div>

                    <!-- Last connected -->
                    <div class="flex items-center gap-2">
                        <i class="pi pi-clock text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300">
                            Last: {{ plexStore.lastConnectedRelative }}
                        </span>
                    </div>
                </div>

                <!-- Retry button -->
                <div v-if="!plexStore.connected || plexStore.hasError" class="mt-auto">
                    <Button
                        label="Retry"
                        icon="pi pi-refresh"
                        severity="secondary"
                        size="small"
                        :loading="plexStore.loading || plexStore.isRetrying"
                        @click="plexStore.retry()"
                        class="w-full"
                    />
                </div>
            </div>
        </template>
    </Card>
</template>
