<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';

const props = defineProps({
    text: { type: [String, Number], default: '' },
    as: { type: String, default: 'span' },
    showDelay: { type: Number, default: 300 }
});

const el = ref(null);
const overflowing = ref(false);

const measure = () => {
    const node = el.value;
    if (!node) return;
    overflowing.value = node.scrollWidth > node.clientWidth + 1;
};

let observer = null;
onMounted(() => {
    measure();
    if (typeof ResizeObserver !== 'undefined') {
        observer = new ResizeObserver(measure);
        observer.observe(el.value);
    }
});
onBeforeUnmount(() => observer?.disconnect());

watch(
    () => props.text,
    () => requestAnimationFrame(measure)
);

const tooltip = computed(() => ({
    value: String(props.text ?? ''),
    disabled: !overflowing.value,
    showDelay: props.showDelay
}));

const tabindex = computed(() => (overflowing.value ? 0 : undefined));
</script>

<template>
    <component :is="as" ref="el" v-tooltip.top="tooltip" class="truncated-text" :tabindex="tabindex">{{ text }}</component>
</template>

<style scoped>
.truncated-text {
    display: block;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
