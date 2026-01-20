<script setup>
import Card from 'primevue/card';
import Button from 'primevue/button';
import Badge from 'primevue/badge';

/**
 * StatusCard Component
 * Reusable card component for displaying service connection status.
 * Follows separation of concerns: presentation only, no business logic.
 *
 * @prop {String} title - Service name (e.g., "Plex", "Discord")
 * @prop {String} icon - PrimeIcons class name
 * @prop {Boolean} connected - Connection status
 * @prop {Boolean} hasError - Error state
 * @prop {Boolean} loading - Loading state
 * @prop {String} lastConnected - Last connection time (formatted)
 * @prop {String} statusLabel - Current status label
 * @prop {Object} metadata - Optional metadata to display (key-value pairs)
 * @prop {Function} onRetry - Retry callback function
 */
const props = defineProps({
    title: {
        type: String,
        required: true
    },
    icon: {
        type: String,
        required: true
    },
    connected: {
        type: Boolean,
        default: false
    },
    hasError: {
        type: Boolean,
        default: false
    },
    loading: {
        type: Boolean,
        default: false
    },
    lastConnected: {
        type: String,
        default: 'Never'
    },
    statusLabel: {
        type: String,
        default: 'Not Connected'
    },
    metadata: {
        type: Object,
        default: () => ({})
    },
    onRetry: {
        type: Function,
        default: null
    }
});

// Emit events for parent
const emit = defineEmits(['retry']);

// Handle retry click
const handleRetry = () => {
    if (props.onRetry) {
        props.onRetry();
    }
    emit('retry');
};

// Compute border color based on status
const borderColorClass = () => {
    if (props.hasError) return 'border-red-500';
    if (props.connected) return 'border-green-500';
    return 'border-yellow-500';
};

// Compute badge severity
const statusSeverity = () => {
    if (props.hasError) return 'danger';
    if (props.connected) return 'success';
    return 'warn';
};
</script>

<template>
    <Card
        :pt="{
            root: `border-2 ${borderColorClass()} bg-surface-50 dark:bg-surface-900`,
            body: 'p-6',
            content: 'p-0'
        }"
    >
        <template #content>
            <div class="flex flex-col h-full">
                <!-- Icon -->
                <div class="mb-4">
                    <i :class="[icon, 'text-3xl', 'text-surface-700', 'dark:text-surface-300']"></i>
                </div>

                <!-- Service name and status badge -->
                <div class="flex items-center justify-between mb-3">
                    <h3 class="text-xl font-semibold text-surface-900 dark:text-surface-50">
                        {{ title }}
                    </h3>
                    <Badge :value="statusLabel" :severity="statusSeverity()" size="small" />
                </div>

                <!-- Metadata -->
                <div v-if="Object.keys(metadata).length > 0" class="space-y-2 mb-3">
                    <div v-for="(value, key) in metadata" :key="key" class="flex items-center gap-2">
                        <i class="pi pi-info-circle text-sm text-surface-600 dark:text-surface-400"></i>
                        <span class="text-sm text-surface-700 dark:text-surface-300"> {{ key }}: {{ value }} </span>
                    </div>
                </div>

                <!-- Last connected -->
                <div class="flex items-center gap-2 mb-4">
                    <i class="pi pi-clock text-sm text-surface-600 dark:text-surface-400"></i>
                    <span class="text-sm text-surface-700 dark:text-surface-300"> Last: {{ lastConnected }} </span>
                </div>

                <!-- Retry button -->
                <div v-if="!connected || hasError" class="mt-auto">
                    <Button :label="connected ? 'Reconnect' : 'Retry'" icon="pi pi-refresh" severity="secondary" size="small" :loading="loading" @click="handleRetry" class="w-full" />
                </div>
            </div>
        </template>
    </Card>
</template>
