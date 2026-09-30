import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { ConfirmDialog } from './ConfirmDialog'

function renderDialog(props: Partial<Parameters<typeof ConfirmDialog>[0]>) {
  const onConfirm = vi.fn()
  const onCancel = vi.fn()
  const view = render(
    <ConfirmDialog
      open
      title="この本を削除しますか？"
      description="元に戻せません。"
      confirmLabel="削除する"
      onConfirm={onConfirm}
      onCancel={onCancel}
      {...props}
    />,
  )
  return { onConfirm, onCancel, ...view }
}

describe('ConfirmDialog', () => {
  it('open が false なら何も見せない', () => {
    renderDialog({ open: false })

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('open なら題名と説明を見せ、キャンセルにフォーカスを置く', () => {
    renderDialog({})

    const dialog = screen.getByRole('dialog', {
      name: 'この本を削除しますか？',
    })
    expect(dialog).toHaveTextContent('元に戻せません。')
    expect(screen.getByRole('button', { name: 'キャンセル' })).toHaveFocus()
  })

  it('説明文を aria-describedby で結び、開いたときに読み上げられるようにする', () => {
    renderDialog({})

    expect(
      screen.getByRole('dialog', { name: 'この本を削除しますか？' }),
    ).toHaveAccessibleDescription('元に戻せません。')
  })

  it('確定を押すと onConfirm だけが呼ばれる', async () => {
    const { onConfirm, onCancel } = renderDialog({})

    await userEvent.click(screen.getByRole('button', { name: '削除する' }))

    expect(onConfirm).toHaveBeenCalledTimes(1)
    expect(onCancel).not.toHaveBeenCalled()
  })

  it('キャンセルを押すと onCancel だけが呼ばれる', async () => {
    const { onConfirm, onCancel } = renderDialog({})

    await userEvent.click(screen.getByRole('button', { name: 'キャンセル' }))

    expect(onCancel).toHaveBeenCalledTimes(1)
    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('Esc を押すと onCancel だけが呼ばれる', async () => {
    const { onConfirm, onCancel } = renderDialog({})

    await userEvent.keyboard('{Escape}')

    expect(onCancel).toHaveBeenCalledTimes(1)
    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('busy の間は確定を押せない', async () => {
    const { onConfirm } = renderDialog({ busy: true })

    const confirm = screen.getByRole('button', { name: '削除する' })
    expect(confirm).toBeDisabled()
    await userEvent.click(confirm)

    expect(onConfirm).not.toHaveBeenCalled()
  })

  it('open が true から false になると閉じる', () => {
    const { rerender, onConfirm, onCancel } = renderDialog({})

    rerender(
      <ConfirmDialog
        open={false}
        title="この本を削除しますか？"
        description="元に戻せません。"
        confirmLabel="削除する"
        onConfirm={onConfirm}
        onCancel={onCancel}
      />,
    )

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
