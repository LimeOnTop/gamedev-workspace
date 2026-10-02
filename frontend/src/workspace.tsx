import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type AssetCategory, type FileType, type Kind, type TreeNode } from './api'

interface Dialog {
  title: string
  initial: string
  confirmOnly?: boolean
  danger?: boolean
  message?: string
  action?: string
  resolve: (value: string | null) => void
}

export interface ConfirmOptions {
  /** Explanation shown under the title. */
  message?: string
  /** Confirm button label; "Удалить" by default. */
  action?: string
  /** Red confirm button; true by default. */
  danger?: boolean
}

interface WorkspaceContextValue {
  tree: TreeNode[]
  assetCategories: AssetCategory[]
  loading: boolean
  error: string | null
  /** Bumped on every mutation so open pages can refetch their node. */
  version: number
  reload: () => Promise<void>
  notify: (message: string) => void
  askName: (title: string, initial?: string) => Promise<string | null>
  confirm: (title: string, options?: ConfirmOptions) => Promise<boolean>
  createNode: (parentId: string | null, kind: Kind, fileType?: FileType) => Promise<void>
  renameNode: (id: string, current: string) => Promise<void>
  deleteNode: (id: string, name: string, kind: Kind) => Promise<void>
  uploadFiles: (
    nodeId: string,
    files: FileList | File[],
    onProgress?: (done: number, total: number) => void,
  ) => Promise<void>
}

const UPLOAD_CONCURRENCY = 3

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null)

export function useWorkspace() {
  const ctx = useContext(WorkspaceContext)
  if (!ctx) throw new Error('useWorkspace must be used inside WorkspaceProvider')
  return ctx
}

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const [tree, setTree] = useState<TreeNode[]>([])
  const [assetCategories, setAssetCategories] = useState<AssetCategory[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [version, setVersion] = useState(0)
  const [dialog, setDialog] = useState<Dialog | null>(null)
  const [toast, setToast] = useState<string | null>(null)
  const toastTimer = useRef<number | undefined>(undefined)

  const notify = useCallback((message: string) => {
    setToast(message)
    window.clearTimeout(toastTimer.current)
    toastTimer.current = window.setTimeout(() => setToast(null), 4000)
  }, [])

  const reload = useCallback(async () => {
    try {
      const [nextTree, categories] = await Promise.all([api.tree(), api.assetCategories()])
      setTree(nextTree)
      setAssetCategories(categories)
      setError(null)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
      setVersion((v) => v + 1)
    }
  }, [])

  useEffect(() => {
    reload()
    // Pick up changes made elsewhere (e.g. by Claude via MCP) when the tab regains focus.
    const onFocus = () => reload()
    window.addEventListener('focus', onFocus)
    return () => window.removeEventListener('focus', onFocus)
  }, [reload])

  const askName = useCallback(
    (title: string, initial = '') =>
      new Promise<string | null>((resolve) => setDialog({ title, initial, resolve })),
    [],
  )

  const confirm = useCallback(
    (title: string, { message, action, danger = true }: ConfirmOptions = {}) =>
      new Promise<boolean>((resolve) =>
        setDialog({
          title, initial: '', confirmOnly: true, danger, message, action, resolve: (v) => resolve(v !== null),
        }),
      ),
    [],
  )

  const createNode = useCallback(
    async (parentId: string | null, kind: Kind, fileType?: FileType) => {
      const title = kind === 'folder' ? 'Новая папка' : fileType === 'scenario' ? 'Новый сценарий' : 'Новый файл'
      const name = await askName(title)
      if (!name) return
      try {
        const node = await api.create(parentId, kind, name, fileType)
        await reload()
        navigate(`/node/${node.id}`)
      } catch (e) {
        notify((e as Error).message)
      }
    },
    [askName, reload, navigate, notify],
  )

  const renameNode = useCallback(
    async (id: string, current: string) => {
      const name = await askName('Переименовать', current)
      if (!name || name === current) return
      try {
        await api.update(id, { name })
        await reload()
      } catch (e) {
        notify((e as Error).message)
      }
    },
    [askName, reload, notify],
  )

  const deleteNode = useCallback(
    async (id: string, name: string, kind: Kind) => {
      const what = kind === 'folder' ? `папку «${name}» со всем содержимым` : `файл «${name}»`
      if (!(await confirm(`Удалить ${what}?`))) return
      try {
        await api.remove(id)
        await reload()
        if (window.location.pathname.startsWith(`/node/${id}`)) navigate('/')
      } catch (e) {
        notify((e as Error).message)
      }
    },
    [confirm, reload, navigate, notify],
  )

  const uploadFiles = useCallback(
    async (nodeId: string, files: FileList | File[], onProgress?: (done: number, total: number) => void) => {
      const queue = Array.from(files)
      const total = queue.length
      const failed: string[] = []
      let done = 0
      let ok = 0
      onProgress?.(0, total)

      // A few parallel uploads keep large batches fast without flooding the server.
      const worker = async () => {
        for (let file = queue.shift(); file; file = queue.shift()) {
          try {
            await api.upload(nodeId, file)
            ok++
          } catch (e) {
            failed.push(`${file.name}: ${(e as Error).message}`)
          }
          onProgress?.(++done, total)
        }
      }
      await Promise.all(Array.from({ length: Math.min(UPLOAD_CONCURRENCY, total) }, worker))

      if (failed.length) {
        notify(failed.length === 1 ? failed[0] : `Не загружено ${failed.length} из ${total}: ${failed[0]}…`)
      } else if (ok) {
        notify(ok === 1 ? 'Референс загружен' : `Загружено референсов: ${ok}`)
      }
      if (ok) await reload()
    },
    [notify, reload],
  )

  const value: WorkspaceContextValue = {
    tree, assetCategories, loading, error, version, reload, notify, askName, confirm,
    createNode, renameNode, deleteNode, uploadFiles,
  }

  return (
    <WorkspaceContext.Provider value={value}>
      {children}
      {dialog && <DialogView dialog={dialog} onClose={() => setDialog(null)} />}
      {toast && <div className="toast" role="status">{toast}</div>}
    </WorkspaceContext.Provider>
  )
}

function DialogView({ dialog, onClose }: { dialog: Dialog; onClose: () => void }) {
  const [value, setValue] = useState(dialog.initial)
  const finish = (result: string | null) => {
    dialog.resolve(result)
    onClose()
  }
  const submit = () => {
    if (dialog.confirmOnly) return finish('yes')
    const trimmed = value.trim()
    if (trimmed) finish(trimmed)
  }

  return (
    <div className="modal-backdrop" onMouseDown={() => finish(null)}>
      <form
        className="modal"
        onMouseDown={(e) => e.stopPropagation()}
        onSubmit={(e) => {
          e.preventDefault()
          submit()
        }}
        onKeyDown={(e) => e.key === 'Escape' && finish(null)}
      >
        <h3>{dialog.title}</h3>
        {dialog.message && <p className="modal-message">{dialog.message}</p>}
        {!dialog.confirmOnly && (
          <input
            autoFocus
            value={value}
            maxLength={200}
            placeholder="Название"
            onChange={(e) => setValue(e.target.value)}
            onFocus={(e) => e.target.select()}
          />
        )}
        <div className="modal-actions">
          <button type="button" className="btn ghost" onClick={() => finish(null)}>
            Отмена
          </button>
          <button type="submit" className={`btn ${dialog.danger ? 'danger' : 'primary'}`} autoFocus={dialog.confirmOnly}>
            {dialog.confirmOnly ? (dialog.action ?? 'Удалить') : 'Сохранить'}
          </button>
        </div>
      </form>
    </div>
  )
}
