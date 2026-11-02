import type { Meta, StoryObj } from '@storybook/react-vite'
import { Button } from '@book-management/ui'

const meta = {
  title: 'DS/Button',
  component: Button,
  args: {
    children: 'ボタン',
    variant: 'solid-fill',
    size: 'md',
  },
  parameters: {
    docs: {
      description: {
        component: 'デジタル庁 DS の Button。Presentational 状態カタログ用。',
      },
    },
  },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const SolidFill: Story = {}

export const Outline: Story = {
  args: { variant: 'outline' },
}

export const Text: Story = {
  args: { variant: 'text' },
}

export const Disabled: Story = {
  args: { 'aria-disabled': true },
}
