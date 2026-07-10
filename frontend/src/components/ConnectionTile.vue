<script>
import { reactive } from 'vue';

// Per-error-code collapse memory (module scope, spec §5.2 Crossfade graft):
// a detail the user collapsed stays collapsed for subsequent status events
// of the SAME failure, but a different error code opens again.
const collapsedByKey = reactive({});
</script>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePlaybackStore } from '@/stores/playback';
import { usePresenceStore } from '@/stores/presence';
import { formatRelativeTime } from '@/utils/timeUtils';

/**
 * Connections-panel tile (spec §5.2) — one component, `source` prop.
 * Replaces PlexStatusCard / DiscordStatusCard / ErrorBanner: brand identity
 * keyline on top, dot + micro status label (M4/M5/M6), 28px key-value rows,
 * and an inline collapsible error/retry detail (M17/M18) instead of a
 * dismissable banner. Errors resolve or collapse — never "dismiss" into lying.
 */
const props = defineProps({
    /** Which endpoint this tile represents. */
    source: {
        type: String,
        required: true,
        validator: (value) => ['plex', 'discord'].includes(value)
    }
});

const { t } = useI18n();

const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();
const playbackStore = usePlaybackStore();
const presenceStore = usePresenceStore();

const isPlex = props.source === 'plex';
const store = isPlex ? plexStore : discordStore;
const name = isPlex ? 'Plex' : 'Discord';

// ---- Status (dot + micro label, §5.0.3) -----------------------------------
const healthy = computed(() => (isPlex ? store.connected && store.polling : store.connected));

const statusLabel = computed(() => {
    if (store.isRetrying) return t('connectionTile.statusRetrying');
    if (store.hasError) return t('connectionTile.statusError');
    if (healthy.value) return t('connectionTile.statusConnected');
    if (store.loading) return t('connectionTile.statusConnecting');
    return t('connectionTile.statusIdle');
});

const statusSeverity = computed(() => {
    if (store.isRetrying) return 'warn';
    if (store.hasError) return 'danger';
    if (healthy.value) return 'success';
    if (store.loading) return 'warn';
    return 'idle';
});

const dotClass = computed(() => {
    const classes = [`pc-dot--${statusSeverity.value}`];
    if (store.isRetrying) classes.push('pc-dot--blink'); // M5
    if (statusSeverity.value === 'success' && playbackStore.isPlaying && !presenceStore.paused) {
        classes.push('pc-dot--pulse'); // M4 — connected & playing
    }
    return classes;
});

