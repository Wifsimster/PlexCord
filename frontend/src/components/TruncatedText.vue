<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';

/**
 * A single line of text that ellipsis-truncates when it runs out of room and —
 * only then — reveals its full value in a tooltip on hover/focus. It keeps two
 * things in one place so call sites (fact strip, connection tiles, presence
 * card) stay declarative: the overflow measurement, and the rule that a tooltip
 * appears only when the text is actually clipped. A value that already fits gets
 * no tooltip and no tab stop, so nothing is announced twice.
 */
const props = defineProps({
    /** The full text to render; also the tooltip content when truncated. */
    text: { type: [String, Number], default: '' },
    /** Host element tag — dd/dt/span/p are all valid single-line hosts. */
    as: { type: String, default: 'span' },
    /** Hover delay before the tooltip appears (app convention is 300ms). */
    showDelay: { type: Number, default: 300 }
});

const el = ref(null);
const overflowing = ref(false);

const measure = () => {
    const node = el.value;
    if (!node) return;
    // +1px guards against sub-pixel rounding reporting a phantom overflow.
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

// A changed value (new track, new host) can flip the overflow state without the
// box resizing; re-measure after the DOM paints the new text.
watch(
    () => props.text,
    () => requestAnimationFrame(measure)
);

// PrimeVue's directive reads `disabled` live, so we bind it always and let the
// measurement decide — cleaner than swapping the binding to null.
const tooltip = computed(() => ({
    value: String(props.text ?? ''),
    disabled: !overflowing.value,
    showDelay: props.showDelay
}));

// Only offer a tab stop while the value is clipped, so keyboard users can reach
// the tooltip without seeding the tab order with fully-visible captions.
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
