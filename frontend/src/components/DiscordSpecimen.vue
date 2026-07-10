<script setup>
import { computed } from 'vue';
import TruncatedText from '@/components/TruncatedText.vue';
import { renderPresenceLines } from '@/utils/presenceFormat';

/**
 * <DiscordSpecimen> — pixel-faithful, theme-exempt Discord activity card
 * (spec §5.0.4). The single source of truth for "what does Discord show":
 * used by the Dashboard Presence panel, the Settings format editor and the
 * wizard Complete step. Pure renderer — no store lifecycle of its own (F35).
 *
 * Uses the --pc-discord-* specimen palette exclusively for the card itself
 * (it does not follow the app theme, per §1.2).
 */
const props = defineProps({
    /** Track/session object (MusicSession shape). null → idle ghost. */
    track: { type: Object, default: null },
    /**
     * Presence format strings: { details, state }. Also accepts the raw
     * GetPresenceFormat() shape ({ detailsFormat, stateFormat }). When both
     * are empty the backend's default rendering is reproduced.
     */
    formats: { type: Object, default: null },
    /** Presence paused (M13): grayscale card + PAUSED badge. */
    paused: { type: Boolean, default: false },
    /** Sample data indicator: shows a SAMPLE badge. */
    sample: { type: Boolean, default: false },
    /** Loading state: skeleton card (M20). */
    loading: { type: Boolean, default: false },
    /** Idle ghost caption (page-specific per spec §5.2/§5.4). */
    idleTitle: { type: String, default: 'Nothing playing on Plex' },
    /** Caption beneath the specimen well; empty string hides it. */
    caption: { type: String, default: 'Exactly what your Discord profile shows.' }
});

const normalizedFormats = computed(() => ({
    details: props.formats?.details ?? props.formats?.detailsFormat ?? '',
    state: props.formats?.state ?? props.formats?.stateFormat ?? ''
}));

const lines = computed(() => renderPresenceLines(normalizedFormats.value, props.track));

const albumLine = computed(() => props.track?.album ?? '');

const hasProgress = computed(() => (props.track?.duration ?? 0) > 0);

const progressPercent = computed(() => {
    if (!hasProgress.value) return 0;
    const percent = ((props.track.viewOffset ?? 0) / props.track.duration) * 100;
    return Math.min(100, Math.max(0, percent));
});

function formatTime(ms) {
    if (!ms || ms <= 0) return '0:00';
    const totalSeconds = Math.floor(ms / 1000);
    const minutes = Math.floor(totalSeconds / 60);
    const seconds = totalSeconds % 60;
    return `${minutes}:${seconds.toString().padStart(2, '0')}`;
}

const position = computed(() => formatTime(props.track?.viewOffset));
const duration = computed(() => formatTime(props.track?.duration));
</script>

<template>
    <div class="pc-specimen">
        <div class="pc-specimen-well">
            <!-- Ambient artwork backdrop slot (§4.1) — the Dashboard Presence
                 panel is the ONLY consumer allowed to fill this. -->
            <slot name="backdrop"></slot>

            <!-- Skeleton (M20) -->
            <div v-if="loading" class="specimen-card specimen-card--skeleton" aria-hidden="true">
                <div class="pc-skeleton skeleton-art"></div>
                <div class="skeleton-lines">
                    <div class="pc-skeleton skeleton-line" style="width: 70%"></div>
                    <div class="pc-skeleton skeleton-line" style="width: 55%"></div>
                    <div class="pc-skeleton skeleton-line" style="width: 45%"></div>
                </div>
            </div>

            <!-- Idle ghost -->
            <div v-else-if="!track" class="specimen-ghost">
                <span class="ghost-glyph" aria-hidden="true">–</span>
                <p class="ghost-title">{{ idleTitle }}</p>
                <slot name="idle-caption"></slot>
            </div>

            <!-- The specimen card -->
            <div v-else class="specimen-card" :class="{ 'specimen-card--paused': paused }">
                <div class="card-header">
                    <i class="pi pi-headphones card-header-glyph" aria-hidden="true"></i>
                    <span class="card-header-label">Listening to Plex</span>
                    <span class="card-header-badges">
                        <span v-if="sample" class="pc-badge">Sample</span>
                        <Transition name="pc-fade">
                            <span v-if="paused" class="pc-badge pc-badge--warn">Paused</span>
                        </Transition>
                    </span>
                </div>
                <div class="card-body">
                    <div class="card-art">
                        <Transition name="pc-fade">
                            <img v-if="track.thumbUrl" :key="track.thumbUrl" :src="track.thumbUrl" :alt="track.album ? `Album art for ${track.album}` : 'Album artwork'" class="card-art-img" />
                            <span v-else class="card-art-ghost" aria-hidden="true">♪</span>
                        </Transition>
                    </div>
                    <Transition name="specimen-swap" mode="out-in">
                        <div :key="track.sessionKey" class="card-lines">
                            <TruncatedText as="p" class="card-line card-line--details" :text="lines.details" />
                            <TruncatedText v-if="lines.state" as="p" class="card-line card-line--state" :text="lines.state" />
                            <TruncatedText v-if="albumLine" as="p" class="card-line card-line--album" :text="albumLine" />
                        </div>
                    </Transition>
                </div>
                <div v-if="hasProgress" class="card-progress">
                    <span class="card-time pc-num">{{ position }}</span>
                    <span class="card-progress-track">
                        <span class="card-progress-fill" :style="{ width: progressPercent + '%' }"></span>
                    </span>
                    <span class="card-time pc-num">{{ duration }}</span>
                </div>
            </div>
        </div>
        <p v-if="caption" class="specimen-caption">{{ caption }}</p>
    </div>
