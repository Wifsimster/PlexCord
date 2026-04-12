<script setup>
import Button from 'primevue/button';
import { useRouter } from 'vue-router';
import { useSetupStore } from '@/stores/setup';

const router = useRouter();
const setupStore = useSetupStore();

const startSetup = () => {
    setupStore.nextStep();
    router.push('/setup/plex');
};
</script>

<template>
    <div class="flex flex-col items-center justify-center py-10">
        <!-- Hero Section -->
        <div class="text-center mb-10 max-w-2xl mx-auto animate-fadein">
            <h2 class="text-3xl md:text-5xl font-bold mb-6 bg-clip-text text-transparent bg-gradient-to-r from-orange-500 to-indigo-500">Bridge Your Music</h2>
            <p class="text-xl text-surface-600 dark:text-surface-300 leading-relaxed mb-4">
                PlexCord shows what you're listening to on Plex as a Discord Rich Presence status &mdash; complete with album art, track info, and live progress.
            </p>
            <p class="text-surface-500 dark:text-surface-400 mb-8">
                Setup takes about <strong>2 minutes</strong>. Let's get you connected.
            </p>
            <Button label="Get Started" icon="pi pi-arrow-right" iconPos="right" size="large" rounded @click="startSetup" class="px-8 py-4 text-xl shadow-lg hover:shadow-xl transition-all" />
        </div>

        <!-- What You'll Need -->
        <div class="w-full max-w-4xl mb-10">
            <h3 class="text-center text-lg font-semibold text-surface-700 dark:text-surface-200 mb-4">Before you start, make sure you have</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="flex items-start gap-3 bg-surface-50 dark:bg-surface-800/50 p-4 rounded-xl border border-surface-100 dark:border-surface-700">
                    <div class="w-10 h-10 bg-orange-100 dark:bg-orange-900/30 rounded-full flex items-center justify-center text-orange-500 shrink-0">
                        <i class="pi pi-server text-lg"></i>
                    </div>
                    <div>
                        <div class="font-medium text-surface-900 dark:text-surface-0">A Plex account</div>
                        <p class="text-sm text-surface-500 dark:text-surface-400">With access to a Plex Media Server that has music libraries.</p>
                    </div>
                </div>
                <div class="flex items-start gap-3 bg-surface-50 dark:bg-surface-800/50 p-4 rounded-xl border border-surface-100 dark:border-surface-700">
                    <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center text-indigo-500 shrink-0">
                        <i class="pi pi-discord text-lg"></i>
                    </div>
                    <div>
                        <div class="font-medium text-surface-900 dark:text-surface-0">Discord running</div>
                        <p class="text-sm text-surface-500 dark:text-surface-400">The Discord desktop app must be open on this computer.</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Features Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-4xl mb-10">
            <div class="bg-surface-50 dark:bg-surface-800/50 p-6 rounded-xl border border-surface-100 dark:border-surface-700 hover:border-primary-500/50 transition-colors">
                <div class="w-12 h-12 bg-orange-100 dark:bg-orange-900/30 rounded-full flex items-center justify-center mb-4 text-orange-500">
                    <i class="pi pi-server text-2xl"></i>
                </div>
                <h3 class="text-lg font-semibold mb-2">Connect Plex</h3>
                <p class="text-surface-600 dark:text-surface-400 text-sm">Link your Plex Media Server securely with a PIN code. Supports local and remote servers.</p>
            </div>

            <div class="bg-surface-50 dark:bg-surface-800/50 p-6 rounded-xl border border-surface-100 dark:border-surface-700 hover:border-primary-500/50 transition-colors">
                <div class="w-12 h-12 bg-indigo-100 dark:bg-indigo-900/30 rounded-full flex items-center justify-center mb-4 text-indigo-500">
                    <i class="pi pi-discord text-2xl"></i>
                </div>
                <h3 class="text-lg font-semibold mb-2">Sync Discord</h3>
                <p class="text-surface-600 dark:text-surface-400 text-sm">Automatically updates your status on Discord with rich album art and track details.</p>
            </div>

            <div class="bg-surface-50 dark:bg-surface-800/50 p-6 rounded-xl border border-surface-100 dark:border-surface-700 hover:border-primary-500/50 transition-colors">
                <div class="w-12 h-12 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center mb-4 text-green-500">
                    <i class="pi pi-bolt text-2xl"></i>
                </div>
                <h3 class="text-lg font-semibold mb-2">Set & Forget</h3>
                <p class="text-surface-600 dark:text-surface-400 text-sm">Runs silently in the background. Auto-starts with your system and minimizes to tray.</p>
            </div>
        </div>
    </div>
</template>

<style scoped>
.animate-fadein {
    animation: fadein 0.8s ease-out;
}

@keyframes fadein {
    from {
        opacity: 0;
        transform: translateY(20px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>
