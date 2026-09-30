import { useEffect, useState } from 'react'
import { EditIcon } from './Icons'

interface Props {
  label: string
  value: string
  placeholder: string
  onSave: (value: string) => Promise<void>
}

export default function EditableText({ label, value, placeholder, onSave }: Props) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(value)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!editing) setDraft(value)
  }, [value, editing])

  const save = async () => {
    setSaving(true)
    try {
      await onSave(draft)
      setEditing(false)
    } catch {
      // error already shown by the caller
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="editable">
      <div className="panel-header">
        <h2>{label}</h2>
        {!editing && (
          <button className="btn ghost small" onClick={() => setEditing(true)}>
            <EditIcon /> {value ? 'Изменить' : 'Добавить'}
          </button>
        )}
      </div>
      {editing ? (
        <>
          <textarea
            autoFocus
            value={draft}
            rows={Math.min(16, Math.max(4, draft.split('\n').length + 1))}
            placeholder={placeholder}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) save()
              if (e.key === 'Escape') setEditing(false)
            }}
          />
          <div className="edit-actions">
            <span className="hint">⌘/Ctrl + Enter — сохранить</span>
            <button className="btn ghost" onClick={() => setEditing(false)} disabled={saving}>Отмена</button>
            <button className="btn primary" onClick={save} disabled={saving}>{saving ? 'Сохранение…' : 'Сохранить'}</button>
          </div>
        </>
      ) : value ? (
        <p className="text-block">{value}</p>
      ) : (
        <p className="empty-text" onClick={() => setEditing(true)}>{placeholder}</p>
      )}
    </div>
  )
}
