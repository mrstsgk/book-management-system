import eslintReact from '@eslint-react/eslint-plugin'
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
      '*/src/api/generated/',
    ],
  },

  js.configs.recommended,
  tseslint.configs.strict,
  // オブジェクトの型は interface に揃う（consistent-type-definitions の既定。architecture.md §9）
  tseslint.configs.stylistic,
  // React 19 で不要になった書き方（forwardRef 等）を検出する
  {
    files: ['**/*.{ts,tsx}'],
    ...eslintReact.configs['recommended-typescript'],
  },

  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      globals: { ...globals.browser, ...globals.es2022 },
    },
    plugins: { 'react-hooks': reactHooks },
    rules: {
      // React Compiler 由来のルール（purity・set-state-in-effect 等）も含む推奨セット
      ...reactHooks.configs.flat.recommended.rules,
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

  // import の規則（frontend/architecture.md §7・§9）。no-restricted-imports は後の設定が前を
  // 置き換えるため、対象ごとに禁止するものを全部書く。feature・共有のディレクトリを足したらここにも足す
  // 動的 import（import()）は no-restricted-imports の対象外なので、no-restricted-syntax で同じ内容を検査する
  ...(() => {
    // アプリごとの feature。アプリや feature を足したらここにも足す
    const appFeatures = {
      web: [
        'admin-book-editor',
        'admin-books-list',
        'admin-login',
        'admin-tags',
        'book-detail',
        'books-list',
        'tag-counts',
      ],
    }
    const parentMessage =
      '別ディレクトリの import は @/ から書く（architecture.md §9）'
    // forbidden: import を禁止する feature 名。'*' は全 feature
    const restrict = (forbidden, message) => {
      const target = forbidden.map((name) => (name === '*' ? '' : `/${name}`))
      return {
        'no-restricted-imports': [
          'error',
          {
            patterns: [
              { group: ['../*', '../**'], message: parentMessage },
              ...target.map((t) => ({
                group: [`@/features${t}`, `@/features${t}/**`],
                message,
              })),
            ],
          },
        ],
        'no-restricted-syntax': [
          'error',
          {
            selector: 'ImportExpression[source.value=/^\\.\\.\\//]',
            message: parentMessage,
          },
          ...target.map((t) => ({
            selector: `ImportExpression[source.value=/^@\\/features${t.replace('/', '\\/')}(\\/|$)/]`,
            message,
          })),
        ],
      }
    }
    return Object.entries(appFeatures).flatMap(([app, features]) => [
      { files: [`${app}/src/**/*.{ts,tsx}`], rules: restrict([]) },
      // feature 同士は直接 import しない
      ...features.map((feature) => ({
        files: [`${app}/src/features/${feature}/**/*.{ts,tsx}`],
        rules: restrict(
          features.filter((other) => other !== feature),
          'feature 同士は直接 import しない。app で合成するか、共有へ昇格する（architecture.md §7）',
        ),
      })),
      // 共有（shared）は features を import しない。依存は shared ← features ← app の向きだけ
      {
        files: [
          `${app}/src/{api,components,hooks,lib,testing,types,utils}/**/*.{ts,tsx}`,
        ],
        rules: restrict(
          ['*'],
          '共有コードは features を import しない（依存は shared ← features ← app。architecture.md §7）',
        ),
      },
    ])
  })(),

  {
    // default export は Storybook の meta だけ（§9）。Story と設定ファイルは対象外
    files: ['web/src/**/*.{ts,tsx}', 'packages/ui/src/**/*.{ts,tsx}'],
    ignores: ['**/*.stories.{ts,tsx}'],
    rules: {
      // `export { Foo as default }` や再 export の default も禁止する
      'no-restricted-exports': [
        'error',
        {
          restrictDefaultExports: {
            direct: true,
            named: true,
            defaultFrom: true,
            namedFrom: true,
            namespaceFrom: true,
          },
        },
      ],
    },
  },

  {
    // Node で動く設定ファイル（Vite / Storybook / Tailwind / Orval / Playwright）
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
    // テスト・Story・テスト用具（E2E を含む）では、no-op のコールバック（`() => {}`）、waitFor で存在を
    // 確かめた後の `!`、Router の位置を外側の変数に写す Probe コンポーネントを許す。
    // いずれも検証のための書き方で、本番コードの規則（純粋性・null 安全）を当てはめると回りくどくなるだけ
    files: [
      '**/*.test.{ts,tsx}',
      '**/*.stories.{ts,tsx}',
      '*/src/testing/**/*.{ts,tsx}',
    ],
    rules: {
      '@typescript-eslint/no-empty-function': 'off',
      '@typescript-eslint/no-non-null-assertion': 'off',
      'react-hooks/globals': 'off',
    },
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
