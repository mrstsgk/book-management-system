import path from 'node:path'
import { fileURLToPath } from 'node:url'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

const rootDir = path.dirname(fileURLToPath(import.meta.url))
const testOrigin = 'http://localhost:3000'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.join(rootDir, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        // Host を書き換えない: バックエンドの Origin 検証（RequireSameOrigin）が
        // Origin と Host を比べるため、ブラウザの localhost:3000 のまま届ける
        changeOrigin: false,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: false,
      },
    },
  },
  test: {
    environment: 'jsdom',
    environmentOptions: { jsdom: { url: testOrigin } },
    setupFiles: ['./src/testing/setup.ts'],
    include: ['src/**/*.test.{ts,tsx}'],
    // .env* の VITE_API_BASE_URL（実 API）を継承すると MSW を素通りしうるため上書き。
    // '' にしないのは、Node の fetch が相対 URL を受け付けず（Failed to parse URL）
    // リクエストが MSW に届く前に失敗するため。jsdom の origin に揃え、MSW の
    // 相対パスハンドラ（location.origin 基準で解決）と一致させる
    env: {
      VITE_API_BASE_URL: testOrigin,
    },
  },
})
