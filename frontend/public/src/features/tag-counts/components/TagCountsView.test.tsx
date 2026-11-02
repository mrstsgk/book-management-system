import { screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { renderWithProviders } from '@/testing/render'
import { TagCountsView } from './TagCountsView'

describe('TagCountsView', () => {
  it('渡された順に分野を並べ、各行をその分野で絞り込んだ一覧へのリンクにする', () => {
    renderWithProviders(
      <TagCountsView
        items={[
          { id: 3, name: '設計', bookCount: 5 },
          { id: 1, name: 'クラウド', bookCount: 2 },
        ]}
      />,
    )

    const links = within(screen.getByRole('list')).getAllByRole('link')
    expect(links).toHaveLength(2)
    expect(links[0]).toHaveTextContent('設計5冊')
    expect(links[0]).toHaveAttribute('href', '/?tag=3')
    expect(links[1]).toHaveTextContent('クラウド2冊')
    expect(links[1]).toHaveAttribute('href', '/?tag=1')
  })
})