</template>

<style scoped>
/* ---- Well framing comes from .pc-specimen-well (components.css) ---- */
.specimen-caption {
    margin: 8px 0 0;
    text-align: center;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}

/* ---- Idle ghost (dashed frame, centered – glyph) ---- */
.specimen-ghost {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    width: 340px;
    max-width: 100%;
    margin: 0 auto;
    min-height: 120px;
    padding: 16px;
    border: 1px dashed var(--pc-border);
    border-radius: 8px;
    text-align: center;
}
.ghost-glyph {
    font-size: 20px;
    line-height: 1;
    color: var(--pc-text-faint);
}
.ghost-title {
    margin: 0;
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
}

/* ---- The theme-exempt Discord card (--pc-discord-* only) ---- */
.specimen-card {
    width: 340px;
    max-width: 100%;
    margin: 0 auto;
    padding: 12px;
    border-radius: 8px;
    background: var(--pc-discord-bg);
    font-family: var(--pc-discord-font);
    transition: filter var(--pc-dur-3) var(--pc-ease-out); /* M13 */
}
.specimen-card--paused {
    filter: grayscale(0.9) brightness(0.8);
}

.card-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 10px;
}
.card-header-glyph {
    font-size: 12px;
    color: var(--pc-discord-green);
}
.card-header-label {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.02em;
    text-transform: uppercase;
    color: var(--pc-discord-muted);
}
.card-header-badges {
    margin-left: auto;
    display: inline-flex;
    gap: 4px;
}

.card-body {
    display: flex;
    gap: 10px;
    align-items: center;
}
.card-art {
    position: relative;
    width: 64px;
    height: 64px;
    flex: none;
    border-radius: 6px;
    overflow: hidden;
    background: var(--pc-discord-raised);
    display: flex;
    align-items: center;
    justify-content: center;
}
.card-art-img {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
}
.card-art-ghost {
    font-size: 22px;
    line-height: 1;
    color: var(--pc-blurple);
}

.card-lines {
    min-width: 0;
    flex: 1;
}
.card-line {
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.card-line--details {
    font-size: 14px;
    font-weight: 600;
    color: var(--pc-discord-text);
}
.card-line--state,
.card-line--album {
    font-size: 13px;
    color: var(--pc-discord-muted);
}

.card-progress {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
}
.card-progress-track {
    flex: 1;
    height: 4px;
    border-radius: 2px;
    background: var(--pc-discord-raised);
    overflow: hidden;
}
.card-progress-fill {
    display: block;
    height: 100%;
    border-radius: 2px;
    background: var(--pc-discord-text);
    transition: width 300ms linear; /* M12 — matches poll cadence */
}
.card-time {
    font-family: var(--pc-font-mono);
    font-size: 11px;
    color: var(--pc-discord-muted);
}

/* ---- Skeleton (M20) ---- */
.specimen-card--skeleton {
    display: flex;
    gap: 10px;
    align-items: center;
}
.skeleton-art {
    width: 64px;
    height: 64px;
    flex: none;
    border-radius: 6px;
}
.skeleton-lines {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.skeleton-line {
    height: 12px;
}

/* ---- M10 — track change: text lines slide up 6px + fade, staggered ---- */
.specimen-swap-enter-active .card-line {
    transition:
        opacity 240ms var(--pc-ease-out),
        transform 240ms var(--pc-ease-out);
}
.specimen-swap-enter-active .card-line--state {
    transition-delay: 30ms;
}
.specimen-swap-enter-active .card-line--album {
    transition-delay: 60ms;
}
.specimen-swap-enter-from .card-line {
    opacity: 0;
    transform: translateY(6px);
}
.specimen-swap-leave-active {
    transition: opacity 160ms var(--pc-ease-in);
}
.specimen-swap-leave-to {
    opacity: 0;
}
</style>
