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
        :class="['server-card', { 'selected': isSelected }]"
        @click="selectServer"
    >
        <template #title>
            {{ server.name }}
        </template>
        <template #content>
            <div class="server-details">
                <p class="server-address">{{ server.address }}:{{ server.port }}</p>
                <Badge :severity="badgeType" :value="badgeLabel" />
            </div>
        </template>
    </Card>
</template>

<style scoped>
.server-card {
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 1rem;
}

.server-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.server-card.selected {
    border: 2px solid var(--primary-color);
}

.server-details {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.server-address {
    font-family: monospace;
    color: var(--text-color-secondary);
    margin: 0;
}
</style>
