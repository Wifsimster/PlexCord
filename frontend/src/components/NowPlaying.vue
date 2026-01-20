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
const isPlaying = computed(() => playbackStore.isPlaying);
const isPaused = computed(() => playbackStore.isPaused);
const formattedPosition = computed(() => playbackStore.formattedPosition);
const formattedDuration = computed(() => playbackStore.formattedDuration);
const progressPercent = computed(() => playbackStore.progressPercent);

// Placeholder image for missing artwork
const placeholderImage = 'data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSI2NCIgaGVpZ2h0PSI2NCIgdmlld0JveD0iMCAwIDY0IDY0Ij48cmVjdCB3aWR0aD0iNjQiIGhlaWdodD0iNjQiIGZpbGw9IiM0YjU1NjMiLz48dGV4dCB4PSIzMiIgeT0iMzYiIGZvbnQtc2l6ZT0iMjQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiM5Y2EzYWYiPjwvdGV4dD48L3N2Zz4=';

const artworkUrl = computed(() => {
    if (currentTrack.value?.thumbUrl) {
        return currentTrack.value.thumbUrl;
    }
    return placeholderImage;
});

const playbackStateIcon = computed(() => {
    if (isPlaying.value) return 'pi pi-play-circle';
    if (isPaused.value) return 'pi pi-pause-circle';
    return 'pi pi-stop-circle';
});

const playbackStateText = computed(() => {
    if (isPlaying.value) return 'Playing';
    if (isPaused.value) return 'Paused';
    return 'Stopped';
});
</script>

<template>
    <div class="card">
        <div class="flex items-center justify-between mb-4">
            <span class="text-xl font-semibold">Now Playing</span>
            <span v-if="hasActiveSession" class="flex items-center gap-2 text-sm">
                <i :class="[playbackStateIcon, 'text-lg', { 'text-green-500': isPlaying, 'text-yellow-500': isPaused }]"></i>
                <span class="text-muted-color">{{ playbackStateText }}</span>
            </span>
        </div>

        <!-- No active session -->
        <div v-if="!hasActiveSession" class="flex flex-col items-center justify-center py-8 text-center">
            <i class="pi pi-volume-off text-4xl text-muted-color mb-4"></i>
            <span class="text-muted-color">No music playing</span>
            <span class="text-sm text-muted-color mt-2">Start playing music on Plex to see it here</span>
        </div>

        <!-- Active session -->
        <div v-else class="flex gap-4">
            <!-- Album artwork -->
            <div class="flex-shrink-0">
                <img
                    :src="artworkUrl"
                    :alt="currentTrack?.album || 'Album artwork'"
                    class="rounded-border shadow-md"
                    style="width: 80px; height: 80px; object-fit: cover;"
                    @error="(e) => e.target.src = placeholderImage"
                />
            </div>

            <!-- Track info -->
            <div class="flex-grow min-w-0">
                <!-- Track title -->
                <div class="font-semibold text-lg truncate text-surface-900 dark:text-surface-0">
                    {{ currentTrack?.track || 'Unknown Track' }}
                </div>

                <!-- Artist -->
                <div class="text-muted-color truncate">
                    {{ currentTrack?.artist || 'Unknown Artist' }}
                </div>

                <!-- Album -->
                <div class="text-sm text-muted-color truncate">
                    {{ currentTrack?.album || 'Unknown Album' }}
                </div>

                <!-- Progress bar and time -->
                <div class="mt-3">
                    <div class="flex justify-between text-xs text-muted-color mb-1">
                        <span>{{ formattedPosition }}</span>
                        <span>{{ formattedDuration }}</span>
                    </div>
                    <div class="w-full bg-surface-200 dark:bg-surface-700 rounded-full h-1.5">
                        <div
                            class="bg-primary h-1.5 rounded-full transition-all duration-300"
                            :style="{ width: `${progressPercent}%` }"
                        ></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Player info -->
        <div v-if="hasActiveSession && currentTrack?.playerName" class="mt-4 pt-4 border-t border-surface-200 dark:border-surface-700">
            <span class="text-xs text-muted-color">
                <i class="pi pi-desktop mr-1"></i>
                Playing on {{ currentTrack.playerName }}
            </span>
        </div>
    </div>
</template>
