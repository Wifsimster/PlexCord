<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import ConnectionTile from '@/components/ConnectionTile.vue';
import DiscordSpecimen from '@/components/DiscordSpecimen.vue';
import TruncatedText from '@/components/TruncatedText.vue';
import { usePlayback } from '@/composables/usePlayback';
import { usePresenceStore } from '@/stores/presence';
import { renderPresenceLines } from '@/utils/presenceFormat';
import { GetPlexConnectionStatus, GetPlexToken, GetPollingInterval, GetPresenceFormat, GetServers } from '../../../wailsjs/go/main/App';

/**
 * Dashboard (spec §5.2) — the signal path expanded. One glance answers
 * live / showing / broken; paused is honest everywhere; a failure appears
 * in exactly one place (its connection tile + topbar node). No page h1 —
 * the topbar headline is the header (F8). No manual refresh (F7).
 */

// Playback event lifecycle initialized once here via the refcounted
// composable (F35); the shell holds its own subscription for the headline.
const { t } = useI18n();
const { currentTrack, isPlaying, isPaused, hasActiveSession, formattedPosition, formattedDuration } = usePlayback();
const presenceStore = usePresenceStore();

// ---- Loading (M20 skeleton, minimum 400ms to avoid flash) ------------------
const ready = ref(false);
let readyTimer = null;

// ---- Live settings the panel narrates ---------------------------------------
const formats = ref(null); // { detailsFormat, stateFormat }
const pollingInterval = ref(5);

// ---- Setup-incomplete resume tile (F22) -------------------------------------
// The backend cannot report "skipped" directly (CheckSetupComplete() is true
// for skipped setups so the router even allows this page) — derive the first
// unfinished wizard step from the persisted configuration instead.
const resumeTarget = ref('');

const deriveSetupResume = async () => {
    try {
        const token = await GetPlexToken();
        if (!token) {
            resumeTarget.value = '/setup/plex';
            return;
        }
        const status = await GetPlexConnectionStatus();
        let hasServer = !!status?.serverUrl;
        if (!hasServer) {
            const servers = await GetServers();
            hasServer = Array.isArray(servers) && servers.length > 0;
        }
        if (!hasServer) {
            resumeTarget.value = '/setup/plex';
            return;
        }
        resumeTarget.value = status?.userId ? '' : '/setup/user';
    } catch (error) {
        console.error('Failed to derive setup progress:', error);
        resumeTarget.value = '';
    }
};

onMounted(async () => {
    const start = performance.now();
    try {
        const [presenceFormats, interval] = await Promise.all([GetPresenceFormat(), GetPollingInterval()]);
        formats.value = presenceFormats;
        if (interval > 0) pollingInterval.value = interval;
    } catch (error) {
        console.error('Failed to load presence settings:', error);
    }
    deriveSetupResume();
    const remaining = Math.max(0, 400 - (performance.now() - start));
    readyTimer = setTimeout(() => {
        ready.value = true;
    }, remaining);
});

onBeforeUnmount(() => {
    if (readyTimer) clearTimeout(readyTimer);
});

// ---- Presence panel header chip ---------------------------------------------
const chip = computed(() => {
    if (presenceStore.paused) return { kind: 'paused-presence', label: t('dashboard.chipPausedByYou'), severity: 'warn' };
    if (isPlaying.value) return { kind: 'live', label: t('dashboard.chipLive'), severity: 'success' };
    if (isPaused.value) return { kind: 'paused-track', label: t('dashboard.chipPaused'), severity: 'warn' };
    return { kind: 'idle', label: t('dashboard.chipIdle'), severity: 'muted' };
});

const resumePresence = () => {
    if (presenceStore.paused) presenceStore.toggle();
};

// ---- Fact strip: what the relay is literally transmitting --------------------
const lines = computed(() =>
    renderPresenceLines(
        {
            details: formats.value?.detailsFormat ?? '',
            state: formats.value?.stateFormat ?? ''
        },
        currentTrack.value
    )
);

const facts = computed(() => [
    { label: t('dashboard.factDetails'), value: lines.value.details || '—' },
    { label: t('dashboard.factState'), value: lines.value.state || '—' },
    { label: t('dashboard.factPlayer'), value: currentTrack.value?.playerName || '—' },
    { label: t('dashboard.factSession'), value: `${formattedPosition.value} / ${formattedDuration.value}` }
]);

// ---- Ambient artwork backdrop (§4.1) ----------------------------------------
// Breathing while live, frozen while paused (either kind), gone when idle.
const ambientPaused = computed(() => presenceStore.paused || !isPlaying.value);

// ---- Captions ----------------------------------------------------------------
const modKey = /mac/i.test(navigator.platform || navigator.userAgent) ? '⌘' : 'Ctrl';
const specimenCaption = computed(() => (ready.value && !hasActiveSession.value ? '' : t('dashboard.specimenCaption')));
</script>

