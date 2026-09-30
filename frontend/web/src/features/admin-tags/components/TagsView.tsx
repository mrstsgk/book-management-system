import { useState } from 'react'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { FormField } from '@/components/form/FormField'
import { describedBy } from '@/components/form/describedBy'
import type { TagWithCount } from '../utils/mergeTagCounts'

type Outcome = { ok: true } | { ok: false; message: string; stale?: boolean }

type TagsViewProps = {
  items: TagWithCount[]
  onAdd: (name: string) => Promise<Outcome>
  onRename: (id: number, version: number, name: string) => Promise<Outcome>
  onDelete: (id: number) => Promise<Outcome>
}

export function TagsView({ items, onAdd, onRename, onDelete }: TagsViewProps) {
  const [newName, setNewName] = useState('')
  const [addError, setAddError] = useState('')
  const [editing, setEditing] = useState<{ id: number; draft: string } | null>(
    null,
  )
  const [editError, setEditError] = useState('')
  const [notice, setNotice] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<TagWithCount | null>(null)
  const [deleteBusy, setDeleteBusy] = useState(false)

  const submitAdd = async () => {
    const trimmed = newName.trim()
    if (!trimmed) return
    const outcome = await onAdd(trimmed)
    if (outcome.ok) {
      setNewName('')
      setAddError('')
    } else {
      setAddError(outcome.message)
    }
  }

  const startEdit = (tag: TagWithCount) => {
    setEditing({ id: tag.id, draft: tag.name })
    setEditError('')
  }

  const submitEdit = async (tag: TagWithCount) => {
    if (!editing) return
    const trimmed = editing.draft.trim()
    if (!trimmed) return
    const outcome = await onRename(tag.id, tag.version, trimmed)
    if (outcome.ok) {
      setEditing(null)
      setEditError('')
    } else if (outcome.stale) {
      setEditing(null)
      setEditError('')
      setNotice(outcome.message)
    } else {
      setEditError(outcome.message)
    }
  }

  const confirmDelete = async () => {
    if (!deleteTarget) return
    setDeleteBusy(true)
    const outcome = await onDelete(deleteTarget.id)
    setDeleteBusy(false)
    setDeleteTarget(null)
    if (!outcome.ok) setNotice(outcome.message)
  }

  return (
    <div className="flex max-w-[760px] flex-col gap-5">
      {notice && (
        <div
          role="alert"
          className="rounded-lg border border-brand-700 bg-brand-50 px-5 py-3.5 text-sm text-brand-800"
        >
          {notice}
        </div>
      )}

      <form
        onSubmit={(e) => {
          e.preventDefault()
          void submitAdd()
        }}
        className="card flex flex-col gap-1.5 px-6 py-5"
      >
        <FormField
          id="new-tag"
          label="タグを追加"
          error={addError}
          hint="1〜30文字"
        >
          <div className="flex gap-2">
            <input
              id="new-tag"
              value={newName}
              aria-describedby={describedBy('new-tag')}
              aria-invalid={addError ? true : undefined}
              onChange={(e) => {
                setNewName(e.target.value)
                setAddError('')
              }}
              className="h-12 flex-grow rounded-lg border border-ink-400 px-4 text-base"
            />
            <button
              type="submit"
              className="h-12 rounded-lg bg-brand-700 px-6 font-bold text-white"
            >
              追加
            </button>
          </div>
        </FormField>
      </form>

      <ul className="card flex flex-col">
        {items.map((tag) => (
          <li
            key={tag.id}
            className="flex items-center gap-3 border-t border-ink-200 px-6 py-3.5 first:border-t-0"
          >
            {editing?.id === tag.id ? (
              <form
                onSubmit={(e) => {
                  e.preventDefault()
                  void submitEdit(tag)
                }}
                className="flex flex-grow flex-col gap-1.5"
              >
                <label
                  htmlFor={`rename-${tag.id}`}
                  className="sr-only"
                >{`「${tag.name}」の新しい名前`}</label>
                <div className="flex items-center gap-3">
                  <input
                    id={`rename-${tag.id}`}
                    value={editing.draft}
                    onChange={(e) =>
                      setEditing({ id: tag.id, draft: e.target.value })
                    }
                    className="h-11 flex-grow rounded-lg border border-ink-400 px-3 text-base"
                  />
                  <button
                    type="submit"
                    className="h-11 rounded-lg bg-brand-700 px-4 font-bold text-white"
                  >
                    保存
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setEditing(null)
                      setEditError('')
                    }}
                    className="h-11 px-3 text-ink-600"
                  >
                    キャンセル
                  </button>
                </div>
                {editError && (
                  <p className="text-[13px] text-brand-800">{editError}</p>
                )}
              </form>
            ) : (
              <>
                <span className="flex-grow font-bold">{tag.name}</span>
                <span className="w-16 text-right text-[13px] text-ink-600">
                  {tag.bookCount}冊
                </span>
                <button
                  type="button"
                  onClick={() => startEdit(tag)}
                  className="h-11 px-3 font-bold text-brand-800"
                >
                  名前を変更
                </button>
                <button
                  type="button"
                  onClick={() => setDeleteTarget(tag)}
                  className="h-11 px-3 font-bold text-brand-800"
                >
                  削除
                </button>
              </>
            )}
          </li>
        ))}
      </ul>

      <ConfirmDialog
        open={deleteTarget !== null}
        title={`「${deleteTarget?.name}」を削除しますか？`}
        description={`付いている${deleteTarget?.bookCount}冊から外れます。本は消えません。`}
        confirmLabel="削除する"
        busy={deleteBusy}
        onConfirm={() => void confirmDelete()}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  )
}
