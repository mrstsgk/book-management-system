import { screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { BookResponse } from '@/api/generated/api.schemas'
import { renderWithProviders } from '@/testing/render'
import { BookDetailView } from './BookDetailView'

const openbdBook: BookResponse = {
  id: 1,
  isbn: '9784297146221',
  title: '良いコード／悪いコードで学ぶ設計入門',
  authors: '仙塲大也',
  publisher: '技術評論社',
  publishedOn: '202405',
  coverUrl: 'https://cover.openbd.jp/9784297146221.jpg',
  coverSource: 'openbd',
  amazonUrl: 'https://www.amazon.co.jp/dp/4297146223',
  summary: '設計の悪い例と良い例を並べて学べる',
  comment: '1段落目\n2段落目',
  rating: 4,
  tags: ['設計', 'リファクタリング'],
}

describe('BookDetailView', () => {
  it('書誌・書影・評価・分野タグ・一言まとめ・感想を見せる', () => {
    renderWithProviders(<BookDetailView book={openbdBook} />)

    expect(
      screen.getByRole('heading', {
        level: 1,
        name: '良いコード／悪いコードで学ぶ設計入門',
      }),
    ).toBeVisible()
    expect(
      screen.getByRole('img', {
        name: '良いコード／悪いコードで学ぶ設計入門の書影',
      }),
    ).toHaveAttribute('src', 'https://cover.openbd.jp/9784297146221.jpg')
    expect(screen.getByRole('img', { name: '評価 4 / 5' })).toBeVisible()
    expect(screen.getByText('仙塲大也')).toBeVisible()
    expect(screen.getByText('技術評論社')).toBeVisible()
    expect(screen.getByText('202405')).toBeVisible()
    expect(screen.getByText('9784297146221')).toBeVisible()
    expect(screen.getByText('設計')).toBeVisible()
    expect(screen.getByText('リファクタリング')).toBeVisible()
    expect(screen.getByText('設計の悪い例と良い例を並べて学べる')).toBeVisible()
  })

  it('感想の改行をそのまま保つ', () => {
    renderWithProviders(<BookDetailView book={openbdBook} />)

    const comment = screen.getByText(/1段落目/)
    expect(comment.textContent).toBe('1段落目\n2段落目')
    expect(comment).toHaveClass('whitespace-pre-wrap')
  })

  it('出版社・出版日が無ければその行を出さない', () => {
    renderWithProviders(
      <BookDetailView
        book={{ ...openbdBook, publisher: undefined, publishedOn: '' }}
      />,
    )

    expect(screen.queryByText('出版社')).not.toBeInTheDocument()
    expect(screen.queryByText('出版日')).not.toBeInTheDocument()
    expect(screen.getByText('ISBN')).toBeVisible()
  })

  it('書影が無ければ「書影なし」の枠を出す', () => {
    renderWithProviders(
      <BookDetailView
        book={{ ...openbdBook, coverUrl: undefined, coverSource: undefined }}
      />,
    )

    expect(screen.queryByRole('img', { name: /書影$/ })).not.toBeInTheDocument()
    expect(screen.getByText('書影なし')).toBeVisible()
  })

  it('Amazon の URL があれば新しいタブで開くリンクを出す', () => {
    renderWithProviders(<BookDetailView book={openbdBook} />)

    const link = screen.getByRole('link', { name: /Amazonで見る/ })
    expect(link).toHaveAttribute(
      'href',
      'https://www.amazon.co.jp/dp/4297146223',
    )
    expect(link).toHaveAttribute('target', '_blank')
    expect(link).toHaveAttribute('rel', 'noopener noreferrer')
  })

  it('Amazon の URL が無ければ Amazon のリンクを出さない', () => {
    renderWithProviders(
      <BookDetailView book={{ ...openbdBook, amazonUrl: undefined }} />,
    )

    expect(
      screen.queryByRole('link', { name: /Amazonで見る/ }),
    ).not.toBeInTheDocument()
  })

  it('一覧へ戻るリンクを出す', () => {
    renderWithProviders(<BookDetailView book={openbdBook} />)

    expect(screen.getByRole('link', { name: /一覧へ戻る/ })).toHaveAttribute(
      'href',
      '/',
    )
  })
})
