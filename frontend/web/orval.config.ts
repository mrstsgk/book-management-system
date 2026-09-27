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
      },
    },
  },
})
