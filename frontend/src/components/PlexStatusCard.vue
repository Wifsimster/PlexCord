<script setup>
import { computed } from 'vue';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import Button from 'primevue/button';

const plexStore = usePlexConnectionStore();

// Computed status color
const statusColorClass = computed(() => {
    if (plexStore.hasError) return 'text-red-500 bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
    if (plexStore.connected && plexStore.polling) return 'text-green-500 bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800';
    return 'text-yellow-500 bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
});

const iconClass = computed(() => {
    if (plexStore.hasError) return 'pi pi-exclamation-circle';
    if (plexStore.connected && plexStore.polling) return 'pi pi-check-circle';
    return 'pi pi-info-circle';
});
</script>

<template>
    <div class="bg-surface-50 dark:bg-surface-800/50 rounded-xl border border-surface-200 dark:border-surface-700 p-5 flex flex-col transition-all hover:border-surface-300 dark:hover:border-surface-600">
        <!-- Header -->
        <div class="flex items-start justify-between mb-4">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-500 shrink-0">
                    <svg class="w-6 h-6" viewBox="0 0 24 24" fill="none" width="24" height="24" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M9 12l2 2 4-4" />
                        <path d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" v-if="plexStore.connected && !plexStore.hasError" />
                        <circle cx="12" cy="12" r="10" v-else />
                    </svg>
                </div>
                <div>
                    <h3 class="font-bold text-surface-900 dark:text-surface-0 leading-tight">Plex Media Server</h3>
                    <div class="flex items-center gap-1.5 mt-1">
                        <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border" :class="statusColorClass">
                            <i :class="[iconClass, 'text-[10px]']"></i>
                            {{ plexStore.statusLabel }}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Details -->
        <div class="space-y-3 flex-grow">
            <!-- User -->
            <div class="flex items-center justify-between text-sm py-2 border-b border-surface-200 dark:border-surface-700/50 border-dashed">
                <span class="text-surface-500 dark:text-surface-400">Account</span>
                <span class="font-medium text-surface-900 dark:text-surface-100 truncate max-w-[150px]">
                    {{ plexStore.connected ? 'Connected' : '-' }}
                </span>
            </div>

            <!-- Last Connected -->
            <div class="flex items-center justify-between text-sm py-2">
                <span class="text-surface-500 dark:text-surface-400">Last Synced</span>
                <span class="font-mono text-xs text-surface-600 dark:text-surface-300 bg-surface-100 dark:bg-surface-700 px-2 py-1 rounded">
                    {{ plexStore.lastConnectedRelative }}
                </span>
            </div>
        </div>

        <!-- Actions -->
        <div v-if="!plexStore.connected || plexStore.hasError" class="mt-5 pt-4 border-t border-surface-200 dark:border-surface-700">
            <Button label="Reconnect" icon="pi pi-refresh" severity="secondary" size="small" :loading="plexStore.loading || plexStore.isRetrying" @click="plexStore.retry()" class="w-full" outlined />
        </div>
    </div>
</template>