// ---- Key-value rows --------------------------------------------------------
const plexHost = computed(() => {
    if (!plexStore.serverUrl) return '—';
    return plexStore.serverUrl.replace(/^[a-z]+:\/\//i, '').replace(/\/+$/, '');
});

const discordPresence = computed(() => {
    if (!discordStore.connected) return t('presenceState.inactive');
    if (presenceStore.paused) return t('presenceState.hiddenPaused');
    return playbackStore.hasActiveSession ? t('presenceState.active') : t('presenceState.inactive');
});

// 1s ticker: keeps the relative sync label fresh and drives the M18 countdown.
const now = ref(Date.now());
let ticker = null;
onMounted(() => {
    ticker = setInterval(() => {
        now.value = Date.now();
    }, 1000);
});
onBeforeUnmount(() => {
    if (ticker) clearInterval(ticker);
});

const syncLabel = computed(() => {
    void now.value; // recompute as time passes
    return formatRelativeTime(store.lastConnected);
});

// ---- Inline error detail (replaces ErrorBanner — F5/F9) --------------------
const errorInfo = computed(() => {
    if (!store.hasError) return null;
    return (
        store.error || {
            code: store.retryState?.lastErrorCode || t('connectionTile.unknown'),
            title: t('connectionTile.connectionError'),
            suggestion: store.retryState?.lastError || ''
        }
    );
});

// Loading into an already-errored backend leaves hasError true (via
// inErrorState/retryState) with no friendly error object until the next
// failure event — hydrate it through the store's own GetErrorInfo path.
watch(
    () => [store.hasError, store.error, store.retryState?.lastErrorCode],
    ([hasError, error, lastErrorCode]) => {
        if (hasError && !error && lastErrorCode) {
            store.setError(lastErrorCode);
        }
    },
    { immediate: true }
);

// Retain the last error so the M17 collapse-out on recovery has content.
const displayError = ref(null);
watch(
    errorInfo,
    (error) => {
        if (error) displayError.value = error;
    },
    { immediate: true }
);

const errorKey = computed(() => (errorInfo.value ? `${props.source}:${errorInfo.value.code || 'UNKNOWN'}` : ''));
const detailOpen = computed(() => !!errorInfo.value && !collapsedByKey[errorKey.value]);
const toggleDetail = () => {
    if (!errorKey.value) return;
    collapsedByKey[errorKey.value] = !collapsedByKey[errorKey.value];
};

// ---- Retry countdown (M18 — store nextRetryIn is NANOSECONDS) --------------
const retryDeadline = ref(0);
watch(
    () => store.retryState,
    (state) => {
        if (!state?.isRetrying) {
            retryDeadline.value = 0;
            return;
        }
        const at = state.nextRetryAt ? Date.parse(state.nextRetryAt) : NaN;
        if (Number.isFinite(at) && at > Date.now()) {
            retryDeadline.value = at;
        } else if (state.nextRetryIn > 0) {
            retryDeadline.value = Date.now() + state.nextRetryIn / 1e6; // ns → ms
        } else {
            retryDeadline.value = 0;
        }
    },
    { immediate: true, deep: true }
);

const countdownSeconds = computed(() => {
    if (!retryDeadline.value) return 0;
    return Math.max(0, Math.ceil((retryDeadline.value - now.value) / 1000));
});
const showCountdown = computed(() => store.isRetrying && countdownSeconds.value > 0);
const attemptNumber = computed(() => store.retryState?.attemptNumber || 0);

const retryNow = () => store.retry();
</script>

<template>
    <article class="tile" :class="[`tile--${source}`, { 'tile--error': !!errorInfo }]">
        <header class="tile-header">
            <span class="tile-glyph" :class="`tile-glyph--${source}`" aria-hidden="true">
                <svg v-if="isPlex" width="16" height="16" viewBox="0 0 48 48" xmlns="http://www.w3.org/2000/svg">
                    <polyline
                        fill="none"
                        stroke="currentColor"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="3"
                        points="4.5 24 23.444 24 12.808 9.342 16.883 9.342 27.519 24 16.883 38.658 20.957 38.658 31.594 24 20.957 9.342 25.032 9.342 35.668 24 25.032 38.658 29.107 38.658 39.743 24 43.5 24"
                    />
                </svg>
                <i v-else class="pi pi-discord"></i>
            </span>
            <h3 class="tile-name">{{ name }}</h3>
            <span class="tile-status" :class="`tile-status--${statusSeverity}`">
                <span class="pc-dot" :class="dotClass" aria-hidden="true"></span>
                <Transition name="pc-state" mode="out-in">
                    <span :key="statusLabel" class="pc-status-label">{{ statusLabel }}</span>
                </Transition>
            </span>
            <button v-if="errorInfo" type="button" class="tile-expand" :aria-expanded="detailOpen" :aria-label="detailOpen ? $t('connectionTile.collapseDetails') : $t('connectionTile.expandDetails')" @click="toggleDetail">
                <i class="pi" :class="detailOpen ? 'pi-chevron-up' : 'pi-chevron-down'" aria-hidden="true"></i>
            </button>
        </header>

        <div class="tile-rows">
            <template v-if="isPlex">
                <div class="tile-row">
                    <span class="tile-row-label">{{ $t('connectionTile.labelAccount') }}</span>
                    <span class="tile-row-value">{{ plexStore.userName || $t('common.none') }}</span>
                </div>
                <div class="tile-row">
                    <span class="tile-row-label">{{ $t('connectionTile.labelServer') }}</span>
                    <span class="pc-chip-mono tile-row-chip">{{ plexHost }}</span>
                </div>
            </template>
            <template v-else>
                <div class="tile-row">
                    <span class="tile-row-label">{{ $t('connectionTile.labelPresence') }}</span>
                    <span class="tile-row-value">{{ discordPresence }}</span>
                </div>
            </template>
            <div class="tile-row">
                <span class="tile-row-label">{{ $t('connectionTile.labelSync') }}</span>
                <span class="tile-row-value">{{ syncLabel }}</span>
            </div>
        </div>

        <!-- Inline collapsible error/retry detail (M17) -->
        <div class="pc-collapse tile-detail" :class="{ 'pc-collapse--open': detailOpen }">
            <div>
                <div v-if="displayError" class="tile-detail-body">
                    <!-- Announce the failure once; the ticking countdown stays
                         outside the live region so it isn't read every second. -->
                    <div role="alert">
                        <p class="tile-error-title">{{ displayError.title || $t('connectionTile.connectionError') }}</p>
                        <p v-if="displayError.suggestion" class="tile-error-suggestion">{{ displayError.suggestion }}</p>
                        <span class="pc-chip-mono">{{ displayError.code || $t('connectionTile.unknown') }}</span>
                    </div>
                    <div class="tile-error-actions">
                        <span v-if="showCountdown" class="tile-countdown">
                            {{ $t('connectionTile.retryCountdown', { attempt: attemptNumber }) }}
                            <span class="tile-countdown-digits pc-num">
                                <Transition name="tile-tick" mode="out-in">
                                    <span :key="countdownSeconds">{{ countdownSeconds }}</span>
                                </Transition> </span
                            ><span class="pc-num">s</span>
                        </span>
                        <button type="button" class="pc-btn pc-btn--ghost-danger pc-btn--sm" :disabled="store.loading" @click="retryNow">{{ $t('common.retryNow') }}</button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Reconnect stays reachable when down without an error object (F5) -->
        <div v-if="!healthy && !errorInfo && !store.loading" class="tile-action">
            <button type="button" class="pc-btn pc-btn--secondary pc-btn--sm" @click="retryNow">
                {{ isPlex ? $t('common.reconnect') : $t('common.connect') }}
            </button>
        </div>
    </article>
</template>

<style scoped>
/* Raised tile with the 2px brand identity keyline on top (§5.2) — the only
   non-neutral borders in the app. Error adds a 2px danger left keyline. */
.tile {
    background: var(--pc-raised);
    border-radius: var(--pc-radius-md);
    border-top: 2px solid transparent;
    border-left: 2px solid transparent;
    padding: 12px 16px;
    transition: border-color var(--pc-dur-2) var(--pc-ease-out);
}
.tile--plex {
    border-top-color: var(--pc-plex);
}
.tile--discord {
    border-top-color: var(--pc-blurple);
}
.tile--error {
    border-left-color: var(--pc-danger);
}

/* ---- Header row ---- */
.tile-header {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 28px;
}
.tile-glyph {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    flex: none;
}
.tile-glyph--plex {
    color: var(--pc-plex);
}
.tile-glyph--discord {
    color: var(--pc-blurple);
}
.tile-glyph .pi {
    font-size: 15px;
}
.tile-name {
    margin: 0;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-text);
}
.tile-status {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
}
.tile-status .pc-dot {
    transition: color var(--pc-dur-2) var(--pc-ease-out); /* M6 dot crossfade */
}
.tile-status--success .pc-status-label {
    color: var(--pc-success);
}
.tile-status--warn .pc-status-label {
    color: var(--pc-warn);
}
.tile-status--danger .pc-status-label {
    color: var(--pc-danger);
}
.tile-status--idle .pc-status-label {
    color: var(--pc-text-muted);
}
.tile-expand {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    margin: -4px -8px -4px 0;
    border: none;
    border-radius: var(--pc-radius-sm);
    background: transparent;
    color: var(--pc-text-secondary);
    cursor: pointer;
    transition:
        background-color var(--pc-dur-1) var(--pc-ease-out),
        color var(--pc-dur-1) var(--pc-ease-out);
}
.tile-expand:hover {
    background: var(--pc-surface-700);
    color: var(--pc-text);
}
:root:not(.dark) .tile-expand:hover {
    background: var(--pc-surface-200);
}
.tile-expand .pi {
    font-size: 11px;
}

