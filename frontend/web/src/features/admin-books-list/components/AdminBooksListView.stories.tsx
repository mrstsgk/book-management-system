import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { fn } from 'storybook/test'
import { AdminBooksListView } from './AdminBooksListView'

const books = [
  {
    id: 1,
    title: 'データ指向アプリケーションデザイン 第2版',
    rating: 5,
    tags: ['データ', '設計'],
    summary: '分散システムの原理を整理した本',
  },
  {
    id: 2,
    title: '改訂新版　良いコード／悪いコードで学ぶ設計入門',
    rating: 3,
    tags: [],
    summary: '（準備中）',
  },
]

// 本の一覧（管理）。表・読み込み中・お知らせ・0件・エラーの状態
const meta = {
  title: 'AdminBooksList/AdminBooksListView',
  component: AdminBooksListView,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <div className="bg-ink-100">
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
  args: {
    books,
    total: 2,
    isLoading: false,
    isError: false,
    hasMore: false,
    isLoadingMore: false,
    onRetry: fn(),
    onLoadMore: fn(),
  },
} satisfies Meta<typeof AdminBooksListView>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Notice: Story = {
  args: { notice: '「ネットワークはなぜつながるのか」を削除しました。' },
}

export const LoadingMore: Story = {
  args: { hasMore: true, isLoadingMore: true },
}

export const Loading: Story = { args: { books: [], isLoading: true } }

export const Empty: Story = { args: { books: [], total: 0 } }

export const Error: Story = { args: { books: [], isError: true } }
