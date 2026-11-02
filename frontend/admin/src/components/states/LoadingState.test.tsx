import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { LoadingState } from './LoadingState'

describe('LoadingState', () => {
  it('既定では「読み込み中…」を状態として伝える', () => {
    render(<LoadingState />)

    expect(screen.getByRole('status')).toHaveTextContent('読み込み中…')
  })

  it('文言を差し替えられる', () => {
    render(<LoadingState label="続きを読み込み中…" />)

    expect(screen.getByRole('status')).toHaveTextContent('続きを読み込み中…')
  })
})
