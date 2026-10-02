import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import Markdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { EditIcon } from './Icons'

interface Props {
  value: string
  onSave: (value: string) => Promise<void>
}

const PLACEHOLDER = `## Цель

Коротко: что описывает документ.

## Правила

- Первое правило
- Второе правило

| Параметр | Значение |
| --- | --- |
| Длительность раунда | 10 мин |

## Открытые вопросы

- …`

/** The structured Markdown description of a scenario file with a table of contents. */
export default function ScenarioDocument({ value, onSave }: Props) {
  const [editing, setEditing] = useState(false)
  const [text, setText] = useState(value)
  const [saving, setSaving] = useState(false)
  const [toc, setToc] = useState<string[]>([])
  const body = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!editing) setText(value)
  }, [value, editing])

  // The contents are read from the rendered document, so they always match its ## headings.
  useLayoutEffect(() => {
    const headings = body.current?.querySelectorAll('h2') ?? []
    setToc(Array.from(headings, (h) => h.textContent ?? ''))
  }, [value, editing])

  const open = () => {
    setText(value)
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

  const scrollTo = (index: number) =>
    body.current?.querySelectorAll('h2')[index]?.scrollIntoView({ behavior: 'smooth', block: 'start' })

  return (
    <div className={`scenario-layout ${!editing && toc.length > 1 ? 'with-toc' : ''}`}>
      <section className="panel scenario-doc">
        <div className="panel-header">
          <h2>Документ</h2>
          {!editing && (
            <div className="panel-actions">
              <button className="btn ghost small" onClick={open}>
                <EditIcon /> {value ? 'Изменить' : 'Добавить'}
              </button>
            </div>
          )}
        </div>
        {editing ? (
          <>
            <textarea
              autoFocus
              className="markdown-editor"
              value={text}
              rows={Math.min(40, Math.max(14, text.split('\n').length + 2))}
              placeholder={PLACEHOLDER}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) save()
                if (e.key === 'Escape') setEditing(false)
              }}
            />
            <div className="edit-actions">
              <span className="hint">Markdown: ## разделы, списки, таблицы · ⌘/Ctrl + Enter — сохранить</span>
              <button className="btn ghost" onClick={() => setEditing(false)} disabled={saving}>Отмена</button>
              <button className="btn primary" onClick={save} disabled={saving}>
                {saving ? 'Сохранение…' : 'Сохранить'}
              </button>
            </div>
          </>
        ) : value ? (
          <div className="markdown" ref={body}>
            <Markdown remarkPlugins={[remarkGfm]}>{value}</Markdown>
          </div>
        ) : (
          <p className="empty-text" onClick={open}>
            Структурированное описание: разделы ##, списки и таблицы — правила, механики, сюжет, UI, открытые вопросы.
          </p>
        )}
      </section>

      {!editing && toc.length > 1 && (
        <nav className="panel scenario-toc" aria-label="Содержание">
          <h2>Содержание</h2>
          <ol>
            {toc.map((title, i) => (
              <li key={i}>
                <button onClick={() => scrollTo(i)}>{title}</button>
              </li>
            ))}
          </ol>
        </nav>
      )}
    </div>
  )
}
