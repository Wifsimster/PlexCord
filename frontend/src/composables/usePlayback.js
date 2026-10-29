import { onMounted, onUnmounted } from 'vue';
import { storeToRefs } from 'pinia';
import { usePlaybackStore } from '@/stores/playback';

let subscribers = 0;

export function usePlayback() {
    const store = usePlaybackStore();

    onMounted(() => {
        subscribers += 1;
        store.initializeEventListeners();
    });

    onUnmounted(() => {
        subscribers = Math.max(0, subscribers - 1);
        if (subscribers === 0) {
            store.cleanupEventListeners();
        }
    });

    const { currentTrack, isPlaying, isPaused, isStopped, hasActiveSession, formattedPosition, formattedDuration, progressPercent, playbackState } = storeToRefs(store);

    return {
        store,
        currentTrack,
        isPlaying,
        isPaused,
        isStopped,
        hasActiveSession,
        formattedPosition,
        formattedDuration,
        progressPercent,
        playbackState
    };
}
