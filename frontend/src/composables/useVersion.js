import { ref, computed } from 'vue';
import { GetVersion } from '../../wailsjs/go/main/App';

// Module-level cache: GetVersion() is called at most once per app run
// (spec §5.0.6 — replaces the three duplicate calls in Footer/Dashboard/Settings).
const versionInfo = ref(null);
let fetchPromise = null;

/**
 * Cached app version info.
 *
 * @returns {{
 *   versionInfo: import('vue').Ref<Object|null>,
 *   version: import('vue').ComputedRef<string>,
 *   commit: import('vue').ComputedRef<string>,   // short (7 chars), '' when unknown
 *   buildDate: import('vue').ComputedRef<string>,
 *   display: import('vue').ComputedRef<string>   // e.g. 'PlexCord v4.3.0 · a1b2c3d'
 * }}
 */
export function useVersion() {
    if (!fetchPromise) {
        fetchPromise = GetVersion()
            .then((info) => {
                versionInfo.value = info;
            })
            .catch((error) => {
                console.error('Failed to get version:', error);
            });
    }

    const version = computed(() => {
        const v = versionInfo.value?.version ?? '';
        return v.replace(/^v/, '');
    });

    const commit = computed(() => {
        const c = versionInfo.value?.commit ?? '';
        if (!c || c === 'unknown' || c === 'none') return '';
        return c.slice(0, 7);
    });

    const buildDate = computed(() => versionInfo.value?.buildDate ?? '');

    const display = computed(() => {
        if (!version.value) return 'PlexCord';
        return commit.value ? `PlexCord v${version.value} · ${commit.value}` : `PlexCord v${version.value}`;
    });

    return { versionInfo, version, commit, buildDate, display };
}
