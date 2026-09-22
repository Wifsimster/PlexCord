<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import BrandSymbol from '@/components/BrandSymbol.vue';
import ConnectionTile from '@/components/ConnectionTile.vue';
import DiscordSpecimen from '@/components/DiscordSpecimen.vue';
import TruncatedText from '@/components/TruncatedText.vue';
import { usePlayback } from '@/composables/usePlayback';
import { usePresenceStatus } from '@/composables/usePresenceStatus';
import { usePresenceStore } from '@/stores/presence';
import { mediaTypeOf } from '@/utils/presenceFormat';
import { GetPlexConnectionStatus, GetPlexToken, GetPollingInterval, GetPresenceFormat, GetServers } from '../../../wailsjs/go/main/App';

/**
 * Dashboard (spec §5.2) — the stage. The media on air is the hero, washed in
 * its own artwork, with the tally lamp saying whether Discord shows it; below,
 * the Discord specimen (the output) sits beside the two connection tiles (the
 * route). A failure appears in exactly one place (its connection tile + the
 * topbar tally). No manual refresh (F7).
 */

// Playback event lifecycle initialized once here via the refcounted
// composable (F35); the shell holds its own subscription for the headline.
const { t } = useI18n();
const { currentTrack, isPlaying, isPaused, hasActiveSession, formattedPosition, formattedDuration, progressPercent } = usePlayback();
const { tally } = usePresenceStatus();
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

const resumePresence = () => {
    if (presenceStore.paused) presenceStore.toggle();
};

// ---- The stage ---------------------------------------------------------------
const isVideo = computed(() => mediaTypeOf(currentTrack.value) !== 'music');
const stageTitle = computed(() => currentTrack.value?.title ?? currentTrack.value?.track ?? '');

// One line under the title, shaped by what is playing:
//   music → artist — album · movie → year · episode → show · S1 · E1
const stageSubtitle = computed(() => {
    const track = currentTrack.value;
    if (!track) return '';
    const type = mediaTypeOf(track);
    if (type === 'tv') {
        const episode = track.season && track.episode ? t('stage.episode', { season: track.season, episode: track.episode }) : '';
        return [track.showTitle, episode].filter(Boolean).join('  ·  ');
    }
    if (type === 'movie') return track.year ? String(track.year) : '';
    return [track.artist, track.album].filter(Boolean).join(' — ');
});

const stageEyebrow = computed(() => (currentTrack.value?.playerName ? t('stage.nowPlayingOn', { player: currentTrack.value.playerName }) : t('stage.nowPlaying')));

// The wash breathes while on air, freezes otherwise (M22).
const ambientPaused = computed(() => presenceStore.paused || !isPlaying.value);

// ---- Captions ----------------------------------------------------------------
const modKey = /mac/i.test(navigator.platform || navigator.userAgent) ? '⌘' : 'Ctrl';
</script>

