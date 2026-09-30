import { defineConfig, devices } from '@playwright/test'

// ローカル実行のみ（ADR: docs/adr/2026-09-30-introduce-playwright-e2e.md）。
// 実バックエンド（:8080）と実フロントエンド開発サーバー（:3000）が起動済みである前提で、
// このファイルからは起動しない（バックエンドはフロントエンドの管理外のプロセスのため）。
export default defineConfig({
  testDir: './src/testing/e2e',
  fullyParallel: false,
  // spec 間で管理画面の状態（本の一覧・タグの一覧）を共有するため、複数 spec を同時に走らせない
  workers: 1,
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
