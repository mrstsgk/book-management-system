import { describe, expect, it } from 'vitest'
import { describedBy } from './describedBy'

describe('describedBy', () => {
  it('ヒント・エラー・文字数の順に、FormField が付ける id を並べる', () => {
    expect(describedBy('summary')).toBe(
      'summary-hint summary-error summary-count',
    )
  })
})
