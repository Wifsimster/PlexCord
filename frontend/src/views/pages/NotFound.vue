<script setup></script>

<template>
    <div class="notfound-wrapper">
        <div class="pc-panel notfound-panel">
            <span class="pc-mono notfound-code">404</span>
            <p class="notfound-caption">This route doesn't exist.</p>
            <router-link to="/" class="pc-btn pc-btn--primary">Back to dashboard</router-link>
        </div>
    </div>
</template>

<style scoped>
.notfound-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 100vh;
    background: var(--pc-bg);
}

.notfound-panel {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    text-align: center;
}

.notfound-code {
    font-size: var(--pc-text-display);
    color: var(--pc-text-faint);
}

.notfound-caption {
    font-size: var(--pc-text-caption);
    color: var(--pc-text-muted);
    margin: 0;
}
</style>
