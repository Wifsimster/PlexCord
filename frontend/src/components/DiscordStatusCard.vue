<script setup>
import { computed } from 'vue';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import Button from 'primevue/button';

const discordStore = useDiscordConnectionStore();

// Computed status color
const statusColorClass = computed(() => {
    if (discordStore.hasError) return 'text-red-500 bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
    if (discordStore.connected) return 'text-green-500 bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800';
    return 'text-yellow-500 bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
});

const iconClass = computed(() => {
    if (discordStore.hasError) return 'pi pi-exclamation-circle';
    if (discordStore.connected) return 'pi pi-check-circle';
    return 'pi pi-info-circle';
});
</script>

<template>
    <div class="bg-surface-50 dark:bg-surface-800/50 rounded-xl border border-surface-200 dark:border-surface-700 p-5 flex flex-col transition-all hover:border-surface-300 dark:hover:border-surface-600">
        <!-- Header -->
        <div class="flex items-start justify-between mb-4">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 rounded-lg bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-500 shrink-0">
                    <i class="pi pi-discord text-xl"></i>
                </div>
                <div>
                    <h3 class="font-bold text-surface-900 dark:text-surface-0 leading-tight">Discord Client</h3>
                    <div class="flex items-center gap-1.5 mt-1">
                        <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium border" :class="statusColorClass">
                            <i :class="[iconClass, 'text-[10px]']"></i>
                            {{ discordStore.statusLabel }}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Details -->
        <div class="space-y-3 flex-grow">
            <!-- RPC Status -->
            <div class="flex items-center justify-between text-sm py-2 border-b border-surface-200 dark:border-surface-700/50 border-dashed">
                <span class="text-surface-500 dark:text-surface-400">Rich Presence</span>
                <span class="font-medium text-surface-900 dark:text-surface-100">
                    {{ discordStore.connected ? 'Active' : 'Inactive' }}
                </span>
            </div>

            <!-- Last Connected -->
            <div class="flex items-center justify-between text-sm py-2">
                <span class="text-surface-500 dark:text-surface-400">Last Synced</span>
                <span class="font-mono text-xs text-surface-600 dark:text-surface-300 bg-surface-100 dark:bg-surface-700 px-2 py-1 rounded">
                    {{ discordStore.lastConnectedRelative }}
                </span>
            </div>
        </div>

        <!-- Actions -->
        <div v-if="!discordStore.connected" class="mt-5 pt-4 border-t border-surface-200 dark:border-surface-700">
            <Button label="Connect" icon="pi pi-link" severity="secondary" size="small" :loading="discordStore.loading || discordStore.isRetrying" @click="discordStore.retry()" class="w-full" outlined />
        </div>
    </div>
</template>
