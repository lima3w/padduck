import globals from 'globals'
import reactPlugin from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'

export default [
  {
    files: ['src/**/*.{js,jsx}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: { ...globals.browser },
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: {
      react: reactPlugin,
      'react-hooks': reactHooks,
    },
    settings: { react: { version: 'detect' } },
    rules: {
      ...reactPlugin.configs.recommended.rules,
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      // Core hooks rules
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      // Disable v7 experimental rules that need larger refactors
      'react-hooks/set-state-in-effect': 'off',
      'react-hooks/static-components': 'off',
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
    },
  },
  {
    files: ['src/test/**/*.{js,jsx}', 'src/**/*.test.{js,jsx}'],
    languageOptions: {
      globals: { ...globals.browser, ...globals.node },
    },
  },
]
