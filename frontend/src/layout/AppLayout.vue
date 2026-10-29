<script setup>
import { onBeforeUnmount, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useToast } from 'primevue/usetoast';
import Toast from 'primevue/toast';
import AppFooter from './AppFooter.vue';
import AppTopbar from './AppTopbar.vue';
import { usePlayback } from '@/composables/usePlayback';
import { useLayout } from '@/layout/composables/layout';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';
import { useUpdatesStore } from '@/stores/updates';
import { OpenReleaseURL, OpenReleasesPage } from '../../wailsjs/go/main/App';

const router = useRouter();
const toast = useToast();
const { t } = useI18n();

const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();
const presenceStore = usePresenceStore();
const updatesStore = useUpdatesStore();
const { toggleDarkMode } = useLayout();

watch(
    () => updatesStore.shouldToast,
    (show) => {
        if (!show) return;
        const version = updatesStore.info?.latestVersion ?? '';
        toast.add(
            updatesStore.updateReady
                ? { group: 'updates', severity: 'success', summary: t('layout.updateReadySummary'), detail: t('layout.updateReadyDetail', { version }) }
                : { group: 'updates', severity: 'info', summary: t('layout.updateAvailableSummary'), detail: t('layout.updateAvailableDetail', { version }) }
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
        toast.add({ severity: 'error', summary: t('settings.toast.restartFailed'), detail: error?.message || t('settings.toast.restartFailedDetail'), life: 8000 });
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

usePlayback();

const handleShortcuts = (event) => {
    if (event.target?.closest?.('input, textarea, [contenteditable], .p-inputtext')) return;

    if (event.altKey && !event.ctrlKey && !event.metaKey && !event.shiftKey) {
        if (event.key?.toLowerCase() === 'd' || event.code === 'KeyD') {
            event.preventDefault();
            toggleDarkMode();
        }
        return;
    }

    if (!(event.ctrlKey || event.metaKey) || event.altKey || event.shiftKey) return;

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
            plexStore.retry().catch((error) => console.error('Plex reconnect failed:', error));
        } else if (discordStore.hasError) {
            discordStore.retry().catch((error) => console.error('Discord reconnect failed:', error));
        } else {
            toast.add({ severity: 'secondary', summary: t('layout.nothingToRetry'), detail: t('layout.connectionsHealthy'), life: 4000 });
        }
    }
};

onMounted(() => {
    plexStore.initialize();
    discordStore.initialize();
    presenceStore.initialize();
    updatesStore.initialize();
    window.addEventListener('keydown', handleShortcuts);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleShortcuts);
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
                    <Transition name="pc-route" mode="out-in">
                        <component :is="Component" />
                    </Transition>
                </router-view>
            </div>
            <app-footer></app-footer>
        </div>

        <Toast group="updates" @close="dismissUpdateToast">
            <template #message="{ message }">
                <div class="update-toast">
                    <span class="update-toast-summary">{{ message.summary }}</span>
                    <p class="update-toast-detail">{{ message.detail }}</p>
                    <div class="update-toast-actions">
                        <button v-if="updatesStore.updateReady" type="button" class="pc-btn pc-btn--primary pc-btn--sm" @click="restartForUpdate"><i class="pi pi-refresh" aria-hidden="true"></i>{{ $t('settings.restartNow') }}</button>
                        <button v-else type="button" class="pc-btn pc-btn--primary pc-btn--sm" @click="openUpdateRelease"><i class="pi pi-download" aria-hidden="true"></i>{{ $t('settings.download') }}</button>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="viewUpdateInSettings">{{ $t('layout.updateViewInSettings') }}</button>
                        <button type="button" class="pc-btn pc-btn--ghost pc-btn--sm" @click="dismissUpdateToast">{{ $t('layout.updateLater') }}</button>
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
