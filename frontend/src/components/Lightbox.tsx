import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { formatSize, isImage, isVideo, type Reference } from '../api'
import { ChevronIcon, CloseIcon, ExternalIcon, FileIcon, TrashIcon } from './Icons'

interface Props {
  references: Reference[]
  index: number
  onIndexChange: (index: number) => void
  onClose: () => void
  onDelete: (reference: Reference) => void
}

const SWIPE_THRESHOLD = 50

export default function Lightbox({ references, index, onIndexChange, onClose, onDelete }: Props) {
  const current = references[index]
  const count = references.length
  const [zoomed, setZoomed] = useState(false)
  const stage = useRef<HTMLDivElement>(null)
  const touchStart = useRef<{ x: number; y: number } | null>(null)
  const filmstrip = useRef<HTMLDivElement>(null)

  const go = (delta: number) => onIndexChange((index + delta + count) % count)

  useEffect(() => setZoomed(false), [index])

  // Keyboard navigation; handlers are refreshed each render so they see current state.
  const keyHandler = useRef<(e: KeyboardEvent) => void>(() => {})
  keyHandler.current = (e: KeyboardEvent) => {
    // A confirmation dialog (e.g. delete) on top of the lightbox owns the keyboard.
    if (document.querySelector('.modal-backdrop')) return
    if (e.key === 'Escape') {
      if (zoomed) setZoomed(false)
      else onClose()
    } else if (e.key === 'ArrowRight') go(1)
    else if (e.key === 'ArrowLeft') go(-1)
    else if (e.key === 'Home') onIndexChange(0)
    else if (e.key === 'End') onIndexChange(count - 1)
    else if ((e.key === 'z' || e.key === ' ') && isImage(current)) {
      e.preventDefault()
      setZoomed((z) => !z)
    } else return
    e.preventDefault()
  }

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => keyHandler.current(e)
    window.addEventListener('keydown', onKey)
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = overflow
    }
  }, [])

  // Preload neighbours so flipping through feels instant.
  useEffect(() => {
    for (const delta of [1, -1]) {
      const ref = references[(index + delta + count) % count]
      if (ref && isImage(ref)) new Image().src = ref.url
    }
  }, [index, references, count])

  useEffect(() => {
    filmstrip.current?.querySelector('.active')?.scrollIntoView({ block: 'nearest', inline: 'center', behavior: 'smooth' })
  }, [index])

  const toggleZoom = (e: React.MouseEvent<HTMLImageElement>) => {
    e.stopPropagation()
    const img = e.currentTarget
    const rect = img.getBoundingClientRect()
    const fx = (e.clientX - rect.left) / rect.width
    const fy = (e.clientY - rect.top) / rect.height
    const willZoom = !zoomed
    setZoomed(willZoom)
    if (willZoom) {
      // Center the zoomed view on the clicked point.
      requestAnimationFrame(() => {
        const el = stage.current
        if (!el) return
        el.scrollLeft = img.naturalWidth * fx - el.clientWidth / 2
        el.scrollTop = img.naturalHeight * fy - el.clientHeight / 2
      })
    }
  }

  if (!current) return null

  return createPortal(
    <div className="lightbox" role="dialog" aria-modal="true" aria-label="Просмотр референсов">
      <div className="lightbox-bar">
        <span className="lightbox-counter">{index + 1} / {count}</span>
        <span className="lightbox-name" title={current.original_name}>{current.original_name}</span>
        <span className="lightbox-meta">{formatSize(current.size)}</span>
        <div className="lightbox-actions">
          <a className="lb-btn" href={current.url} target="_blank" rel="noreferrer" title="Открыть оригинал">
            <ExternalIcon />
          </a>
          <button className="lb-btn danger" title="Удалить" onClick={() => onDelete(current)}>
            <TrashIcon />
          </button>
          <button className="lb-btn" title="Закрыть (Esc)" onClick={onClose}>
            <CloseIcon />
          </button>
        </div>
      </div>

      <div
        ref={stage}
        className={`lightbox-stage ${zoomed ? 'zoomed' : ''}`}
        onClick={() => (zoomed ? setZoomed(false) : onClose())}
        onTouchStart={(e) => {
          if (zoomed) return
          touchStart.current = { x: e.touches[0].clientX, y: e.touches[0].clientY }
        }}
        onTouchEnd={(e) => {
          const start = touchStart.current
          touchStart.current = null
          if (!start || zoomed) return
          const dx = e.changedTouches[0].clientX - start.x
          const dy = e.changedTouches[0].clientY - start.y
          if (Math.abs(dx) > SWIPE_THRESHOLD && Math.abs(dx) > Math.abs(dy)) go(dx < 0 ? 1 : -1)
        }}
      >
        {isImage(current) ? (
          <img
            key={current.id}
            src={current.url}
            alt={current.original_name}
            className="lightbox-media"
            onClick={toggleZoom}
            draggable={false}
          />
        ) : isVideo(current) ? (
          <video key={current.id} src={current.url} className="lightbox-media" controls autoPlay onClick={(e) => e.stopPropagation()} />
        ) : (
          <a className="lightbox-file" href={current.url} target="_blank" rel="noreferrer" onClick={(e) => e.stopPropagation()}>
            <FileIcon size={56} />
            <strong>{current.original_name}</strong>
            <span>Открыть файл</span>
          </a>
        )}
      </div>

      {count > 1 && (
        <>
          <button className="lightbox-nav prev" title="Назад (←)" onClick={() => go(-1)}>
            <ChevronIcon size={28} style={{ transform: 'rotate(180deg)' }} />
          </button>
          <button className="lightbox-nav next" title="Вперёд (→)" onClick={() => go(1)}>
            <ChevronIcon size={28} />
          </button>
          <div className="lightbox-filmstrip" ref={filmstrip}>
            {references.map((r, i) => (
              <button
                key={r.id}
                className={`lb-thumb ${i === index ? 'active' : ''}`}
                onClick={() => onIndexChange(i)}
                title={r.original_name}
              >
                {isImage(r) ? <img src={r.url} alt="" loading="lazy" /> : <FileIcon size={20} />}
              </button>
            ))}
          </div>
        </>
      )}
    </div>,
    document.body,
  )
}
