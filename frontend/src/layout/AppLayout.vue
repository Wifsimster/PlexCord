<script setup>
import { onBeforeUnmount, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useToast } from 'primevue/usetoast';
import AppFooter from './AppFooter.vue';
import AppTopbar from './AppTopbar.vue';
import { usePlayback } from '@/composables/usePlayback';
import { usePlexConnectionStore } from '@/stores/plexConnection';
import { useDiscordConnectionStore } from '@/stores/discordConnection';
import { usePresenceStore } from '@/stores/presence';

const router = useRouter();
const toast = useToast();
const { t } = useI18n();

const plexStore = usePlexConnectionStore();
const discordStore = useDiscordConnectionStore();
const presenceStore = usePresenceStore();

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
            toast.add({ severity: 'secondary', summary: t('layout.nothingToRetry'), detail: t('layout.connectionsHealthy'), life: 4000 });
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
    window.addEventListener('keydown', handleShortcuts);
});

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleShortcuts);
    // Leaving the shell entirely (e.g. into /setup): release the listeners.
    plexStore.cleanup();
    discordStore.cleanup();
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
    </div>
</template>