<template>
    <div class="dashboard">
        <!-- ---- Presence panel (§5.2 left) ---- -->
        <section class="pc-panel presence-panel pc-panel-enter" :aria-label="$t('dashboard.presence')">
            <header class="panel-header">
                <h2 class="pc-eyebrow">{{ $t('dashboard.presence') }}</h2>
                <Transition name="pc-state" mode="out-in">
                    <button v-if="chip.kind === 'paused-presence'" :key="chip.kind" type="button" class="pc-badge pc-badge--warn state-chip state-chip--button" :title="$t('dashboard.resumePresence')" @click="resumePresence">
                        <i class="pi pi-pause state-chip-glyph" aria-hidden="true"></i>
                        {{ chip.label }}
                    </button>
                    <span v-else :key="chip.kind" class="pc-badge state-chip" :class="{ 'pc-badge--success': chip.severity === 'success', 'pc-badge--warn': chip.severity === 'warn' }">
                        <span v-if="chip.kind === 'live'" class="pc-eq" aria-hidden="true"><i></i><i></i><i></i></span>
                        <i v-else-if="chip.kind === 'paused-track'" class="pi pi-pause state-chip-glyph" aria-hidden="true"></i>
                        <span v-else class="state-chip-glyph" aria-hidden="true">–</span>
                        {{ chip.label }}
                    </span>
                </Transition>
            </header>

            <DiscordSpecimen class="presence-specimen" :track="currentTrack" :formats="formats" :paused="presenceStore.paused" :loading="!ready" :idle-title="$t('dashboard.idleTitle')" :caption="specimenCaption">
                <template #backdrop>
                    <!-- §4.1 ambient artwork backdrop — Dashboard-only; keyed img
                         + non-out-in fade = M10 crossfade on track change,
                         320ms unmount fade when idle (M22). -->
                    <Transition name="pc-fade-slow">
                        <img v-if="ready && currentTrack?.thumbUrl" :key="currentTrack.thumbUrl" :src="currentTrack.thumbUrl" class="pc-ambient" :class="{ 'pc-ambient--paused': ambientPaused }" alt="" aria-hidden="true" />
                    </Transition>
                </template>
                <template #idle-caption>
                    <p class="idle-sub">{{ $t('dashboard.idleSub', { seconds: pollingInterval }) }}</p>
                </template>
            </DiscordSpecimen>

            <!-- Mono fact strip: the relay's literal transmission -->
            <dl v-if="ready && hasActiveSession" class="fact-strip" :class="{ 'fact-strip--dim': presenceStore.paused }">
                <div v-for="fact in facts" :key="fact.label" class="fact">
                    <dt class="fact-label">{{ fact.label }}</dt>
                    <TruncatedText as="dd" class="fact-value" :text="fact.value" />
                </div>
            </dl>
        </section>

        <!-- ---- Connections panel (§5.2 right) ---- -->
        <section class="pc-panel connections-panel pc-panel-enter pc-panel-enter--2" :aria-label="$t('dashboard.connections')">
            <header class="panel-header">
                <h2 class="pc-eyebrow">{{ $t('dashboard.connections') }}</h2>
            </header>

            <div class="tiles">
                <ConnectionTile source="plex" />
                <ConnectionTile source="discord" />

                <!-- Setup-skipped resume tile (F22) -->
                <router-link v-if="resumeTarget" :to="resumeTarget" class="resume-tile">
                    <span class="resume-text">{{ $t('dashboard.setupIncomplete') }}</span>
                    <span class="resume-link">{{ $t('dashboard.resumeSetup') }}</span>
                </router-link>
            </div>

            <p class="poll-caption">{{ $t('dashboard.pollCaption', { seconds: pollingInterval, modKey }) }}</p>
        </section>
    </div>
</template>

<style scoped>
/* Content grid (§5.2): max 1200px centered; ≥lg 7fr/5fr, below single
   column with Presence first. The shell provides page padding/canvas. */
.dashboard {
    max-width: 1200px;
    margin: 0 auto;
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: var(--pc-space-panel-gap);
    align-items: start;
}
@media (min-width: 992px) {
    .dashboard {
        grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
    }
}

.panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 12px;
}
.panel-header .pc-eyebrow {
    margin: 0;
}

/* ---- Presence panel ---- */
.presence-panel {
    padding: 24px; /* §5.2: 24px for the Presence panel */
}
.presence-specimen {
    max-width: 460px;
    margin: 0 auto;
}
.idle-sub {
    margin: 0;
    max-width: 300px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}

/* Header playback-state chip */
.state-chip {
    gap: 6px;
}
.state-chip-glyph {
    font-size: 10px;
    line-height: 1;
}
.state-chip--button {
    cursor: pointer;
}
.state-chip .pc-eq {
    height: 9px;
}

/* Mono fact strip: 2×2, 32px rows, 12.5px muted labels / 13px mono values */
.fact-strip {
    margin: 16px auto 0;
    max-width: 460px;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 24px;
    transition: opacity var(--pc-dur-2) var(--pc-ease-out);
}
.fact-strip--dim {
    opacity: 0.5;
}
.fact {
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 32px;
    border-top: 1px solid var(--pc-border-subtle);
}
.fact:nth-child(-n + 2) {
    border-top: none;
}
.fact-label {
    margin: 0;
    flex: none;
    width: 52px;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.fact-value {
    margin: 0;
    min-width: 0;
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-mono);
    color: var(--pc-text-secondary);
    font-variant-numeric: tabular-nums;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
@media (max-width: 560px) {
    .fact-strip {
        grid-template-columns: minmax(0, 1fr);
    }
    .fact:nth-child(2) {
        border-top: 1px solid var(--pc-border-subtle);
    }
}

/* ---- Connections panel ---- */
.tiles {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
}
/* Below lg the panels stack — let the two tiles sit side-by-side ≥ md */
@media (min-width: 768px) and (max-width: 991.98px) {
    .tiles {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

.resume-tile {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 44px;
    padding: 8px 16px;
    background: var(--pc-raised);
    border-radius: var(--pc-radius-md);
    text-decoration: none;
    transition: background-color var(--pc-dur-1) var(--pc-ease-out);
}
.dark .resume-tile:hover {
    background: var(--pc-surface-700);
}
:root:not(.dark) .resume-tile:hover {
    background: var(--pc-surface-200);
}
.resume-text {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
}
.resume-link {
    font-size: var(--pc-text-caption);
    font-weight: 500;
    color: var(--pc-accent);
    white-space: nowrap;
}
.resume-tile:hover .resume-link {
    color: var(--pc-accent-hover);
}

.poll-caption {
    margin: 12px 0 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
</style>
