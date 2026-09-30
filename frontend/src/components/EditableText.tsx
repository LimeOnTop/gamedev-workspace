import { useEffect, useState, type ReactNode } from 'react'
import { EditIcon } from './Icons'

interface Props {
  label: ReactNode
  value: string
  placeholder: string
  onSave: (value: string) => Promise<void>
  /** Extra header buttons shown while not editing. */
  actions?: ReactNode
  /** When set, a "draft" button opens the editor pre-filled with the generated text. */
  draft?: { label: string; build: () => string }
  hint?: ReactNode
  className?: string
}

export default function EditableText({ label, value, placeholder, onSave, actions, draft, hint, className }: Props) {
  const [editing, setEditing] = useState(false)
  const [text, setText] = useState(value)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!editing) setText(value)
  }, [value, editing])

  const open = (initial = value) => {
    setText(initial)
    setEditing(true)
  }

  const save = async () => {
    setSaving(true)
    try {
      await onSave(text)
      setEditing(false)
    } catch {
      // error already shown by the caller
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className={`editable ${className ?? ''}`}>
      <div className="panel-header">
        <h2>{label}</h2>
        {!editing && (
          <div className="panel-actions">
            {actions}
            {draft && (
              <button className="btn ghost small" onClick={() => open(draft.build())} title="Заполнить редактор черновиком">
                {draft.label}
              </button>
            )}
            <button className="btn ghost small" onClick={() => open()}>
              <EditIcon /> {value ? 'Изменить' : 'Добавить'}
            </button>
          </div>
        )}
      </div>
      {hint && !editing && <p className="panel-hint">{hint}</p>}
      {editing ? (
        <>
          <textarea
            autoFocus
            value={text}
            rows={Math.min(16, Math.max(4, text.split('\n').length + 1))}
            placeholder={placeholder}
            onChange={(e) => setText(e.target.value)}
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
        <p className="empty-text" onClick={() => open()}>{placeholder}</p>
      )}
    </div>
  )
}
