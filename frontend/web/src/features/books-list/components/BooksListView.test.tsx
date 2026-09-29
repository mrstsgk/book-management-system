import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { ComponentProps } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { BooksListView } from './BooksListView'

type Props = ComponentProps<typeof BooksListView>

const book = {
  id: 7,
  title: 'データ指向アプリケーションデザイン 第2版',
  authors: 'Martin Kleppmann',
  summary: '分散システムの原理を整理した本',
  rating: 5,
  tags: ['データ', '設計'],
  coverUrl: 'https://cover.openbd.jp/9784814401802.jpg',
}

function setupView(overrides: Partial<Props> = {}) {
  const props: Props = {
    books: [book],
    total: 1,
    tags: [
      { id: 1, name: 'データ' },
      { id: 2, name: '設計' },
    ],
    hasFilter: false,
    isLoading: false,
    isError: false,
    onRetry: vi.fn(),
    hasMore: false,
    isLoadingMore: false,
    onLoadMore: vi.fn(),
    onSearch: vi.fn(),
    onSelectTag: vi.fn(),
    onClear: vi.fn(),
    ...overrides,
  }
  renderWithProviders(<BooksListView {...props} />)
  return props
}

describe('BooksListView', () => {
  it('本ごとに書影・書名・著者・一言まとめ・評価・分野タグを並べ、詳細へリンクする', () => {
    setupView()

    const card = screen.getByRole('link', {
      name: /データ指向アプリケーションデザイン/,
    })
    expect(card).toHaveAttribute('href', '/books/7')
    expect(within(card).getByRole('heading', { level: 2 })).toHaveTextContent(
      'データ指向アプリケーションデザイン 第2版',
    )
    expect(within(card).getByText('Martin Kleppmann')).toBeVisible()
    expect(
      within(card).getByText('分散システムの原理を整理した本'),
    ).toBeVisible()
    expect(within(card).getByRole('img', { name: '評価 5 / 5' })).toBeVisible()
    expect(within(card).getByText('データ')).toBeVisible()
    expect(within(card).getByText('設計')).toBeVisible()
    // 書影は飾り（alt=""）なので presentation として問い合わせる
    expect(within(card).getByRole('presentation')).toHaveAttribute(
      'src',
      book.coverUrl,
    )
  })

  it('Google Books の書影の本は、カードの外にその本の Google Books へのリンクを出し、「Powered by Google」を添える', () => {
    setupView({
      books: [
        {
          ...book,
          coverUrl: 'https://books.google.com/books/content?id=a',
          coverSource: 'googlebooks',
          coverPageUrl: 'https://books.google.co.jp/books?id=a',
        },
      ],
    })

    const google = screen.getByRole('link', { name: /Google Books/ })
    expect(google).toHaveAttribute(
      'href',
      'https://books.google.co.jp/books?id=a',
    )
    expect(google).toHaveAttribute('target', '_blank')
    expect(google).toHaveAttribute('rel', 'noopener noreferrer')
    // リンクの入れ子にしない（カードのリンクの中に置くと、どちらに移るか定まらないため）
    const card = screen.getByRole('link', { name: /データ指向/ })
    expect(card).not.toContainElement(google)
    expect(screen.getByText('Powered by Google')).toBeVisible()
  })

  it('openBD の書影だけなら Google の表示を出さない', () => {
    setupView()

    expect(
      screen.queryByRole('link', { name: /Google Books/ }),
    ).not.toBeInTheDocument()
    expect(screen.queryByText('Powered by Google')).not.toBeInTheDocument()
  })

  it('書影が無い本は「書影なし」の枠を出す', () => {
    setupView({ books: [{ ...book, coverUrl: undefined }] })

    const card = screen.getByRole('link', { name: /データ指向/ })
    expect(within(card).getByText('書影なし')).toBeVisible()
    expect(within(card).queryByRole('presentation')).not.toBeInTheDocument()
  })

  it('総数と表示中の冊数を出し、続きがあれば「もっと見る」で onLoadMore を呼ぶ', async () => {
    const props = setupView({ total: 45, hasMore: true })

    expect(screen.getByText('45冊中 1冊を表示')).toBeVisible()
    await userEvent.click(screen.getByRole('button', { name: 'もっと見る' }))
    expect(props.onLoadMore).toHaveBeenCalledTimes(1)
  })

  it('続きが無ければ「もっと見る」を出さない', () => {
    setupView({ hasMore: false })

    expect(
      screen.queryByRole('button', { name: 'もっと見る' }),
    ).not.toBeInTheDocument()
  })

  it('続きの読み込み中は末尾に知らせ、「もっと見る」を押せなくする', () => {
    setupView({ total: 45, hasMore: true, isLoadingMore: true })

    expect(screen.getByRole('status')).toHaveTextContent('続きを読み込み中…')
    expect(screen.getByRole('button', { name: 'もっと見る' })).toBeDisabled()
  })

  it('キーワードを送信すると入力した文字で onSearch を呼ぶ', async () => {
    const props = setupView({ query: '設計' })

    const input = screen.getByRole('searchbox', {
      name: 'キーワード（書名・著者）',
    })
    expect(input).toHaveValue('設計')
    await userEvent.clear(input)
    await userEvent.type(input, 'Kleppmann')
    await userEvent.click(screen.getByRole('button', { name: '検索' }))

    expect(props.onSearch).toHaveBeenCalledWith('Kleppmann')
  })

  it('選んでいるタグが押された状態になり、タグで onSelectTag、「すべて」で undefined を渡す', async () => {
    const props = setupView({ tagId: 2, hasFilter: true })

    const group = screen.getByRole('group', { name: '分野タグ' })
    expect(within(group).getByRole('button', { name: '設計' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
    expect(
      within(group).getByRole('button', { name: 'すべて' }),
    ).toHaveAttribute('aria-pressed', 'false')

    await userEvent.click(within(group).getByRole('button', { name: 'データ' }))
    expect(props.onSelectTag).toHaveBeenLastCalledWith(1)
    await userEvent.click(within(group).getByRole('button', { name: 'すべて' }))
    expect(props.onSelectTag).toHaveBeenLastCalledWith(undefined)
  })

  it('絞り込みが無ければ「すべて」が押された状態', () => {
    setupView()

    expect(screen.getByRole('button', { name: 'すべて' })).toHaveAttribute(
      'aria-pressed',
      'true',
    )
  })

  it('タグ一覧が無い（取れなかった）ときはタグのボタンを出さないが、本は出す', () => {
    setupView({ tags: undefined })

    expect(
      screen.queryByRole('group', { name: '分野タグ' }),
    ).not.toBeInTheDocument()
    expect(screen.getByRole('link', { name: /データ指向/ })).toBeVisible()
  })

  it('最初の読み込み中は読み込み中を表示する', () => {
    setupView({ isLoading: true, books: [] })

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
  })

  it('エラーなら分かる言葉で伝え、再試行で onRetry を呼ぶ', async () => {
    const props = setupView({ isError: true, books: [] })

    expect(screen.getByRole('alert')).toHaveTextContent(
      '本の一覧を読み込めませんでした',
    )
    await userEvent.click(screen.getByRole('button', { name: '再試行' }))
    expect(props.onRetry).toHaveBeenCalledTimes(1)
  })

  it('本が1冊も無いときは、まだ登録されていないことを伝える', () => {
    setupView({ books: [], total: 0 })

    expect(screen.getByText('まだ本が登録されていません')).toBeVisible()
    expect(
      screen.queryByRole('button', { name: '検索と絞り込みを外す' }),
    ).not.toBeInTheDocument()
  })

  it('検索・絞り込みの結果が無いときは、条件の外し方を示し、ボタンで onClear を呼ぶ', async () => {
    const props = setupView({
      books: [],
      total: 0,
      hasFilter: true,
      query: 'Kubernetes',
    })

    expect(screen.getByText('条件に当てはまる本はありません')).toBeVisible()
    expect(
      screen.getByText(
        'キーワードを変えるか、分野タグを「すべて」に戻してください。',
      ),
    ).toBeVisible()
    await userEvent.click(
      screen.getByRole('button', { name: '検索と絞り込みを外す' }),
    )
    expect(props.onClear).toHaveBeenCalledTimes(1)
  })
})
