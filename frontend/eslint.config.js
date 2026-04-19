import js from '@eslint/js';
import pluginVue from 'eslint-plugin-vue';
import prettierConfig from '@vue/eslint-config-prettier';
import globals from 'globals';

export default [
    {
        ignores: ['node_modules/**', 'dist/**', 'wailsjs/**']
    },
    js.configs.recommended,
    ...pluginVue.configs['flat/essential'],
    prettierConfig,
    {
        languageOptions: {
            ecmaVersion: 'latest',
            sourceType: 'module',
            globals: {
                ...globals.browser,
                ...globals.node
            }
        },
        rules: {
            'vue/multi-word-component-names': 'off',
            'vue/no-reserved-component-names': 'off',
            'vue/block-order': [
                'error',
                {
                    order: ['script', 'template', 'style']
                }
            ],
            'no-unused-vars': ['error', { caughtErrors: 'none' }]
        }
    }
];
