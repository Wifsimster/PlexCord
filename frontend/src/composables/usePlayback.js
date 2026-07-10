import { onMounted, onUnmounted } from 'vue';
import { storeToRefs } from 'pinia';
import { usePlaybackStore } from '@/stores/playback';

// Reference count so multiple consumers (shell headline + Dashboard panel +
// Settings specimen) can share the single set of Wails playback listeners.
// The store's own initialize/cleanup are unconditional; this wrapper makes
// them safe to use from more than one component at once (spec §5.0.6 / F35).
let subscribers = 0;

/**
 * Playback event lifecycle + reactive playback state.
 *
 * Call once per component that needs live playback data. Listeners are
 * registered on first mount and removed only when the last subscribed
 * component unmounts.
 *
 * @returns {{
 *   store: ReturnType<typeof usePlaybackStore>,
 *   currentTrack: import('vue').Ref<Object|null>,
 *   isPlaying: import('vue').Ref<boolean>,
 *   isPaused: import('vue').Ref<boolean>,
 *   isStopped: import('vue').Ref<boolean>,
 *   hasActiveSession: import('vue').ComputedRef<boolean>,
 *   formattedPosition: import('vue').ComputedRef<string>,
 *   formattedDuration: import('vue').ComputedRef<string>,
 *   progressPercent: import('vue').ComputedRef<number>,
 *   playbackState: import('vue').ComputedRef<string>
 * }}
 */
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
