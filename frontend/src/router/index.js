import AppLayout from '@/layout/AppLayout.vue';
import { createRouter, createWebHashHistory } from 'vue-router';
import { CheckSetupComplete } from '../../wailsjs/go/main/App';

const router = createRouter({
    history: createWebHashHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/setup',
            component: () => import('@/views/SetupWizard.vue'),
            redirect: '/setup/welcome',
            children: [
                {
                    path: 'welcome',
                    name: 'setup-welcome',
                    component: () => import('@/views/SetupWelcome.vue')
                },
                {
                    path: 'plex',
                    name: 'setup-plex',
                    component: () => import('@/views/SetupPlex.vue')
                },
                {
                    path: 'user',
                    name: 'setup-user',
                    component: () => import('@/views/SetupPlexUser.vue')
                },
                {
                    path: 'discord',
                    name: 'setup-discord',
                    component: () => import('@/views/SetupDiscord.vue')
                },
                {
                    path: 'complete',
                    name: 'setup-complete',
                    component: () => import('@/views/SetupComplete.vue')
                }
            ]
        },
        {
            path: '/',
            component: AppLayout,
            children: [
                {
                    path: '',
                    name: 'dashboard',
                    component: () => import('@/views/pages/Dashboard.vue')
                },
                {
                    path: 'settings',
                    name: 'settings',
                    component: () => import('@/views/pages/Settings.vue')
                },
                {
                    path: 'pages/dashboard',
                    redirect: '/'
                }
            ]
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'notfound',
            component: () => import('@/views/pages/NotFound.vue')
        }
    ]
});

router.beforeEach(async (to, from, next) => {
    if (to.path.startsWith('/setup')) {
        next();
        return;
    }

    try {
        const isSetupComplete = await CheckSetupComplete();

        if (!isSetupComplete) {
            next('/setup/welcome');
        } else {
            next();
        }
    } catch (error) {
        console.error('Failed to check setup status:', error);
        next();
    }
});

export default router;
