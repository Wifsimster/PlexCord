<script setup>
import { onMounted } from 'vue';
import DiscordPreview from '@/components/setup/DiscordPreview.vue';
import { StartSessionPolling, ConnectDiscord, IsDiscordConnected } from '../../wailsjs/go/main/App';

// Note: The "Finish Setup" button is handled by SetupWizard.vue
// This component displays the completion content with live preview

// Start polling and ensure Discord is connected when component mounts
onMounted(async () => {
    try {
        // Ensure Discord is connected for live preview
        const isConnected = await IsDiscordConnected();
        if (!isConnected) {
            console.log('Discord not connected, connecting now...');
            await ConnectDiscord('');
        }

        await StartSessionPolling();
        console.log('Session polling started for preview');
    } catch (error) {
        // Silently handle error - polling might already be running
        console.log('Setup complete initialization:', error);
    }
});
</script>

<template>
    <div class="max-w-4xl mx-auto">
        <div class="py-4">
            <!-- Success Header -->
            <div class="text-center mb-8">
                <div class="mb-6 animate-scaleIn inline-block">
                    <i class="pi pi-check-circle text-6xl text-green-500"></i>
                </div>
                <h2 class="text-3xl font-bold mb-3 text-surface-900 dark:text-surface-0">You're All Set!</h2>
                <p class="text-lg text-surface-600 dark:text-surface-400">PlexCord is ready to display your Plex music activity on Discord.</p>
            </div>

            <!-- Discord Preview Section -->
            <div class="mb-8 max-w-lg mx-auto">
                <h3 class="text-lg font-semibold mb-4 text-center text-surface-900 dark:text-surface-0">Discord Status Preview</h3>
                <div class="bg-surface-100 dark:bg-surface-800 p-6 rounded-xl shadow-inner">
                    <DiscordPreview />
                </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- What's Next Section -->
                <div class="bg-surface-50 dark:bg-surface-800/50 p-6 rounded-xl border border-surface-200 dark:border-surface-700">
                    <h3 class="text-lg font-semibold mb-4 text-surface-900 dark:text-surface-0">What happens next?</h3>
                    <ul class="space-y-4">
                        <li class="flex items-start gap-3">
                            <div class="mt-1 bg-green-100 dark:bg-green-900/30 p-1.5 rounded-full text-green-600 dark:text-green-400">
                                <i class="pi pi-check text-xs font-bold"></i>
                            </div>
                            <span class="text-surface-700 dark:text-surface-300">Your Discord status will automatically update when you play music on Plex</span>
                        </li>
                        <li class="flex items-start gap-3">
                            <div class="mt-1 bg-green-100 dark:bg-green-900/30 p-1.5 rounded-full text-green-600 dark:text-green-400">
                                <i class="pi pi-check text-xs font-bold"></i>
                            </div>
                            <span class="text-surface-700 dark:text-surface-300">PlexCord will run in the background and minimize to the system tray</span>
                        </li>
                        <li class="flex items-start gap-3">
                            <div class="mt-1 bg-green-100 dark:bg-green-900/30 p-1.5 rounded-full text-green-600 dark:text-green-400">
                                <i class="pi pi-check text-xs font-bold"></i>
                            </div>
                            <span class="text-surface-700 dark:text-surface-300">You can change settings anytime from the dashboard</span>
                        </li>
                    </ul>
                </div>

                <!-- Tip Section -->
                <div class="bg-blue-50 dark:bg-blue-900/10 p-6 rounded-xl border border-blue-100 dark:border-blue-800/30 flex flex-col justify-center">
                    <div class="flex items-start gap-4">
                        <i class="pi pi-info-circle text-2xl text-blue-500 mt-1"></i>
                        <div>
                            <span class="font-bold text-blue-700 dark:text-blue-300 block mb-2">Pro Tip</span>
                            <span class="text-surface-700 dark:text-surface-300 leading-relaxed">
                                Start playing music on Plex now to see the live preview update above! It's the best way to verify everything is working correctly before you finish.
                            </span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.animate-scaleIn {
    animation: scaleIn 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes scaleIn {
    from {
        transform: scale(0);
        opacity: 0;
    }
    to {
        transform: scale(1);
        opacity: 1;
    }
}
</style>
