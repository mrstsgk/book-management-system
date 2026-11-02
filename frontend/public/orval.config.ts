import { defineConfig } from 'orval'

export default defineConfig({
  public: {
    // 閲覧チャネルの API 契約 openapi/ をバンドルしたもの（BE の oapi-codegen と同じ入力。手編集しない）
    input: '../../openapi/dist/public.yaml',
    output: {
      mode: 'split',
      target: 'src/api/generated/api.ts',
      clean: true,
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
})
