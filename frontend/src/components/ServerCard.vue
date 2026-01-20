<script setup>
import { computed } from 'vue';
import Card from 'primevue/card';
import Badge from 'primevue/badge';

const props = defineProps({
    server: {
        type: Object,
        required: true
    },
    isSelected: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['server-selected']);

const badgeType = computed(() => props.server.isLocal ? 'success' : 'info');
const badgeLabel = computed(() => props.server.isLocal ? 'Local' : 'Remote');

const selectServer = () => {
    emit('server-selected', props.server);
};
</script>

<template>
    <Card
        :class="[
            'group cursor-pointer transition-all duration-200 hover:-translate-y-1 hover:shadow-lg',
            { 'ring-2 ring-primary-500 border-primary-500': isSelected }
        ]"
        @click="selectServer"
    >
        <template #title>
            <div class="flex items-center justify-between gap-3">
                <h4 class="font-semibold text-lg">{{ server.name }}</h4>
                <Badge :severity="badgeType" :value="badgeLabel" />
            </div>
        </template>
        <template #content>
            <div class="mt-3">
                <p class="font-mono text-sm text-surface-600 dark:text-surface-400">
                    {{ server.address }}:{{ server.port }}
                </p>
            </div>
        </template>
    </Card>
</template>
