import js from '@eslint/js'
import vitest from '@vitest/eslint-plugin'
import prettier from 'eslint-config-prettier'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import testingLibrary from 'eslint-plugin-testing-library'
import globals from 'globals'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  {
    ignores: [
      '**/node_modules/',
      '**/dist/',
      '**/storybook-static/',
      // Orval 生成物。手で直さないうえ、TS のオーバーロード宣言などを誤検知する
      'web/src/api/generated/',
    ],
  },

  js.configs.recommended,
  tseslint.configs.recommended,

  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      globals: { ...globals.browser, ...globals.es2022 },
    },
    plugins: { 'react-hooks': reactHooks },
    rules: {
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'warn',
      // 分岐の多すぎる関数を読みにくさとして弾く（バックエンドの gocyclo と同じ上限 10）
      complexity: ['error', 10],
      // Orval の ErrorType<_Body> のように、_ 始まりは意図的な未使用
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
    },
  },

  {
    // HMR 対象はアプリのみ。packages/ui は variants 等を部品と同居で export する
    files: ['web/src/**/*.tsx'],
    plugins: { 'react-refresh': reactRefresh },
    rules: {
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
    },
  },

  {
    files: ['**/*.tsx'],
    ...jsxA11y.flatConfigs.recommended,
  },

  {
    // Node で動く設定ファイル（Vite / Storybook / Tailwind / Orval）
    files: [
      '**/*.config.{js,cjs,mjs,ts}',
      '**/tailwind.preset.cjs',
      '**/.storybook/**',
    ],
    languageOptions: { globals: globals.node },
  },
  {
    files: ['**/*.cjs'],
    rules: { '@typescript-eslint/no-require-imports': 'off' },
  },

  {
    files: ['**/*.test.{ts,tsx}'],
    ...vitest.configs.recommended,
  },
  {
    // 表形式のテストはケースを並べるほど数値が上がるだけで、読みにくさとは関係しない
    files: ['**/*.test.{ts,tsx}'],
    rules: { complexity: 'off' },
  },
  {
    files: ['**/*.test.{ts,tsx}'],
    ...testingLibrary.configs['flat/react'],
  },

  // 整形は Prettier に任せ、衝突するスタイル系ルールを無効化する（最後に置く）
  prettier,
)