<template>
    <div class="dashboard">
        <!-- ---- The stage (§5.2): what is on air, in its own colors ---- -->
        <section class="stage pc-panel-enter" :class="{ 'stage--off': presenceStore.paused, 'stage--idle': ready && !hasActiveSession }" :aria-label="$t('stage.aria')">
            <Transition name="pc-fade-slow">
                <img v-if="ready && currentTrack?.thumbUrl" :key="currentTrack.thumbUrl" :src="currentTrack.thumbUrl" class="stage-wash" :class="{ 'stage-wash--paused': ambientPaused }" alt="" aria-hidden="true" />
            </Transition>
            <div class="stage-scrim" aria-hidden="true"></div>

            <!-- Loading (M20) -->
            <div v-if="!ready" class="stage-body" aria-hidden="true">
                <div class="pc-skeleton stage-art"></div>
                <div class="stage-meta">
                    <div class="pc-skeleton" style="width: 140px; height: 26px; border-radius: 999px"></div>
                    <div class="pc-skeleton" style="width: 60%; height: 36px; margin-top: 16px"></div>
                    <div class="pc-skeleton" style="width: 40%; height: 16px; margin-top: 12px"></div>
                </div>
            </div>

            <!-- Idle: dead air -->
            <div v-else-if="!hasActiveSession" class="stage-body">
                <div class="stage-art stage-art--ghost" aria-hidden="true">
                    <BrandSymbol :size="56" variant="small" unlit />
                </div>
                <div class="stage-meta">
                    <span class="pc-tally pc-tally--lg" :class="`pc-tally--${tally.kind}`"><span class="pc-tally-lamp" aria-hidden="true"></span>{{ tally.label }}</span>
                    <h1 class="stage-title">{{ $t('stage.idleTitle') }}</h1>
                    <p class="stage-caption">{{ $t('dashboard.idleSub', { seconds: pollingInterval }) }}</p>
                </div>
            </div>

            <!-- On air / held / off air -->
            <div v-else class="stage-body">
                <div class="stage-art" :class="{ 'stage-art--poster': isVideo }">
                    <Transition name="pc-fade">
                        <img v-if="currentTrack.thumbUrl" :key="currentTrack.thumbUrl" :src="currentTrack.thumbUrl" alt="" />
                        <span v-else class="stage-art-glyph" aria-hidden="true">{{ isVideo ? '▶' : '♪' }}</span>
                    </Transition>
                </div>
                <div class="stage-meta">
                    <div class="stage-eyebrow">
                        <span class="pc-tally pc-tally--lg" :class="`pc-tally--${tally.kind}`"><span class="pc-tally-lamp" aria-hidden="true"></span>{{ tally.label }}</span>
                        <span v-if="tally.kind === 'on' || tally.kind === 'hold'" class="stage-player">{{ stageEyebrow }}</span>
                    </div>
                    <Transition name="pc-state" mode="out-in">
                        <div :key="currentTrack.sessionKey" class="stage-lines">
                            <TruncatedText as="h1" class="stage-title" :text="stageTitle" />
                            <TruncatedText v-if="stageSubtitle" as="p" class="stage-subtitle" :text="stageSubtitle" />
                        </div>
                    </Transition>

                    <div v-if="presenceStore.paused" class="stage-offair">
                        <span class="stage-caption">{{ $t('stage.offAirCaption') }}</span>
                        <button type="button" class="pc-btn pc-btn--primary" @click="resumePresence">
                            <i class="pi pi-play" aria-hidden="true"></i>
                            {{ $t('stage.resume') }}
                        </button>
                    </div>
                    <div v-else class="stage-progress" :class="{ 'stage-progress--held': tally.kind !== 'on' }">
                        <span class="pc-num stage-time">{{ formattedPosition }}</span>
                        <span class="stage-track"><span class="stage-fill" :style="{ width: `${progressPercent}%` }"></span></span>
                        <span class="pc-num stage-time">{{ formattedDuration }}</span>
                    </div>
                </div>
            </div>
        </section>

        <!-- ---- The output: exactly what Discord shows ---- -->
        <section class="pc-panel profile-panel pc-panel-enter pc-panel-enter--2" :aria-label="$t('stage.onProfile')">
            <header class="panel-header">
                <h2 class="pc-eyebrow">{{ $t('stage.onProfile') }}</h2>
                <i class="pi pi-discord profile-glyph" aria-hidden="true"></i>
            </header>
            <DiscordSpecimen class="presence-specimen" :track="currentTrack" :formats="formats" :paused="presenceStore.paused" :loading="!ready" :idle-title="$t('dashboard.idleTitle')" caption="" />
        </section>

        <!-- ---- The route: the two connections ---- -->
        <section class="pc-panel connections-panel pc-panel-enter pc-panel-enter--3" :aria-label="$t('dashboard.connections')">
            <header class="panel-header">
                <h2 class="pc-eyebrow">{{ $t('dashboard.connections') }}</h2>
                <span class="poll-caption">{{ $t('dashboard.pollCaption', { seconds: pollingInterval, modKey }) }}</span>
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
        </section>
    </div>
</template>

<style scoped>
/* Content grid (§5.2): max 1200px centered. The stage spans the width; below
   it, the specimen (5fr) sits beside the connections (7fr). Single column
   under lg. The shell provides page padding/canvas. */
.dashboard {
    max-width: 1200px;
    margin: 0 auto;
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: var(--pc-space-panel-gap);
    align-items: stretch;
}
@media (min-width: 992px) {
    .dashboard {
        grid-template-columns: minmax(0, 5fr) minmax(0, 7fr);
    }
    .stage {
        grid-column: 1 / -1;
    }
}

.panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 14px;
}
.panel-header .pc-eyebrow {
    margin: 0;
}

