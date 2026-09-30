import { useState } from 'react'
import type { Characteristic } from '../api'
import { CloseIcon, EditIcon, PlusIcon } from './Icons'

interface Props {
  value: Characteristic[]
  onSave: (value: Characteristic[]) => Promise<void>
}

export default function CharacteristicsEditor({ value, onSave }: Props) {
  const [editing, setEditing] = useState(false)
  const [rows, setRows] = useState<Characteristic[]>([])
  const [saving, setSaving] = useState(false)

  const start = () => {
    setRows(value.length ? value.map((c) => ({ ...c })) : [{ key: '', value: '' }])
    setEditing(true)
  }

  const update = (i: number, field: keyof Characteristic, v: string) =>
    setRows((prev) => prev.map((r, idx) => (idx === i ? { ...r, [field]: v } : r)))

  const save = async () => {
    setSaving(true)
    try {
      await onSave(rows.filter((r) => r.key.trim() || r.value.trim()))
      setEditing(false)
    } catch {
      // error already shown by the caller
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <div className="panel-header">
        <h2>Характеристики</h2>
        {!editing && (
          <button className="btn ghost small" onClick={start}>
            <EditIcon /> {value.length ? 'Изменить' : 'Добавить'}
          </button>
        )}
      </div>

      {editing ? (
        <>
          <div className="chars-edit">
            {rows.map((row, i) => (
              <div className="chars-edit-row" key={i}>
                <input
                  placeholder="Параметр"
                  value={row.key}
                  autoFocus={i === rows.length - 1 && !row.key}
                  onChange={(e) => update(i, 'key', e.target.value)}
                />
                <input placeholder="Значение" value={row.value} onChange={(e) => update(i, 'value', e.target.value)} />
                <button
                  className="icon-btn danger"
                  title="Удалить строку"
                  onClick={() => setRows((prev) => prev.filter((_, idx) => idx !== i))}
                >
                  <CloseIcon />
                </button>
              </div>
            ))}
          </div>
          <div className="edit-actions">
            <button className="btn ghost small add-row" onClick={() => setRows((prev) => [...prev, { key: '', value: '' }])}>
              <PlusIcon /> Строка
            </button>
            <button className="btn ghost" onClick={() => setEditing(false)} disabled={saving}>Отмена</button>
            <button className="btn primary" onClick={save} disabled={saving}>{saving ? 'Сохранение…' : 'Сохранить'}</button>
          </div>
        </>
      ) : value.length ? (
        <dl className="chars">
          {value.map((c, i) => (
            <div className="chars-row" key={i}>
              <dt>{c.key}</dt>
              <dd>{c.value}</dd>
            </div>
          ))}
        </dl>
      ) : (
        <p className="empty-text" onClick={start}>Урон, размер, вес, скорость — любые параметры объекта.</p>
      )}
    </div>
  )
}
