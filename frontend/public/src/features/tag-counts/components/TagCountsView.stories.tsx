import type { Meta, StoryObj } from '@storybook/react-vite'
import { MemoryRouter } from 'react-router-dom'
import { TagCountsView } from './TagCountsView'

// 分野別の集計の一覧部分。各行はその分野で絞り込んだ一覧へのリンク（棒は最多の分野を100%とした割合）
const meta = {
  title: 'features/tag-counts/TagCountsView',
  component: TagCountsView,
  decorators: [
    (Story) => (
      <MemoryRouter>
        <Story />
      </MemoryRouter>
    ),
  ],
} satisfies Meta<typeof TagCountsView>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    items: [
      { id: 1, name: '設計', bookCount: 5 },
      { id: 2, name: 'クラウド', bookCount: 3 },
      { id: 3, name: 'データ', bookCount: 2 },
      { id: 4, name: 'ネットワーク', bookCount: 1 },
      { id: 5, name: 'プロダクト', bookCount: 1 },
    ],
  },
}

export const SingleTag: Story = {
  args: { items: [{ id: 1, name: '設計', bookCount: 1 }] },
}
