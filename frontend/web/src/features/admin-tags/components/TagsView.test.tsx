import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { TagsView } from './TagsView'

const items = [
  { id: 1, name: 'データ', version: 1, bookCount: 2 },
  { id: 2, name: '設計', version: 3, bookCount: 0 },
]

describe('TagsView', () => {
  it('タグの一覧を名前と冊数で表示する', () => {
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={vi.fn()}
        onDelete={vi.fn()}
      />,
    )

    expect(screen.getByText('データ')).toBeVisible()
    expect(screen.getByText('2冊')).toBeVisible()
    expect(screen.getByText('設計')).toBeVisible()
    expect(screen.getByText('0冊')).toBeVisible()
  })

  it('タグ名を入れて追加すると onAdd が呼ばれ、成功すると入力欄が空になる', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn().mockResolvedValue({ ok: true })
    render(
      <TagsView
        items={items}
        onAdd={onAdd}
        onRename={vi.fn()}
        onDelete={vi.fn()}
      />,
    )

    const input = screen.getByLabelText('タグを追加', { exact: false })
    await user.type(input, 'ネットワーク')
    await user.click(screen.getByRole('button', { name: '追加' }))

    expect(onAdd).toHaveBeenCalledWith('ネットワーク')
    expect(
      await screen.findByLabelText('タグを追加', { exact: false }),
    ).toHaveValue('')
  })

  it('空白だけの入力は追加を呼ばない', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn()
    render(
      <TagsView
        items={items}
        onAdd={onAdd}
        onRename={vi.fn()}
        onDelete={vi.fn()}
      />,
    )

    await user.type(
      screen.getByLabelText('タグを追加', { exact: false }),
      '   ',
    )
    await user.click(screen.getByRole('button', { name: '追加' }))

    expect(onAdd).not.toHaveBeenCalled()
  })

  it('追加が失敗すれば文言を表示し、入力は残す', async () => {
    const user = userEvent.setup()
    const onAdd = vi.fn().mockResolvedValue({
      ok: false,
      message: '「設計」というタグはすでにあります',
    })
    render(
      <TagsView
        items={items}
        onAdd={onAdd}
        onRename={vi.fn()}
        onDelete={vi.fn()}
      />,
    )

    await user.type(
      screen.getByLabelText('タグを追加', { exact: false }),
      '設計',
    )
    await user.click(screen.getByRole('button', { name: '追加' }))

    expect(
      await screen.findByText('「設計」というタグはすでにあります'),
    ).toBeVisible()
    expect(screen.getByLabelText('タグを追加', { exact: false })).toHaveValue(
      '設計',
    )
  })

  it('「名前を変更」でその場の入力欄になり、保存すると onRename が呼ばれる', async () => {
    const user = userEvent.setup()
    const onRename = vi.fn().mockResolvedValue({ ok: true })
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={onRename}
        onDelete={vi.fn()}
      />,
    )

    await user.click(screen.getAllByRole('button', { name: '名前を変更' })[0])
    const input = screen.getByLabelText('「データ」の新しい名前')
    await user.clear(input)
    await user.type(input, 'データ分析')
    await user.click(screen.getByRole('button', { name: '保存' }))

    expect(onRename).toHaveBeenCalledWith(1, 1, 'データ分析')
  })

  it('その場の入力欄で「キャンセル」を押すと onRename を呼ばずに元の表示へ戻る', async () => {
    const user = userEvent.setup()
    const onRename = vi.fn()
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={onRename}
        onDelete={vi.fn()}
      />,
    )

    await user.click(screen.getAllByRole('button', { name: '名前を変更' })[0])
    await user.click(screen.getByRole('button', { name: 'キャンセル' }))

    expect(onRename).not.toHaveBeenCalled()
    expect(
      screen.queryByLabelText('「データ」の新しい名前'),
    ).not.toBeInTheDocument()
  })

  it('名前の変更が先に更新されていたときの文言を表示する', async () => {
    const user = userEvent.setup()
    const onRename = vi.fn().mockResolvedValue({
      ok: false,
      stale: true,
      message:
        'このタグは、ほかの画面で先に更新されていました。最新の内容を読み込みました。',
    })
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={onRename}
        onDelete={vi.fn()}
      />,
    )

    await user.click(screen.getAllByRole('button', { name: '名前を変更' })[0])
    await user.click(screen.getByRole('button', { name: '保存' }))

    expect(
      await screen.findByText(
        'このタグは、ほかの画面で先に更新されていました。最新の内容を読み込みました。',
      ),
    ).toBeVisible()
    expect(
      screen.queryByLabelText('「データ」の新しい名前'),
    ).not.toBeInTheDocument()
  })

  it('「削除」は確認ダイアログを出し、確定したときだけ onDelete を呼ぶ', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn().mockResolvedValue({ ok: true })
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={vi.fn()}
        onDelete={onDelete}
      />,
    )

    await user.click(screen.getAllByRole('button', { name: '削除' })[0])
    expect(
      screen.getByText('付いている2冊から外れます。本は消えません。'),
    ).toBeVisible()
    await user.click(screen.getByRole('button', { name: '削除する' }))

    expect(onDelete).toHaveBeenCalledWith(1)
  })

  it('確認ダイアログをキャンセルすると onDelete を呼ばない', async () => {
    const user = userEvent.setup()
    const onDelete = vi.fn()
    render(
      <TagsView
        items={items}
        onAdd={vi.fn()}
        onRename={vi.fn()}
        onDelete={onDelete}
      />,
    )

    await user.click(screen.getAllByRole('button', { name: '削除' })[0])
    await user.click(screen.getByRole('button', { name: 'キャンセル' }))

    expect(onDelete).not.toHaveBeenCalled()
  })
})
