<script setup>
import Badge from 'primevue/badge';

/**
 * MetricDisplay Component
 * Reusable component for displaying status metrics with icons.
 * Pure presentation component with no business logic.
 *
 * @prop {String} icon - PrimeIcons class name
 * @prop {String} label - Metric label text
 * @prop {String|Number} value - Metric value to display
 * @prop {String} severity - Badge severity (null, 'secondary', 'info', 'success', 'warn', 'danger', 'contrast')
 * @prop {Boolean} showBadge - Whether to display value as badge
 */
defineProps({
    icon: {
        type: String,
        required: true
    },
    label: {
        type: String,
        default: ''
    },
    value: {
        type: [String, Number],
        required: true
    },
    severity: {
        type: String,
        default: null,
        validator: (value) => [null, 'secondary', 'info', 'success', 'warn', 'danger', 'contrast'].includes(value)
    },
    showBadge: {
        type: Boolean,
        default: false
    }
});
</script>

<template>
    <div class="flex items-center gap-2">
        <i :class="[icon, 'text-sm', 'text-surface-600', 'dark:text-surface-400']"></i>

        <span v-if="label" class="text-sm text-surface-700 dark:text-surface-300"> {{ label }}: </span>

        <Badge v-if="showBadge" :value="value" :severity="severity" size="small" />

        <span v-else class="text-sm text-surface-700 dark:text-surface-300">
            {{ value }}
        </span>
    </div>
</template>
