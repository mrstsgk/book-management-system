import { defineConfig } from 'orval'

export default defineConfig({
  web: {
    // API 契約の正本 openapi/ をバンドルしたもの（BE の oapi-codegen と同じ入力。手編集しない）
    input: '../../openapi/dist/web.yaml',
    output: {
      mode: 'split',
      target: 'src/api/generated/api.ts',
      // 削除されたエンドポイントの生成物を残さない。zod/ は webZod の出力先なので消さない
      clean: ['**', '!zod', '!zod/**'],
      client: 'react-query',
      httpClient: 'fetch',
      mock: true,
      packageJson: 'package.json',
      override: {
        mutator: { path: 'src/api/mutator.ts', name: 'apiMutator' },
        // レスポンスは { status, data, headers } ではなくボディそのものを返す
        fetch: { includeHttpResponseReturnType: false },
        operations: {
          // 一覧の「もっと見る」用。offset をページの引数にした infinite query も生成する
          searchBooks: {
            query: { useInfinite: true, useInfiniteQueryParam: 'offset' },
          },
        },
      },
    },
  },
  // フォームの入力チェック用。OpenAPI の制約（required・長さ・範囲・形式）から zod スキーマを生成する
  webZod: {
    input: '../../openapi/dist/web.yaml',
    output: {
      mode: 'split',
      target: 'src/api/generated/zod/api.ts',
      client: 'zod',
      clean: true,
      packageJson: 'package.json',
    },
  },
})
