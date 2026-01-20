<script setup>
import { onMounted } from 'vue';

onMounted(() => {
    // Initialize dark mode
    document.documentElement.classList.add('dark');
});
</script>

<template>
    <router-view />
</template>

<style scoped></style>
