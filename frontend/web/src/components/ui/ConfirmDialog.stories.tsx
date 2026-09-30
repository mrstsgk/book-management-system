import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import { ConfirmDialog } from './ConfirmDialog'

const meta = {
  title: 'UI/ConfirmDialog',
  component: ConfirmDialog,
  args: {
    open: true,
    title: 'この本を削除しますか？',
    description:
      '「データ指向アプリケーションデザイン 第2版」と、その一言まとめ・感想・評価を削除します。元に戻せません。',
    confirmLabel: '削除する',
    onConfirm: fn(),
    onCancel: fn(),
  },
  parameters: {
    docs: {
      description: {
        component:
          '取り消せない操作（本・タグの削除）の確認。既定のフォーカスはキャンセル。Esc でも閉じる。',
      },
    },
  },
} satisfies Meta<typeof ConfirmDialog>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Busy: Story = { args: { busy: true } }
