<script setup>
import Card from 'primevue/card';
import Avatar from 'primevue/avatar';

const props = defineProps({
    user: {
        type: Object,
        required: true
    },
    selected: {
        type: Boolean,
        default: false
    }
});

const emit = defineEmits(['select']);

const selectUser = () => {
    emit('select', props.user);
};
</script>

<template>
    <Card
        :class="['user-card', { 'selected': selected }]"
        @click="selectUser"
    >
        <template #content>
            <div class="user-content">
                <Avatar
                    v-if="user.thumb"
                    :image="user.thumb"
                    size="xlarge"
                    shape="circle"
                    class="user-avatar"
                />
                <Avatar
                    v-else
                    icon="pi pi-user"
                    size="xlarge"
                    shape="circle"
                    class="user-avatar default-avatar"
                />
                <div class="user-info">
                    <h3 class="user-name">{{ user.name }}</h3>
                    <span v-if="selected" class="selected-indicator">
                        <i class="pi pi-check-circle"></i>
                        Selected
                    </span>
                </div>
            </div>
        </template>
    </Card>
</template>

<style scoped>
.user-card {
    cursor: pointer;
    transition: all 0.2s;
    min-width: 180px;
}

.user-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.user-card.selected {
    border: 2px solid var(--primary-color);
    background: var(--primary-50);
}

:deep(.p-card-body) {
    padding: 1rem;
}

:deep(.p-card-content) {
    padding: 0;
}

.user-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.75rem;
}

.user-avatar {
    width: 80px !important;
    height: 80px !important;
}

.user-avatar.default-avatar {
    background: var(--surface-200);
    color: var(--text-color-secondary);
}

.user-info {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
}

.user-name {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--text-color);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 150px;
}

.selected-indicator {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
    color: var(--primary-color);
}

.selected-indicator i {
    font-size: 0.875rem;
}
</style>