/* ---- The stage ---- */
.stage {
    position: relative;
    overflow: hidden;
    isolation: isolate;
    border-radius: var(--pc-radius-lg);
    border: 1px solid var(--pc-border);
    background: var(--pc-panel);
    box-shadow: var(--pc-shadow-panel);
    min-height: 248px;
}
/* The artwork, blown up and blurred into a wash of its own colors. */
.stage-wash {
    position: absolute;
    inset: -30%;
    width: 160%;
    height: 160%;
    object-fit: cover;
    filter: blur(64px) saturate(1.5);
    opacity: var(--pc-backdrop-opacity);
    z-index: -2;
    pointer-events: none;
    animation: pc-breathe var(--pc-loop-breathe) ease-in-out infinite;
}
.stage-wash--paused {
    animation-play-state: paused;
}
.stage--off .stage-wash {
    filter: blur(64px) grayscale(1);
}
/* Keeps text contrast whatever the artwork is: heavier on the text side. */
.stage-scrim {
    position: absolute;
    inset: 0;
    z-index: -1;
    background: var(--pc-stage-scrim);
    pointer-events: none;
}

.stage-body {
    display: flex;
    align-items: center;
    gap: 32px;
    padding: 32px;
}
.stage-art {
    position: relative;
    flex: none;
    width: 184px;
    height: 184px;
    border-radius: var(--pc-radius-md);
    overflow: hidden;
    background: var(--pc-raised);
    box-shadow:
        0 18px 40px -12px rgba(0, 0, 0, 0.55),
        0 0 0 1px rgba(255, 255, 255, 0.06);
    transition: filter var(--pc-dur-3) var(--pc-ease-out);
}
.stage-art--poster {
    width: 132px;
    height: 198px;
}
.stage-art img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.stage-art-glyph {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    font-size: 48px;
    color: var(--pc-text-faint);
}
.stage--off .stage-art {
    filter: grayscale(0.9) brightness(0.8);
}
.stage-art--ghost {
    display: grid;
    place-items: center;
    background: transparent;
    border: 1.5px dashed var(--pc-border-strong);
    box-shadow: none;
}
.stage-art--ghost {
    --pc-symbol-ink: var(--pc-border-strong);
}

.stage-meta {
    min-width: 0;
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
}
.stage-eyebrow {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    max-width: 100%;
}
.stage-player {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.stage-lines {
    min-width: 0;
    max-width: 100%;
}
.stage-title {
    margin: 18px 0 0;
    max-width: 100%;
    font-family: var(--pc-font-display);
    font-size: var(--pc-text-hero);
    font-weight: 700;
    line-height: 1.08;
    letter-spacing: -0.035em;
    color: var(--pc-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.stage-subtitle {
    margin: 8px 0 0;
    font-size: 16px;
    color: var(--pc-text-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.stage-caption {
    margin: 10px 0 0;
    max-width: 440px;
    font-size: var(--pc-text-body);
    color: var(--pc-text-secondary);
}

.stage-progress {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    max-width: 560px;
    margin-top: 24px;
}
.stage-time {
    font-family: var(--pc-font-mono);
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
.stage-track {
    position: relative;
    flex: 1;
    height: 4px;
    border-radius: var(--pc-radius-full);
    background: color-mix(in srgb, var(--pc-text) 14%, transparent);
    overflow: hidden;
}
.stage-fill {
    position: absolute;
    inset: 0 auto 0 0;
    border-radius: inherit;
    background: var(--pc-tally);
    transition: width 300ms linear; /* M12 */
}
.stage-progress--held .stage-fill {
    background: var(--pc-text-muted);
}

.stage-offair {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 16px;
    margin-top: 20px;
}
.stage-offair .stage-caption {
    margin: 0;
}

@media (max-width: 720px) {
    .stage-body {
        flex-direction: column;
        align-items: flex-start;
        gap: 20px;
        padding: 24px;
    }
    .stage-art {
        width: 120px;
        height: 120px;
    }
    .stage-title {
        font-size: var(--pc-text-display);
    }
}

/* ---- Profile (specimen) panel ---- */
.profile-glyph {
    font-size: 14px;
    color: var(--pc-blurple);
}
.presence-specimen {
    max-width: 460px;
    margin: 0 auto;
}

/* ---- Connections panel ---- */
.tiles {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 12px;
}
@media (min-width: 768px) {
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
    font-weight: 600;
    color: var(--pc-text);
    white-space: nowrap;
    text-decoration: underline;
    text-underline-offset: 3px;
}

.poll-caption {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}
</style>
