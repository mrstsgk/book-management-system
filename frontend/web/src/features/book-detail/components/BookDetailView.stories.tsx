import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { BookDetailView } from './BookDetailView'

const meta = {
  title: 'BookDetail/BookDetailView',
  component: BookDetailView,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
  args: {
    book: {
      id: 1,
      isbn: '9784297146221',
      title: '改訂新版　良いコード／悪いコードで学ぶ設計入門',
      authors: '仙塲大也',
      publisher: '技術評論社',
      coverSource: 'openbd',
      amazonUrl: 'https://www.amazon.co.jp/dp/4297146223',
      summary: '[一言まとめ]',
      comment: '[感想の本文]\n\n[2段落目]',
      rating: 5,
      tags: ['設計'],
    },
  },
  parameters: {
    docs: {
      description: {
        component:
          '読んだ本の詳細。書誌・書影・評価・分野タグ・一言まとめ・感想と、Amazon へのリンクを出す。',
      },
    },
  },
} satisfies Meta<typeof BookDetailView>

export default meta
type Story = StoryObj<typeof meta>

export const OpenBD: Story = {}

export const NoCover: Story = {
  args: {
    book: { ...meta.args.book, coverSource: undefined, amazonUrl: undefined },
  },
}
