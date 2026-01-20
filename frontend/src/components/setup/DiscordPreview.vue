<script setup>
import { computed, onMounted, onUnmounted } from 'vue';
import { usePlaybackStore } from '@/stores/playback';
import Badge from 'primevue/badge';

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
const isPlaying = computed(() => playbackStore.isPlaying);
const isPaused = computed(() => playbackStore.isPaused);
const formattedPosition = computed(() => playbackStore.formattedPosition);
const formattedDuration = computed(() => playbackStore.formattedDuration);

// Placeholder image for missing artwork (Discord-style music icon)
const placeholderImage = 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI2NCIgaGVpZ2h0PSI2NCIgdmlld0JveD0iMCAwIDY0IDY0Ij48cmVjdCB3aWR0aD0iNjQiIGhlaWdodD0iNjQiIGZpbGw9IiM1ODY1RjIiLz48dGV4dCB4PSIzMiIgeT0iNDAiIGZvbnQtc2l6ZT0iMjgiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiNmZmYiPuKZqjwvdGV4dD48L3N2Zz4=';

const artworkUrl = computed(() => {
    if (currentTrack.value?.thumbUrl) {
        return currentTrack.value.thumbUrl;
    }
    return placeholderImage;
});
</script>

<template>
    <div class="discord-preview-wrapper">
        <!-- Discord-style Rich Presence Card -->
        <div class="discord-presence-card">
            <!-- Header -->
            <div class="presence-header">
                <i class="pi pi-headphones"></i>
                <span>Listening to Plex</span>
            </div>

            <!-- Content when music is playing -->
            <div v-if="hasActiveSession" class="presence-content">
                <!-- Album artwork -->
                <div class="presence-artwork">
                    <img
                        :src="artworkUrl"
                        :alt="currentTrack?.album || 'Album artwork'"
                        @error="(e) => e.target.src = placeholderImage"
                    />
                </div>

                <!-- Track info -->
                <div class="presence-info">
                    <div class="presence-title">{{ currentTrack?.track || 'Unknown Track' }}</div>
                    <div class="presence-artist">by {{ currentTrack?.artist || 'Unknown Artist' }}</div>
                    <div class="presence-album">on {{ currentTrack?.album || 'Unknown Album' }}</div>
                    <div class="presence-time">
                        <span v-if="isPaused" class="paused-indicator">
                            <i class="pi pi-pause"></i> Paused
                        </span>
                        <span v-else>
                            <i class="pi pi-play"></i> {{ formattedPosition }} / {{ formattedDuration }}
                        </span>
                    </div>
                </div>
            </div>

            <!-- Empty state when no music playing -->
            <div v-else class="presence-empty">
                <div class="empty-icon">
                    <i class="pi pi-volume-off"></i>
                </div>
                <div class="empty-text">
                    <p class="font-semibold">No music playing</p>
                    <p class="text-sm text-muted-color">Start playing music on Plex to see the preview</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.discord-preview-wrapper {
    max-width: 400px;
    margin: 0 auto;
}

.preview-label {
    display: flex;
    align-items: center;
    margin-bottom: 1rem;
}

.discord-presence-card {
    background: #2f3136; /* Discord dark background */
    border-radius: 8px;
    overflow: hidden;
    color: #dcddde;
    font-family: 'gg sans', 'Noto Sans', 'Helvetica Neue', Helvetica, Arial, sans-serif;
}

.presence-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: #202225;
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.02em;
    color: #b9bbbe;
}

.presence-header i {
    color: #3ba55c; /* Discord green */
}

.presence-content {
    display: flex;
    gap: 1rem;
    padding: 1rem;
}

.presence-artwork {
    flex-shrink: 0;
}

.presence-artwork img {
    width: 80px;
    height: 80px;
    border-radius: 8px;
    object-fit: cover;
}

.presence-info {
    display: flex;
    flex-direction: column;
    justify-content: center;
    min-width: 0;
    flex: 1;
}

.presence-title {
    font-weight: 600;
    font-size: 0.9rem;
    color: #ffffff;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-bottom: 0.25rem;
}

.presence-artist,
.presence-album {
    font-size: 0.8rem;
    color: #b9bbbe;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.presence-time {
    margin-top: 0.5rem;
    font-size: 0.75rem;
    color: #72767d;
    display: flex;
    align-items: center;
    gap: 0.25rem;
}

.presence-time i {
    font-size: 0.65rem;
}

.paused-indicator {
    color: #faa61a; /* Discord yellow */
}

.presence-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
    text-align: center;
}

.empty-icon {
    font-size: 2.5rem;
    color: #72767d;
    margin-bottom: 1rem;
}

.empty-text p {
    margin: 0;
    color: #b9bbbe;
}

.empty-text .text-sm {
    color: #72767d;
    margin-top: 0.25rem;
}

/* Dark mode already matches Discord theme */
</style>
