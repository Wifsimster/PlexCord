// ESLint v9+ flat config — replaces .eslintrc.cjs after the eslint@^10 bump.
const pluginVue = require('eslint-plugin-vue');
const skipFormatting = require('@vue/eslint-config-prettier/skip-formatting');

module.exports = [
    {
        ignores: ['node_modules/**', 'dist/**', 'wailsjs/**', 'build/**', 'coverage/**']
    },
    ...pluginVue.configs['flat/essential'],
    skipFormatting,
    {
        rules: {
            'vue/multi-word-component-names': 'off',
            'vue/no-reserved-component-names': 'off',
            // `component-tags-order` was renamed to `block-order` in eslint-plugin-vue v10.
            'vue/block-order': ['error', { order: ['script', 'template', 'style'] }]
        }
    }
];
