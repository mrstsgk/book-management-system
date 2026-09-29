import type { Meta, StoryObj } from '@storybook/react-vite'
import { ErrorSummary } from './ErrorSummary'
import { FormField, describedBy } from './FormField'

const inputClass =
  'h-12 rounded-lg border border-ink-500 px-4 aria-[invalid=true]:border-2 aria-[invalid=true]:border-brand-700'

// 管理画面のフォームの部品。誤りは上部の要約と項目ごとの文言の両方で伝える（要件定義 §3.2）
const meta = {
  title: 'Form',
  parameters: {
    docs: {
      description: {
        component:
          'FormField（ラベル・必須/任意・文字数・エラー）と ErrorSummary（誤りの要約と各項目へのリンク）。',
      },
    },
  },
} satisfies Meta

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  render: () => (
    <FormField
      id="summary"
      label="一言まとめ"
      required
      count={{ current: 14, max: 100 }}
    >
      <input
        id="summary"
        defaultValue="[一言まとめ]"
        aria-describedby={describedBy('summary')}
        className={inputClass}
      />
    </FormField>
  ),
}

export const WithErrors: Story = {
  render: () => (
    <div className="flex flex-col gap-5">
      <ErrorSummary
        errors={[{ id: 'summary', message: '一言まとめ: 入力してください' }]}
      />
      <FormField
        id="summary"
        label="一言まとめ"
        required
        error="入力してください"
        count={{ current: 0, max: 100 }}
      >
        <input
          id="summary"
          aria-invalid="true"
          aria-describedby={describedBy('summary')}
          className={inputClass}
        />
      </FormField>
    </div>
  ),
}
