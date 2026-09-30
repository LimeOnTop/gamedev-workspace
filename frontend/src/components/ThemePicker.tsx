import { useEffect, useRef, useState } from 'react'
import { applyTheme, loadTheme, THEMES, type Theme } from '../theme'
import { CheckIcon, PaletteIcon } from './Icons'

export default function ThemePicker() {
  const [current, setCurrent] = useState(loadTheme)
  const [open, setOpen] = useState(false)
  const root = useRef<HTMLDivElement>(null)

  useEffect(() => applyTheme(current), [current])

  useEffect(() => {
    if (!open) return
    const onPointer = (e: PointerEvent) => {
      if (!root.current?.contains(e.target as Node)) setOpen(false)
    }
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false)
    document.addEventListener('pointerdown', onPointer)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('pointerdown', onPointer)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  const active = THEMES.find((t) => t.id === current) ?? THEMES[0]

  return (
    <div className="theme-picker" ref={root}>
      <button
        className="theme-trigger"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="listbox"
        aria-expanded={open}
        title="Цветовая тема"
      >
        <PaletteIcon />
        <Swatch theme={active} />
        <span className="theme-trigger-name">{active.name}</span>
      </button>
      {open && (
        <ul className="theme-menu" role="listbox" aria-label="Цветовая тема">
          {THEMES.map((theme) => (
            <li key={theme.id}>
              <button
                role="option"
                aria-selected={theme.id === current}
                className={theme.id === current ? 'active' : ''}
                onClick={() => {
                  setCurrent(theme.id)
                  setOpen(false)
                }}
                // Preview on hover so themes can be compared quickly.
                onMouseEnter={() => (document.documentElement.dataset.theme = theme.id)}
                onMouseLeave={() => (document.documentElement.dataset.theme = current)}
              >
                <Swatch theme={theme} large />
                <span>{theme.name}</span>
                {theme.id === current && <CheckIcon className="theme-check" />}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

function Swatch({ theme, large }: { theme: Theme; large?: boolean }) {
  const [bg, panel, accent] = theme.swatch
  return (
    <span className={`swatch ${large ? 'large' : ''}`} aria-hidden>
      <span style={{ background: bg }} />
      <span style={{ background: panel }} />
      <span style={{ background: accent }} />
    </span>
  )
}
