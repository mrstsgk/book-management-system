import type { Meta, StoryObj } from '@storybook/react-vite'
import { TagsView } from './TagsView'

// タグの管理の一覧・追加・その場の名前変更・削除の確認。管理画面（自分だけ）専用
const meta = {
  title: 'features/admin-tags/TagsView',
  component: TagsView,
  args: {
    onAdd: async () => ({ ok: true }) as const,
    onRename: async () => ({ ok: true }) as const,
    onDelete: async () => ({ ok: true }) as const,
  },
} satisfies Meta<typeof TagsView>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: {
    items: [
      { id: 1, name: 'データ', version: 1, bookCount: 2 },
      { id: 2, name: '設計', version: 3, bookCount: 0 },
      { id: 3, name: 'ネットワーク', version: 1, bookCount: 5 },
    ],
  },
}

export const Empty: Story = {
  args: { items: [] },
}
