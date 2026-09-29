import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { fn } from 'storybook/test'
import { BooksListView } from './BooksListView'

const books = [
  {
    id: 1,
    title:
      '徹底攻略 AWS認定 ソリューションアーキテクト − アソシエイト教科書 第3版［SAA-C03］対応',
    authors: '鳥谷部昭寛, 宮口光平, 半田大樹, 株式会社ソキウス・ジャパン',
    summary: '[一言まとめ]',
    rating: 4,
    tags: ['クラウド'],
  },
  {
    id: 2,
    title: '改訂新版　良いコード／悪いコードで学ぶ設計入門',
    authors: '仙塲大也',
    summary: '[一言まとめ]',
    rating: 5,
    tags: ['設計'],
  },
  {
    id: 3,
    title: 'データ指向アプリケーションデザイン 第2版',
    authors: 'Martin Kleppmann',
    summary: '[一言まとめ]',
    rating: 5,
    tags: ['データ', '設計'],
  },
]

// 読んだ本の一覧（公開画面）。検索・分野タグの絞り込み・続きの読み込みと、読み込み中・0件・エラーの状態
const meta = {
  title: 'BooksList/BooksListView',
  component: BooksListView,
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
    total: 12,
    tags: [
      { id: 1, name: 'クラウド' },
      { id: 2, name: '設計' },
      { id: 3, name: 'データ' },
    ],
    hasFilter: false,
    isLoading: false,
    isError: false,
    hasMore: true,
    isLoadingMore: false,
    onRetry: fn(),
    onLoadMore: fn(),
    onSearch: fn(),
    onSelectTag: fn(),
    onClear: fn(),
  },
} satisfies Meta<typeof BooksListView>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Filtered: Story = {
  args: { query: '設計', tagId: 2, hasFilter: true, total: 2, hasMore: false },
}

export const LoadingMore: Story = { args: { isLoadingMore: true } }

export const Loading: Story = { args: { books: [], isLoading: true } }

export const Empty: Story = {
  args: { books: [], total: 0, hasMore: false },
}

export const NoResults: Story = {
  args: {
    books: [],
    total: 0,
    hasMore: false,
    query: 'Kubernetes',
    hasFilter: true,
  },
}

export const Error: Story = { args: { books: [], isError: true } }
