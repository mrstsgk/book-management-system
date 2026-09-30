import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'
import type { TagResponse } from '@/api/generated/api.schemas'
import { BookForm } from './BookForm'

const tags: TagResponse[] = [
  { id: 1, name: '設計' },
  { id: 2, name: 'データ' },
  { id: 3, name: 'クラウド' },
]

// 本の登録・編集で共用するフォーム。一覧の要約（ErrorSummary）と項目ごとの文言の両方で誤りを伝える
const meta = {
  title: 'AdminBookEditor/BookForm',
  component: BookForm,
  args: {
    tags,
    submitLabel: '登録する',
    onChange: fn(),
    onSubmit: fn(),
    submitting: false,
  },
} satisfies Meta<typeof BookForm>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = {
  args: {
    values: {
      titleOverride: '',
      summary: '',
      comment: '',
      rating: 0,
      tagIds: [],
    },
    errors: {},
  },
}

export const Filled: Story = {
  args: {
    values: {
      titleOverride: '',
      summary: '設計の考え方が体系的に学べる良書',
      comment:
        '悪いコードの例と良いコードの例を並べて説明していて分かりやすかった。',
      rating: 5,
      tagIds: [1],
    },
    errors: {},
  },
}

export const WithErrors: Story = {
  args: {
    values: {
      titleOverride: '',
      summary: '',
      comment: '',
      rating: 0,
      tagIds: [],
    },
    errors: {
      summary: '一言まとめを入力してください',
      comment: '感想を入力してください',
      rating: '評価を選んでください',
    },
  },
}
