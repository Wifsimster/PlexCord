import { ref, computed } from 'vue';
import { GetVersion } from '../../wailsjs/go/main/App';

const versionInfo = ref(null);
let fetchPromise = null;

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
