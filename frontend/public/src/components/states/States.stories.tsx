import type { Meta, StoryObj } from '@storybook/react-vite'
import { EmptyState } from './EmptyState'
import { ErrorState } from './ErrorState'
import { LoadingState } from './LoadingState'

// 画面の状態（読み込み中・0件・エラー）の共通部品。要件定義 §2.4
const meta = {
  title: 'States',
  parameters: {
    docs: {
      description: {
        component:
          '一覧・詳細・分野別で共通の「読み込み中」「0件」「エラー」の表示。',
      },
    },
  },
} satisfies Meta

export default meta
type Story = StoryObj<typeof meta>

export const Loading: Story = { render: () => <LoadingState /> }

export const LoadingMore: Story = {
  render: () => <LoadingState label="続きを読み込み中…" />,
}

export const Empty: Story = {
  render: () => (
    <EmptyState
      title="条件に当てはまる本はありません"
      description="キーワードを変えるか、分野タグを「すべて」に戻してください。"
      action={
        <button
          type="button"
          className="h-11 rounded-lg border border-brand-700 bg-white px-6 font-bold text-brand-700"
        >
          検索と絞り込みを外す
        </button>
      }
    />
  ),
}

export const Error: Story = {
  render: () => (
    <ErrorState
      title="本の一覧を読み込めませんでした"
      description="サーバーに接続できません。時間をおいてもう一度お試しください。"
      onRetry={() => {}}
    />
  ),
}
