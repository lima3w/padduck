import js from '@eslint/js'
import globals from 'globals'
import eslintReact from '@eslint-react/eslint-plugin'
import reactHooks from 'eslint-plugin-react-hooks'

const { configs: reactConfigs } = eslintReact

export default [
  js.configs.recommended,
  {
    files: ['src/**/*.{js,jsx}'],
    ...reactConfigs.recommended,
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: { ...globals.browser },
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: {
      ...reactConfigs.recommended.plugins,
      'react-hooks': reactHooks,
    },
    rules: {
      ...reactConfigs.recommended.rules,
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      'no-empty': ['error', { allowEmptyCatch: true }],
      'preserve-caught-error': 'warn',
      '@eslint-react/set-state-in-effect': 'off',
      '@eslint-react/static-components': 'off',
      '@eslint-react/no-nested-component-definitions': 'off',
      '@eslint-react/unsupported-syntax': 'off',
    },
  },
  {
    files: ['src/test/**/*.{js,jsx}', 'src/**/*.test.{js,jsx}'],
    languageOptions: {
      globals: { ...globals.browser, ...globals.node, ...globals.vitest },
    },
  },
]
