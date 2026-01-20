<script setup>
import { computed, onMounted, onUnmounted } from 'vue';
import { usePlaybackStore } from '@/stores/playback';

const playbackStore = usePlaybackStore();

// Initialize event listeners when component mounts
onMounted(() => {
    playbackStore.initializeEventListeners();
});

// Clean up when component unmounts
onUnmounted(() => {
    playbackStore.cleanupEventListeners();
});

// Computed properties for template
const hasActiveSession = computed(() => playbackStore.hasActiveSession);
const currentTrack = computed(() => playbackStore.currentTrack);
const isPaused = computed(() => playbackStore.isPaused);
const formattedPosition = computed(() => playbackStore.formattedPosition);
const formattedDuration = computed(() => playbackStore.formattedDuration);

// Placeholder image for missing artwork (Discord-style music icon)
const placeholderImage =
    'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI2NCIgaGVpZ2h0PSI2NCIgdmlld0JveD0iMCAwIDY0IDY0Ij48cmVjdCB3aWR0aD0iNjQiIGhlaWdodD0iNjQiIGZpbGw9IiM1ODY1RjIiLz48dGV4dCB4PSIzMiIgeT0iNDAiIGZvbnQtc2l6ZT0iMjgiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiNmZmYiPuKZqjwvdGV4dD48L3N2Zz4=';

const artworkUrl = computed(() => {
    if (currentTrack.value?.thumbUrl) {
        return currentTrack.value.thumbUrl;
    }
    return placeholderImage;
});
</script>

<template>
    <div class="discord-preview-container">
        <!-- Discord-style Rich Presence Card -->
        <div class="discord-card">
            <!-- Header -->
            <div class="discord-header">
                <i class="pi pi-headphones text-[#3ba55c] mr-2"></i>
                <span>Listening to Plex</span>
            </div>

            <!-- Content when music is playing -->
            <div v-if="hasActiveSession" class="discord-content">
                <!-- Album artwork -->
                <div class="discord-artwork">
                    <img :src="artworkUrl" :alt="currentTrack?.album || 'Album artwork'" @error="(e) => (e.target.src = placeholderImage)" />
                </div>

                <!-- Track info -->
                <div class="discord-info">
                    <div class="track-title" :title="currentTrack?.track">
                        {{ currentTrack?.track || 'Unknown Track' }}
                    </div>
                    <div class="track-artist" :title="currentTrack?.artist">by {{ currentTrack?.artist || 'Unknown Artist' }}</div>
                    <div class="track-album" :title="currentTrack?.album">on {{ currentTrack?.album || 'Unknown Album' }}</div>
                    <div class="track-progress">
                        <span v-if="isPaused" class="paused-state"> <i class="pi pi-pause text-[10px] mr-1"></i> Paused </span>
                        <span v-else class="playing-state"> <i class="pi pi-play text-[10px] mr-1"></i> {{ formattedPosition }} / {{ formattedDuration }} </span>
                    </div>
                </div>
            </div>

            <!-- Empty state when no music playing -->
            <div v-else class="discord-empty">
                <div class="empty-icon">
                    <i class="pi pi-volume-off"></i>
                </div>
                <div class="empty-text">
                    <p class="empty-title">No music playing</p>
                    <p class="empty-subtitle">Start playing music on Plex to see the preview</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.discord-preview-container {
    width: 100%;
    max-width: 400px;
    margin: 0 auto;
}

.discord-card {
    background-color: #2f3136;
    border-radius: 8px;
    overflow: hidden;
    font-family: 'gg sans', 'Noto Sans', 'Helvetica Neue', Helvetica, Arial, sans-serif;
    color: #dcddde;
    box-shadow:
        0 4px 6px -1px rgba(0, 0, 0, 0.1),
        0 2px 4px -1px rgba(0, 0, 0, 0.06);
}

.discord-header {
    background-color: #202225;
    padding: 12px 16px;
    font-size: 12px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    color: #b9bbbe;
    display: flex;
    align-items: center;
}

.discord-content {
    display: flex;
    gap: 16px;
    padding: 16px;
}

.discord-artwork {
    flex-shrink: 0;
}

.discord-artwork img {
    width: 80px;
    height: 80px;
    border-radius: 8px;
    object-fit: cover;
    background-color: #202225;
}

.discord-info {
    display: flex;
    flex-direction: column;
    justify-content: center;
    min-width: 0;
    flex: 1;
    gap: 2px;
}

.track-title {
    font-weight: 600;
    font-size: 14px;
    color: #ffffff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}

.track-title:hover {
    text-decoration: underline;
}

.track-artist,
.track-album {
    font-size: 12px;
    color: #b9bbbe;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}

.track-artist:hover,
.track-album:hover {
    text-decoration: underline;
}

.track-progress {
    margin-top: 4px;
    font-size: 12px;
    color: #b9bbbe;
    display: flex;
    align-items: center;
}

.paused-state {
    color: #faa61a;
    font-weight: 500;
    display: flex;
    align-items: center;
}

.playing-state {
    display: flex;
    align-items: center;
}

.discord-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 32px 16px;
    text-align: center;
}

.empty-icon {
    width: 64px;
    height: 64px;
    background-color: #36393f;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 12px;
}

.empty-icon i {
    font-size: 24px;
    color: #72767d;
}

.empty-text {
    color: #b9bbbe;
}

.empty-title {
    font-weight: 600;
    margin-bottom: 4px;
}

.empty-subtitle {
    font-size: 12px;
    color: #72767d;
}
</style>
