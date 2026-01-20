<script setup>
import { useLayout } from '@/layout/composables/layout';
import { computed } from 'vue';
import AppFooter from './AppFooter.vue';
import AppTopbar from './AppTopbar.vue';

const { layoutConfig, layoutState } = useLayout();

const containerClass = computed(() => {
    return {};
});
</script>

<template>
    <div class="layout-wrapper" :class="containerClass">
        <app-topbar></app-topbar>
        <div class="layout-main-container">
            <div class="layout-main">
                <router-view></router-view>
            </div>
            <app-footer></app-footer>
        </div>
        <div class="layout-mask animate-fadein"></div>
    </div>
    <Toast />
</template>
