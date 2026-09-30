import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { AdminBooksListView } from './AdminBooksListView'

type Props = ComponentProps<typeof AdminBooksListView>

const filledBook = {
  id: 7,
  title: 'データ指向アプリケーションデザイン 第2版',
  rating: 5,
  tags: ['データ', '設計'],
  summary: '分散システムの原理を整理した本',
}

const unfilledBook = {
  id: 8,
  title: '改訂新版　良いコード／悪いコードで学ぶ設計入門',
  rating: 3,
  tags: [],
  summary: '（準備中）',
}

function setup(overrides: Partial<Props> = {}) {
  const props: Props = {
    books: [filledBook],
    total: 1,
    isLoading: false,
    isError: false,
    onRetry: vi.fn(),
    hasMore: false,
    isLoadingMore: false,
    onLoadMore: vi.fn(),
    ...overrides,
  }
  renderWithProviders(<AdminBooksListView {...props} />)
  return props
}

describe('AdminBooksListView', () => {
  it('本を1行ずつ、書名・評価・分野タグ・一言まとめの状態を並べ、編集へリンクする', () => {
    setup()

    const row = screen.getByRole('row', {
      name: /データ指向アプリケーションデザイン/,
    })
    expect(
      within(row).getByText('データ指向アプリケーションデザイン 第2版'),
    ).toBeVisible()
    expect(within(row).getByRole('img', { name: '評価 5 / 5' })).toBeVisible()
    expect(within(row).getByText('データ・設計')).toBeVisible()
    expect(within(row).getByText('記入済み')).toBeVisible()
    expect(within(row).getByRole('link', { name: '編集' })).toHaveAttribute(
      'href',
      '/admin/books/7/edit',
    )
  })

  it('分野タグが無ければ「—」、一言まとめが未記入なら「未記入」と出す', () => {
    setup({ books: [unfilledBook] })

    const row = screen.getByRole('row', { name: /良いコード/ })
    expect(within(row).getByText('—')).toBeVisible()
    expect(within(row).getByText('未記入')).toBeVisible()
  })

  it('「本を登録」へのリンクを出す', () => {
    setup()

    expect(screen.getByRole('link', { name: /本を登録/ })).toHaveAttribute(
      'href',
      '/admin/books/new',
    )
  })

  it('notice があれば完了のお知らせを表示する', () => {
    setup({ notice: '「ネットワークはなぜつながるのか」を削除しました。' })

    expect(screen.getByRole('status')).toHaveTextContent(
      '「ネットワークはなぜつながるのか」を削除しました。',
    )
  })

  it('notice が無ければお知らせを出さない', () => {
    setup()

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('総数と表示中の冊数を出し、続きがあれば「もっと見る」で onLoadMore を呼ぶ', async () => {
    const props = setup({ total: 25, hasMore: true })

    expect(screen.getByText('25冊中 1冊を表示')).toBeVisible()
    await userEvent.click(screen.getByRole('button', { name: 'もっと見る' }))
    expect(props.onLoadMore).toHaveBeenCalledTimes(1)
  })

  it('続きが無ければ「もっと見る」を出さない', () => {
    setup({ hasMore: false })

    expect(
      screen.queryByRole('button', { name: 'もっと見る' }),
    ).not.toBeInTheDocument()
  })

  it('読み込み中は表を出さず読み込み中の表示だけ出す', () => {
    setup({ isLoading: true, books: [] })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('0件ならまだ本が無いことと登録へのリンクを示す', () => {
    setup({ books: [], total: 0 })

    expect(screen.getByText('まだ本が登録されていません')).toBeVisible()
    expect(screen.getByRole('link', { name: /本を登録/ })).toBeVisible()
  })

  it('エラーなら伝え、再試行すると onRetry を呼ぶ', async () => {
    const props = setup({ isError: true, books: [] })

    expect(screen.getByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    await userEvent.click(screen.getByRole('button', { name: '再試行' }))
    expect(props.onRetry).toHaveBeenCalledTimes(1)
  })
})
