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
        <div class="text-center mb-12 max-w-2xl mx-auto animate-fadein">
            <h2 class="text-3xl md:text-5xl font-bold mb-6 bg-clip-text text-transparent bg-gradient-to-r from-orange-500 to-indigo-500">Bridge Your Music</h2>
            <p class="text-xl text-surface-600 dark:text-surface-300 leading-relaxed mb-8">
                Seamlessly display your Plex music activity on Discord. <br />
                Rich Presence, automatic updates, and zero hassle.
            </p>
            <Button label="Get Started" icon="pi pi-arrow-right" iconPos="right" size="large" rounded @click="startSetup" class="px-8 py-4 text-xl shadow-lg hover:shadow-xl transition-all" />
        </div>

        <!-- Features Grid -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 w-full max-w-4xl mb-12">
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

        <!-- Requirements -->
        <div class="text-center text-sm text-surface-500 dark:text-surface-500">
            <p>Requires Plex Media Server and Discord Desktop App</p>
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
