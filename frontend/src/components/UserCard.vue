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
        :class="[
            'cursor-pointer transition-all duration-200 hover:-translate-y-0.5 hover:shadow-xl min-w-45',
            { 'border-2 border-primary-500 bg-primary-50 dark:bg-primary-900/20': selected }
        ]"
        @click="selectUser"
    >
        <template #content>
            <div class="flex flex-col items-center text-center gap-3 p-0">
                <Avatar
                    v-if="user.thumb"
                    :image="user.thumb"
                    size="xlarge"
                    shape="circle"
                    class="w-20! h-20!"
                />
                <Avatar
                    v-else
                    icon="pi pi-user"
                    size="xlarge"
                    shape="circle"
                    class="w-20! h-20! bg-surface-200 dark:bg-surface-700 text-surface-600 dark:text-surface-400"
                />
                <div class="flex flex-col items-center gap-1">
                    <h3 class="m-0 text-base font-semibold overflow-hidden text-ellipsis whitespace-nowrap max-w-37.5">
                        {{ user.name || `User ${user.id}` }}
                    </h3>
                    <span v-if="selected" class="flex items-center gap-1 text-xs text-primary-500">
                        <i class="pi pi-check-circle text-sm"></i>
                        Selected
                    </span>
                </div>
            </div>
        </template>
    </Card>
</template>
