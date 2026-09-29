import type { Meta, StoryObj } from '@storybook/react-vite'
import { RatingStars } from './RatingStars'

const meta = {
  title: 'UI/RatingStars',
  component: RatingStars,
  args: { rating: 4 },
  argTypes: { rating: { control: { type: 'range', min: 1, max: 5 } } },
  parameters: {
    docs: {
      description: {
        component: '本の評価（1〜5）を星で表す。読み上げは「評価 N / 5」。',
      },
    },
  },
} satisfies Meta<typeof RatingStars>

export default meta
type Story = StoryObj<typeof meta>

export const Small: Story = {}

export const Large: Story = { args: { rating: 5, size: 'lg' } }
