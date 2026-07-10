import { describe, it, expect, beforeEach } from 'vitest';
import { nextTick } from 'vue';
import { mount } from '@vue/test-utils';
import TruncatedText from '@/components/TruncatedText.vue';

// jsdom computes no layout, so scrollWidth/clientWidth are both 0 and
// ResizeObserver is absent. We capture the RO callback to re-drive measurement
// after stubbing element widths, and record the v-tooltip binding via a stub
// directive so we can assert whether the tooltip was enabled.
let roCallback = null;
let lastTooltip = null;

const tooltipStub = {
    mounted(el, binding) {
        lastTooltip = binding.value;
    },
    updated(el, binding) {
        lastTooltip = binding.value;
    }
};

const stubWidths = (el, { scroll, client }) => {
    Object.defineProperty(el, 'scrollWidth', { configurable: true, value: scroll });
    Object.defineProperty(el, 'clientWidth', { configurable: true, value: client });
};

const mountText = (props) =>
    mount(TruncatedText, {
        props,
        global: { directives: { tooltip: tooltipStub } }
    });

beforeEach(() => {
    roCallback = null;
    lastTooltip = null;
    global.ResizeObserver = class {
        constructor(cb) {
            roCallback = cb;
        }
        observe() {}
        disconnect() {}
    };
});

describe('TruncatedText', () => {
    it('renders the text into the requested host element', () => {
        const wrapper = mountText({ text: 'Pour Some Sugar On Me', as: 'dd' });
        expect(wrapper.element.tagName).toBe('DD');
        expect(wrapper.text()).toBe('Pour Some Sugar On Me');
    });

    it('leaves the tooltip disabled and adds no tab stop when the text fits', async () => {
        const wrapper = mountText({ text: 'short' });
        stubWidths(wrapper.element, { scroll: 40, client: 100 });
        roCallback?.([]);
        await nextTick();

        expect(lastTooltip.disabled).toBe(true);
        expect(wrapper.element.hasAttribute('tabindex')).toBe(false);
    });

    it('enables the tooltip and exposes a tab stop when the text is clipped', async () => {
        const wrapper = mountText({ text: 'a very long value that overflows its box' });
        stubWidths(wrapper.element, { scroll: 300, client: 100 });
        roCallback?.([]);
        await nextTick();

        expect(lastTooltip.disabled).toBe(false);
        expect(lastTooltip.value).toBe('a very long value that overflows its box');
        expect(wrapper.element.getAttribute('tabindex')).toBe('0');
    });

    it('ignores a sub-pixel width difference (no phantom overflow)', async () => {
        const wrapper = mountText({ text: 'rounding' });
        stubWidths(wrapper.element, { scroll: 101, client: 100 });
        roCallback?.([]);
        await nextTick();

        expect(lastTooltip.disabled).toBe(true);
    });
});
