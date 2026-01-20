<script setup>
import { ref, onMounted } from 'vue';
import { GetVersion } from '../../wailsjs/go/main/App';

const version = ref('');

onMounted(async () => {
    try {
        const versionInfo = await GetVersion();
        version.value = `v${versionInfo.version}`;
    } catch (error) {
        console.error('Failed to get version:', error);
        version.value = 'Unknown';
    }
});
</script>

<template>
    <div class="layout-footer">
        <span class="font-semibold">PlexCord {{ version }}</span>
    </div>
</template>
