<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useToast } from 'primevue/usetoast';
import Toast from 'primevue/toast';
import AppFooter from './AppFooter.vue';
import AppTopbar from './AppTopbar.vue';
import { usePlayback } from '@/composables/usePlayback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';
import { useUpdatesStore } from '@/stores/updates';
import { OpenReleaseURL, OpenReleasesPage } from '../../wailsjs/go/main/App';

const router = useRouter();
const toast = useToast();

const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();
const presenceStore = usePresenceStore();
const updatesStore = useUpdatesStore();

// ---- Update notifications (auto-updater) ----------------------------------
// The backend checks for updates in the background (startup + periodic) and
// downloads them silently where self-update is supported. Surface the result
// as a sticky toast with actions; dedup per version via the store.
watch(
    () => updatesStore.shouldToast,
    (show) => {
        if (!show) return;
        const version = updatesStore.info?.latestVersion ?? '';
        toast.add(
            updatesStore.updateReady
                ? { group: 'updates', severity: 'success', summary: 'Update ready', detail: `PlexCord ${version} has been downloaded — restart to apply.` }
                : { group: 'updates', severity: 'info', summary: 'Update available', detail: `PlexCord ${version} is out. This platform updates manually — download it from the releases page.` }
        );
    },
    { immediate: true }
);

function dismissUpdateToast() {
    toast.removeGroup('updates');
    updatesStore.dismissToast();
}

async function restartForUpdate() {
    toast.removeGroup('updates');
    try {
        await updatesStore.restart();
    } catch (error) {
        toast.add({ severity: 'error', summary: 'Failed to restart', detail: error?.message || 'PlexCord could not restart itself — please restart it manually.', life: 8000 });
    }
}

function openUpdateRelease() {
    if (updatesStore.info?.releaseUrl) {
        OpenReleaseURL(updatesStore.info.releaseUrl);
    } else {
        OpenReleasesPage();
    }
}

function viewUpdateInSettings() {
    toast.removeGroup('updates');
    updatesStore.dismissToast();
    router.push('/settings');
}

// Shell-level playback subscription: the topbar headline needs live
// playback events on every page, not just the Dashboard (refcounted, F35).
usePlayback();

// ---- Keyboard shortcuts (spec §5.1 / §6.7) --------------------------------
// Ctrl/⌘+P toggle pause · Ctrl/⌘+, settings · Ctrl/⌘+R retry failed connection
const handleShortcuts = (event) => {
    if (!(event.ctrlKey || event.metaKey) || event.altKey || event.shiftKey) return;
    // Never fire while an input-like element has focus (§6.7).
    if (event.target?.closest?.('input, textarea, [contenteditable], .p-inputtext')) return;

    const key = event.key?.toLowerCase();
    if (key === 'p') {
        event.preventDefault();
        presenceStore.toggle();
    } else if (key === ',') {
        event.preventDefault();
        router.push(router.currentRoute.value.path.startsWith('/settings') ? '/' : '/settings');
    } else if (key === 'r') {
        event.preventDefault();
        if (plexStore.hasError) {
            plexStore.retry();
        } else if (discordStore.hasError) {
            discordStore.retry();
        } else {
            toast.add({ severity: 'secondary', summary: 'Nothing to retry', detail: 'Both connections are healthy.', life: 4000 });
        }
    }
};

onMounted(() => {
    // Global store lifecycle lives here: the shell (topbar nodes + headline)
    // needs connection + presence state on every page. All initialize()
    // calls are idempotent; pages may call them too.
    plexStore.initialize();
    discordStore.initialize();
    presenceStore.initialize();
    updatesStore.initialize();
    window.addEventListener('keydown', handleShortcuts);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleShortcuts);
    // Leaving the shell entirely (e.g. into /setup): release the listeners.
    plexStore.cleanup();
    discordStore.cleanup();
    updatesStore.cleanup();
});
</script>

<template>
    <div class="layout-wrapper">
        <app-topbar></app-topbar>
        <div class="layout-main-container">
            <div class="layout-main">
                <router-view v-slot="{ Component }">
                    <!-- M7 route transition (Dashboard ⇄ Settings) -->
                    <Transition name="pc-route" mode="out-in">
                        <component :is="Component" />
                    </Transition>
                </router-view>
            </div>
            <app-footer></app-footer>
        </div>

        <!-- Sticky update notification (no life => stays until acted on).
             Separate group so ordinary toasts keep the default host in App.vue. -->
        <Toast group="updates" @close="dismissUpdateToast">
            <template #message="{ message }">
                <div class="update-toast">
                    <span class="update-toast-summary">{{ message.summary }}</span>
                    <p class="update-toast-detail">{{ message.detail }}</p>
                    <div class="update-toast-actions">
                        <button v-if="updatesStore.updateReady" type="button" class="pc-btn pc-btn--primary pc-btn--sm" @click="restartForUpdate"><i class="pi pi-refresh" aria-hidden="true"></i>Restart now</button>
                        <button v-else type="button" class="pc-btn pc-btn--primary pc-btn--sm" @click="openUpdateRelease"><i class="pi pi-download" aria-hidden="true"></i>Download</button>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="viewUpdateInSettings">View in Settings</button>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="dismissUpdateToast">Later</button>
                    </div>
                </div>
            </template>
        </Toast>
    </div>
</template>

<style scoped>
.update-toast {
    flex: 1;
    min-width: 0;
}
.update-toast-summary {
    font-weight: 600;
}
.update-toast-detail {
    margin: 4px 0 0;
    font-size: var(--pc-text-caption);
}
.update-toast-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 10px;
}
.update-toast-actions .pc-btn {
    white-space: nowrap;
}
.update-toast-actions .pi {
    font-size: 12px;
}
</style>
