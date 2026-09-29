import { defineConfig } from 'orval'

export default defineConfig({
  web: {
    // swag 排出の Swagger 2.0。Orval が OpenAPI 3 へ変換して読む
    input: '../../backend/api/docs/swagger.yaml',
    output: {
      mode: 'split',
      target: 'src/api/generated/api.ts',
      // 削除されたエンドポイントの生成物を残さない
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
          // 一覧の「もっと見る」用。offset をページの引数にした infinite query も生成する。
          // swag は operationId を出さないため、キーは Orval が method + path から作る名前（PascalCase）
          GetApiBooks: {
            query: { useInfinite: true, useInfiniteQueryParam: 'offset' },
          },
        },
      },
    },
  },
})
