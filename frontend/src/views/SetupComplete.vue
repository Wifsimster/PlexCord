<script setup>
import { ref, onMounted } from 'vue';
import { useSetupStore } from '@/stores/setup';
import { usePlayback } from '@/composables/usePlayback';
import { GetPresenceFormat } from '../../wailsjs/go/main/App';
import DiscordSpecimen from '@/components/DiscordSpecimen.vue';
import DrawnCheck from '@/components/setup/DrawnCheck.vue';

const setupStore = useSetupStore();
const { currentTrack, isPaused } = usePlayback();

const formats = ref(null);

onMounted(async () => {
    try {
        formats.value = await GetPresenceFormat();
    } catch (error) {
        console.error('Failed to load presence format:', error);
    }
});
</script>

<template>
    <div>
        <div class="complete-hero">
            <DrawnCheck :size="40" circle />
            <h1 class="setup-title complete-title">{{ $t('complete.title') }}</h1>
            <p class="setup-lede complete-lede">{{ $t('complete.lede') }}</p>
        </div>

        <div class="setup-panels">
            <section v-if="setupStore.finishError" class="pc-panel complete-error" role="alert">
                <p class="complete-error-title"><i class="pi pi-times-circle" aria-hidden="true"></i> {{ $t('complete.errTitle') }}</p>
                <p class="complete-error-text">{{ setupStore.finishError }}</p>
                <p class="complete-error-suggestion">{{ $t('complete.errSuggestion') }}</p>
            </section>

            <section v-if="setupStore.discordSkipped && !setupStore.discordConnected" class="pc-panel complete-warn">
                <p class="complete-warn-text"><i class="pi pi-exclamation-triangle" aria-hidden="true"></i> {{ $t('complete.discordSkipped') }}</p>
            </section>

            <section class="pc-panel">
                <span class="pc-eyebrow complete-eyebrow">{{ $t('complete.yourPresence') }}</span>
                <DiscordSpecimen :track="currentTrack" :formats="formats" :paused="isPaused" :idle-title="$t('complete.idleTitle')" />
            </section>

            <section class="pc-panel">
                <span class="pc-eyebrow complete-eyebrow">{{ $t('complete.whatNext') }}</span>
                <ul class="next-list">
                    <li class="next-row">
                        <i class="pi pi-check next-glyph" aria-hidden="true"></i>
                        <span>{{ $t('complete.next1') }}</span>
                    </li>
                    <li class="next-row">
                        <i class="pi pi-check next-glyph" aria-hidden="true"></i>
                        <span>{{ $t('complete.next2') }}</span>
                    </li>
                    <li class="next-row">
                        <i class="pi pi-check next-glyph" aria-hidden="true"></i>
                        <span>{{ $t('complete.next3') }}</span>
                    </li>
                </ul>
            </section>
        </div>
    </div>
</template>

<style scoped>
.complete-hero {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: var(--pc-space-section);
}
.complete-title {
    margin: 0;
}
.complete-lede {
    margin: 0;
}
.complete-eyebrow {
    display: block;
    margin-bottom: 12px;
}

.complete-error {
    border-color: color-mix(in srgb, var(--pc-danger) 40%, transparent);
    background: var(--pc-danger-dim);
}
.complete-error-title {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 4px;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-danger);
}
.complete-error-text {
    margin: 0 0 4px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    overflow-wrap: anywhere;
}
.complete-error-suggestion {
    margin: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}

.complete-warn {
    border-color: color-mix(in srgb, var(--pc-warn) 40%, transparent);
    background: var(--pc-warn-dim);
}
.complete-warn-text {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-warn);
}

.next-list {
    list-style: none;
    margin: 0;
    padding: 0;
}
.next-row {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 12px 0;
    font-size: var(--pc-text-body);
    color: var(--pc-text-secondary);
}
.next-row + .next-row {
    border-top: 1px solid var(--pc-border-subtle);
}
.next-glyph {
    font-size: 14px;
    color: var(--pc-success);
    flex: none;
}
</style>