/* ---- 28px key-value caption rows ---- */
.tile-rows {
    margin-top: 6px;
}
.tile-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    min-height: 28px;
}
.tile-row-label {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
    flex: none;
}
.tile-row-value {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.tile-row-chip {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

/* ---- Error detail region ---- */
.tile-detail-body {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--pc-border-subtle);
}
.tile-error-title {
    margin: 0 0 2px;
    font-size: var(--pc-text-body);
    font-weight: 600;
    color: var(--pc-text);
}
.tile-error-suggestion {
    margin: 0 0 8px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.tile-error-actions {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 10px;
}
.tile-countdown {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.tile-countdown-digits {
    display: inline-block;
    font-family: var(--pc-font-mono);
    color: var(--pc-text-secondary);
    min-width: 2ch;
    text-align: right;
}
.tile-countdown-digits > span {
    display: inline-block;
}
/* M18 — per tick the digit slides up 4px + fades over 120ms */
.tile-tick-enter-active,
.tile-tick-leave-active {
    transition:
        opacity 120ms var(--pc-ease-out),
        transform 120ms var(--pc-ease-out);
}
.tile-tick-enter-from {
    opacity: 0;
    transform: translateY(4px);
}
.tile-tick-leave-to {
    opacity: 0;
    transform: translateY(-4px);
}

.tile-action {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--pc-border-subtle);
    display: flex;
    justify-content: flex-end;
}
</style>
