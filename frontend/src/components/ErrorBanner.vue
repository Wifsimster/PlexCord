<script setup>
import { computed, ref, onUnmounted } from 'vue';
import Button from 'primevue/button';
import Message from 'primevue/message';

/**
 * ErrorBanner Component
 * Displays connection errors prominently but non-intrusively at the top of the dashboard.
 * Supports retry functionality, countdown display, and dismissal.
 */
const props = defineProps({
    /** Error information from backend */
    errorInfo: {
        type: Object,
        required: true
    },
    /** Current retry state from backend */
    retryState: {
        type: Object,
        default: null
    },
    /** Source of the error: 'plex' or 'discord' */
    source: {
        type: String,
        required: true,
        validator: (value) => ['plex', 'discord'].includes(value)
    },
    /** Whether a retry is currently in progress */
    isRetrying: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['dismiss', 'retry']);

// Local state for dismiss animation
const isDismissing = ref(false);
let dismissTimer = null;

// Cleanup timer on unmount to prevent memory leak
onUnmounted(() => {
    if (dismissTimer) {
        clearTimeout(dismissTimer);
        dismissTimer = null;
    }
});

// Computed properties
const sourceIcon = computed(() => {
    return props.source === 'plex' ? 'pi pi-server' : 'pi pi-discord';
});

const sourceLabel = computed(() => {
    return props.source === 'plex' ? 'Plex' : 'Discord';
});

const sourceColor = computed(() => {
    return props.source === 'plex' ? 'text-orange-500' : 'text-indigo-500';
});

const countdownSeconds = computed(() => {
    if (!props.retryState?.isRetrying || !props.retryState?.nextRetryIn) {
        return 0;
    }
    // nextRetryIn is in nanoseconds, convert to seconds
    return Math.ceil(props.retryState.nextRetryIn / 1000000000);
});

const showCountdown = computed(() => {
    return props.retryState?.isRetrying && countdownSeconds.value > 0;
});

const attemptNumber = computed(() => {
    return props.retryState?.attemptNumber || 0;
});

// Methods
const handleDismiss = () => {
    isDismissing.value = true;
    // Small delay for animation
    dismissTimer = setTimeout(() => {
        emit('dismiss');
        dismissTimer = null;
    }, 150);
};

const handleRetry = () => {
    emit('retry');
};
</script>

<template>
    <Message :closable="false" severity="error" role="alert" class="transition-[opacity,transform] duration-150 ease-out" :class="{ 'opacity-0 scale-95': isDismissing }">
        <template #messageicon>
            <i :class="[sourceIcon, sourceColor, 'text-lg mr-2']"></i>
        </template>

        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 w-full">
            <!-- Error content -->
            <div class="flex-1 min-w-0">
                <!-- Title with source label -->
                <div class="font-semibold text-surface-900 dark:text-surface-0">
                    <span class="mr-1">{{ sourceLabel }}:</span>
                    {{ errorInfo.title || 'Connection Error' }}
                </div>

                <!-- Description -->
                <div v-if="errorInfo.description" class="text-sm text-surface-700 dark:text-surface-300 mt-1">
                    {{ errorInfo.description }}
                </div>

                <!-- Suggestion -->
                <div v-if="errorInfo.suggestion" class="text-sm text-surface-600 dark:text-surface-400 mt-1 italic">
                    {{ errorInfo.suggestion }}
                </div>

                <!-- Error code for troubleshooting -->
                <div class="text-xs text-surface-600 dark:text-surface-400 mt-2 font-mono">Code: {{ errorInfo.code || 'UNKNOWN' }}</div>

                <!-- Retry countdown -->
                <div v-if="showCountdown" class="text-sm text-surface-600 dark:text-surface-400 mt-2 flex items-center gap-2">
                    <i class="pi pi-spin pi-spinner text-xs"></i>
                    <span> Retry #{{ attemptNumber }} in {{ countdownSeconds }}s... </span>
                </div>
            </div>

            <!-- Action buttons -->
            <div class="flex items-center gap-2 shrink-0">
                <!-- Retry button (if retryable) -->
                <Button v-if="errorInfo.retryable" label="Retry" icon="pi pi-refresh" severity="secondary" size="small" :loading="isRetrying" @click="handleRetry" />

                <!-- Dismiss button -->
                <Button icon="pi pi-times" severity="secondary" text rounded size="small" @click="handleDismiss" v-tooltip.left="'Dismiss'" aria-label="Dismiss error" />
            </div>
        </div>
    </Message>
</template>

<style scoped>
/* Ensure Message component spans full width */
:deep(.p-message-wrapper) {
    width: 100%;
}

:deep(.p-message-content) {
    width: 100%;
}
</style>
