import { defineConfig, devices } from '@playwright/test'

// ローカル実行のみ（CI には載せない。実 DB・実バックエンドが要り重いため）。
// 管理 API（:8081）と管理画面の開発サーバー（:3001）が起動済みである前提で、
// このファイルからは起動しない（バックエンドはフロントエンドの管理外のプロセスのため）。
export default defineConfig({
  testDir: './src/testing/e2e',
  fullyParallel: false,
  // spec 間で管理画面の状態（本の一覧・タグの一覧）を共有するため、複数 spec を同時に走らせない
  workers: 1,
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:3001',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
