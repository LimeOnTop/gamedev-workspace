import { useEffect, useRef, useState } from 'react'
import { api, isImage, isVideo, type NodeDetail, type Reference } from '../api'
import { useWorkspace } from '../workspace'
import { FileIcon, ImageIcon, PlayIcon, TrashIcon, UploadIcon } from './Icons'
import Lightbox from './Lightbox'

type TileSize = 's' | 'm' | 'l'

const SIZE_KEY = 'workspace.collageSize'
const SIZES: { value: TileSize; label: string }[] = [
  { value: 's', label: 'S' },
  { value: 'm', label: 'M' },
  { value: 'l', label: 'L' },
]
const ACCEPT = 'image/*,video/*,application/pdf'

function loadSize(): TileSize {
  try {
    const v = localStorage.getItem(SIZE_KEY)
    if (v === 's' || v === 'm' || v === 'l') return v
  } catch {
    // storage unavailable
  }
  return 'm'
}

function isEditableTarget(target: EventTarget | null) {
  return target instanceof HTMLElement && (target.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName))
}

export default function ReferenceGallery({ node, onChanged }: { node: NodeDetail; onChanged: () => Promise<void> }) {
  const { uploadFiles, confirm, notify } = useWorkspace()
  const refs = node.references
  const [openIndex, setOpenIndex] = useState<number | null>(null)
  const [dragging, setDragging] = useState(false)
  const [progress, setProgress] = useState<{ done: number; total: number } | null>(null)
  const [size, setSize] = useState<TileSize>(loadSize)
  const input = useRef<HTMLInputElement>(null)
  const dragDepth = useRef(0)

  useEffect(() => {
    try {
      localStorage.setItem(SIZE_KEY, size)
    } catch {
      // storage unavailable
    }
  }, [size])

  // Keep the lightbox on a valid item when references are removed.
  useEffect(() => {
    if (openIndex === null) return
    if (refs.length === 0) setOpenIndex(null)
    else if (openIndex >= refs.length) setOpenIndex(refs.length - 1)
  }, [refs.length, openIndex])

  const upload = async (files: FileList | File[]) => {
    const list = Array.from(files)
    if (!list.length || progress) return
    try {
      await uploadFiles(node.id, list, (done, total) => setProgress({ done, total }))
    } finally {
      setProgress(null)
    }
  }

  // Paste screenshots or copied images straight into the page.
  const uploadRef = useRef(upload)
  uploadRef.current = upload
  useEffect(() => {
    const onPaste = (e: ClipboardEvent) => {
      if (isEditableTarget(e.target)) return
      const files = Array.from(e.clipboardData?.files ?? [])
      if (!files.length) return
      e.preventDefault()
      uploadRef.current(files)
    }
    window.addEventListener('paste', onPaste)
    return () => window.removeEventListener('paste', onPaste)
  }, [])

  const remove = async (ref: Reference) => {
    if (!(await confirm(`Удалить референс «${ref.original_name}»?`))) return
    try {
      await api.removeReference(ref.id)
      await onChanged()
    } catch (e) {
      notify((e as Error).message)
    }
  }

  const images = refs.filter(isImage).length

  return (
    <div
      className={`gallery ${dragging ? 'dragging' : ''}`}
      onDragEnter={(e) => {
        if (!e.dataTransfer.types.includes('Files')) return
        dragDepth.current++
        setDragging(true)
      }}
      onDragOver={(e) => {
        if (e.dataTransfer.types.includes('Files')) e.preventDefault()
      }}
      onDragLeave={() => {
        dragDepth.current = Math.max(0, dragDepth.current - 1)
        if (dragDepth.current === 0) setDragging(false)
      }}
      onDrop={(e) => {
        e.preventDefault()
        dragDepth.current = 0
        setDragging(false)
        if (e.dataTransfer.files.length) upload(e.dataTransfer.files)
      }}
    >
      <div className="panel-header gallery-header">
        <div className="gallery-title">
          <h2>Референсы</h2>
          {refs.length > 0 && (
            <span className="muted count">
              {refs.length}
              {images !== refs.length && ` · изображений ${images}`}
            </span>
          )}
        </div>
        <div className="gallery-tools">
          {refs.length > 0 && (
            <div className="segmented" role="group" aria-label="Размер плиток">
              {SIZES.map((s) => (
                <button key={s.value} className={size === s.value ? 'active' : ''} onClick={() => setSize(s.value)}>
                  {s.label}
                </button>
              ))}
            </div>
          )}
          <button className="btn primary small" onClick={() => input.current?.click()} disabled={!!progress}>
            <UploadIcon />
            {progress ? `Загрузка ${progress.done}/${progress.total}` : 'Загрузить'}
          </button>
        </div>
        <input
          ref={input}
          type="file"
          multiple
          hidden
          accept={ACCEPT}
          onChange={(e) => {
            if (e.target.files?.length) upload(e.target.files)
            e.target.value = ''
          }}
        />
      </div>

      {progress && (
        <div className="upload-progress" aria-hidden>
          <div style={{ width: `${(progress.done / progress.total) * 100}%` }} />
        </div>
      )}

      {refs.length === 0 ? (
        <button className="dropzone" onClick={() => input.current?.click()}>
          <ImageIcon size={44} />
          <strong>Добавьте референсы объекта</strong>
          <span>Перетащите сюда сразу несколько файлов, нажмите для выбора или вставьте из буфера (Ctrl/⌘ + V)</span>
        </button>
      ) : (
        <div className={`collage size-${size}`}>
          {refs.map((ref, i) => (
            <Tile key={ref.id} reference={ref} onOpen={() => setOpenIndex(i)} onDelete={() => remove(ref)} />
          ))}
          <button className="tile tile-add" onClick={() => input.current?.click()} disabled={!!progress}>
            <UploadIcon size={22} />
            <span>Добавить ещё</span>
          </button>
        </div>
      )}

      {dragging && (
        <div className="drop-overlay">
          <UploadIcon size={32} />
          <strong>Отпустите, чтобы загрузить</strong>
        </div>
      )}

      {openIndex !== null && refs[openIndex] && (
        <Lightbox
          references={refs}
          index={openIndex}
          onIndexChange={setOpenIndex}
          onClose={() => setOpenIndex(null)}
          onDelete={remove}
        />
      )}
    </div>
  )
}

function Tile({ reference, onOpen, onDelete }: { reference: Reference; onOpen: () => void; onDelete: () => void }) {
  return (
    <div className={`tile ${isImage(reference) ? '' : 'tile-doc'}`}>
      <button className="tile-open" onClick={onOpen} title={reference.original_name}>
        {isImage(reference) ? (
          <img src={reference.url} alt={reference.original_name} loading="lazy" />
        ) : isVideo(reference) ? (
          <>
            <video src={`${reference.url}#t=0.1`} preload="metadata" muted playsInline />
            <span className="tile-play"><PlayIcon size={22} /></span>
          </>
        ) : (
          <span className="tile-file">
            <FileIcon size={30} />
            <span>{reference.original_name}</span>
          </span>
        )}
        <span className="tile-caption">{reference.original_name}</span>
      </button>
      <button className="tile-delete" title="Удалить" onClick={onDelete}>
        <TrashIcon size={15} />
      </button>
    </div>
  )
}
