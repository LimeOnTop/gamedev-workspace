import { useEffect, useRef, useState } from 'react'
import { api, formatSize, isImage, isVideo, type NodeDetail, type Reference } from '../api'
import { useWorkspace } from '../workspace'
import { FileIcon, ImageIcon, TrashIcon, UploadIcon } from './Icons'

export default function ReferenceGallery({ node, onChanged }: { node: NodeDetail; onChanged: () => Promise<void> }) {
  const { uploadFiles, confirm, notify } = useWorkspace()
  const refs = node.references
  const [selectedId, setSelectedId] = useState<string | null>(refs[0]?.id ?? null)
  const [dragging, setDragging] = useState(false)
  const [uploading, setUploading] = useState(false)
  const input = useRef<HTMLInputElement>(null)

  // Keep a valid selection; jump to the newest reference after an upload.
  const prevCount = useRef(refs.length)
  useEffect(() => {
    if (refs.length > prevCount.current) setSelectedId(refs[refs.length - 1].id)
    else if (!refs.some((r) => r.id === selectedId)) setSelectedId(refs[0]?.id ?? null)
    prevCount.current = refs.length
  }, [refs, selectedId])

  const selected = refs.find((r) => r.id === selectedId) ?? null

  const upload = async (files: FileList | File[]) => {
    setUploading(true)
    try {
      await uploadFiles(node.id, files)
    } finally {
      setUploading(false)
    }
  }

  const remove = async (ref: Reference) => {
    if (!(await confirm(`Удалить референс «${ref.original_name}»?`))) return
    try {
      await api.removeReference(ref.id)
      await onChanged()
    } catch (e) {
      notify((e as Error).message)
    }
  }

  return (
    <div
      className={`gallery ${dragging ? 'dragging' : ''}`}
      onDragOver={(e) => {
        e.preventDefault()
        setDragging(true)
      }}
      onDragLeave={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget as Node)) setDragging(false)
      }}
      onDrop={(e) => {
        e.preventDefault()
        setDragging(false)
        if (e.dataTransfer.files.length) upload(e.dataTransfer.files)
      }}
    >
      <div className="panel-header">
        <h2>Референсы</h2>
        <button className="btn primary small" onClick={() => input.current?.click()} disabled={uploading}>
          <UploadIcon /> {uploading ? 'Загрузка…' : 'Загрузить референс'}
        </button>
        <input
          ref={input}
          type="file"
          multiple
          hidden
          accept="image/*,video/*,application/pdf"
          onChange={(e) => {
            if (e.target.files?.length) upload(e.target.files)
            e.target.value = ''
          }}
        />
      </div>

      <div className="viewer">
        {selected ? (
          <Preview reference={selected} />
        ) : (
          <button className="viewer-empty" onClick={() => input.current?.click()}>
            <ImageIcon size={44} />
            <strong>Здесь будет референс объекта</strong>
            <span>Нажмите или перетащите изображение, видео или PDF</span>
          </button>
        )}
      </div>

      {selected && (
        <div className="viewer-caption">
          <a href={selected.url} target="_blank" rel="noreferrer" className="caption-name">{selected.original_name}</a>
          <span className="muted">{formatSize(selected.size)}</span>
          <button className="icon-btn danger" title="Удалить референс" onClick={() => remove(selected)}>
            <TrashIcon />
          </button>
        </div>
      )}

      {refs.length > 1 && (
        <div className="thumbs">
          {refs.map((r) => (
            <button
              key={r.id}
              className={`thumb ${r.id === selectedId ? 'active' : ''}`}
              onClick={() => setSelectedId(r.id)}
              title={r.original_name}
            >
              {isImage(r) ? <img src={r.url} alt={r.original_name} loading="lazy" /> : <FileIcon size={22} />}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}

function Preview({ reference }: { reference: Reference }) {
  if (isImage(reference)) {
    return (
      <a href={reference.url} target="_blank" rel="noreferrer" className="viewer-media">
        <img src={reference.url} alt={reference.original_name} />
      </a>
    )
  }
  if (isVideo(reference)) {
    return <video className="viewer-media" src={reference.url} controls />
  }
  return (
    <a href={reference.url} target="_blank" rel="noreferrer" className="viewer-empty">
      <FileIcon size={44} />
      <strong>{reference.original_name}</strong>
      <span>Открыть файл</span>
    </a>
  )
}
