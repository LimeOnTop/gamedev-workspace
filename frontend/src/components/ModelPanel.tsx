import { useEffect, useRef, useState } from 'react'
import { api, formatSize, type Model3D, type NodeDetail } from '../api'
import { useWorkspace } from '../workspace'
import { CubeIcon } from './AssetIcons'
import { TrashIcon, UploadIcon } from './Icons'

const ACCEPT = '.glb,model/gltf-binary'
const MAX_SIZE = 100 * 1024 * 1024
// Three-quarter view with a little margin around the auto-framed model.
const INITIAL_ORBIT = '30deg 70deg 115%'

// The viewer pulls in three.js, so it is loaded only on pages that show a model.
let viewerModule: Promise<unknown> | null = null
const loadViewer = () => (viewerModule ??= import('@google/model-viewer'))

type ViewerElement = HTMLElement & {
  cameraOrbit: string
  cameraTarget: string
  fieldOfView: string
  jumpCameraToGoal: () => void
}

export default function ModelPanel({ node, onChanged }: { node: NodeDetail; onChanged: () => Promise<void> }) {
  const { confirm, notify } = useWorkspace()
  const model = node.model
  const [progress, setProgress] = useState<number | null>(null)
  const [dragging, setDragging] = useState(false)
  const input = useRef<HTMLInputElement>(null)
  const dragDepth = useRef(0)

  const upload = async (files: FileList | File[]) => {
    const file = Array.from(files)[0]
    if (!file || progress !== null) return
    if (!file.name.toLowerCase().endsWith('.glb')) {
      notify('Поддерживаются только модели в формате .glb')
      return
    }
    if (file.size > MAX_SIZE) {
      notify(`Файл слишком большой: максимум ${formatSize(MAX_SIZE)}`)
      return
    }
    setProgress(0)
    try {
      await api.uploadModel(node.id, file, setProgress)
      await onChanged()
    } catch (e) {
      notify((e as Error).message)
    } finally {
      setProgress(null)
    }
  }

  const remove = async () => {
    if (!model || !(await confirm(`Удалить 3D-модель «${model.original_name}»?`))) return
    try {
      await api.removeModel(node.id)
      await onChanged()
    } catch (e) {
      notify((e as Error).message)
    }
  }

  const pick = () => input.current?.click()
  const busy = progress !== null

  return (
    <div
      className={`model-panel ${dragging ? 'dragging' : ''}`}
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
        // Keep the drop from bubbling to the reference gallery.
        e.stopPropagation()
        dragDepth.current = 0
        setDragging(false)
        if (e.dataTransfer.files.length) upload(e.dataTransfer.files)
      }}
    >
      <div className="panel-header gallery-header">
        <div className="gallery-title">
          <h2>3D-модель</h2>
          {model && <span className="muted count">GLB · {formatSize(model.size)}</span>}
        </div>
        {model && (
          <div className="gallery-tools">
            <button className="btn ghost small" onClick={pick} disabled={busy}>
              <UploadIcon />
              {busy ? `Загрузка ${Math.round((progress ?? 0) * 100)}%` : 'Заменить'}
            </button>
            <button className="btn ghost small danger" onClick={remove} disabled={busy}>
              <TrashIcon /> Удалить
            </button>
          </div>
        )}
        <input
          ref={input}
          type="file"
          hidden
          accept={ACCEPT}
          onChange={(e) => {
            if (e.target.files?.length) upload(e.target.files)
            e.target.value = ''
          }}
        />
      </div>

      {busy && (
        <div className="upload-progress" aria-hidden>
          <div style={{ width: `${(progress ?? 0) * 100}%` }} />
        </div>
      )}

      {model ? (
        <ModelViewer model={model} />
      ) : (
        <button className="dropzone model-dropzone" onClick={pick} disabled={busy}>
          <CubeIcon size={44} />
          <strong>{busy ? `Загрузка ${Math.round((progress ?? 0) * 100)}%` : 'Добавьте 3D-модель объекта'}</strong>
          <span>Перетащите сюда файл .glb или нажмите для выбора (до {formatSize(MAX_SIZE)})</span>
        </button>
      )}

      {dragging && (
        <div className="drop-overlay">
          <UploadIcon size={32} />
          <strong>{model ? 'Отпустите, чтобы заменить модель' : 'Отпустите, чтобы загрузить'}</strong>
        </div>
      )}
    </div>
  )
}

function ModelViewer({ model }: { model: Model3D }) {
  const viewer = useRef<ViewerElement>(null)
  const [state, setState] = useState<{ loaded: number; error: boolean }>({ loaded: 0, error: false })
  const [autoRotate, setAutoRotate] = useState(false)

  useEffect(() => {
    loadViewer().catch(() => setState({ loaded: 0, error: true }))
  }, [])

  useEffect(() => {
    const el = viewer.current
    if (!el) return
    setState({ loaded: 0, error: false })
    const onProgress = (e: Event) => {
      const total = (e as CustomEvent<{ totalProgress: number }>).detail.totalProgress
      setState((s) => ({ ...s, loaded: total }))
    }
    const onLoad = () => setState({ loaded: 1, error: false })
    const onError = () => setState({ loaded: 0, error: true })
    el.addEventListener('progress', onProgress)
    el.addEventListener('load', onLoad)
    el.addEventListener('error', onError)
    return () => {
      el.removeEventListener('progress', onProgress)
      el.removeEventListener('load', onLoad)
      el.removeEventListener('error', onError)
    }
  }, [model.url])

  const resetView = () => {
    const el = viewer.current
    if (!el) return
    el.cameraOrbit = INITIAL_ORBIT
    el.cameraTarget = 'auto auto auto'
    el.fieldOfView = 'auto'
    el.jumpCameraToGoal()
  }

  return (
    <div className="model-stage">
      <model-viewer
        ref={viewer}
        key={model.url}
        src={model.url}
        alt={model.original_name}
        camera-controls=""
        camera-orbit={INITIAL_ORBIT}
        touch-action="pan-y"
        interaction-prompt="none"
        environment-image="neutral"
        shadow-intensity="1"
        shadow-softness="0.8"
        exposure="1"
        auto-rotate={autoRotate ? '' : undefined}
        auto-rotate-delay="0"
        rotation-per-second="20deg"
      />
      {state.loaded < 1 && !state.error && (
        <div className="model-loading">
          <span>Загрузка модели…</span>
          <div className="upload-progress">
            <div style={{ width: `${state.loaded * 100}%` }} />
          </div>
        </div>
      )}
      {state.error && <div className="model-loading error">Не удалось отобразить модель</div>}
      <div className="model-toolbar">
        <span className="model-name" title={model.original_name}>{model.original_name}</span>
        <div className="segmented" role="group" aria-label="Управление видом">
          <button className={autoRotate ? 'active' : ''} onClick={() => setAutoRotate((v) => !v)} title="Автовращение">
            Вращение
          </button>
          <button onClick={resetView} title="Вернуть камеру в исходное положение">
            Сброс
          </button>
        </div>
      </div>
      <div className="model-hint">ЛКМ — вращение · колесо — масштаб · ПКМ — сдвиг</div>
    </div>
  )
}
