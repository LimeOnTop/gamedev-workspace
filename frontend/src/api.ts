export type Kind = 'folder' | 'file'

export interface TreeNode {
  id: string
  parent_id: string | null
  kind: Kind
  name: string
  summary: string
  preview: string | null
  asset_category: string | null
  children: TreeNode[]
}

export interface Characteristic {
  key: string
  value: string
}

export interface Reference {
  id: string
  node_id: string
  original_name: string
  url: string
  content_type: string
  size: number
  created_at: string
}

/** The file's 3D model (GLB); a file has at most one. */
export interface Model3D {
  id: string
  node_id: string
  original_name: string
  url: string
  content_type: string
  size: number
  created_at: string
}

export interface NodeDetail {
  id: string
  parent_id: string | null
  kind: Kind
  name: string
  description: string
  mechanics: string
  characteristics: Characteristic[]
  reference_prompt: string
  asset_category: string | null
  created_at: string
  updated_at: string
  references: Reference[]
  model: Model3D | null
  path: { id: string; name: string }[]
}

export interface NodePatch {
  name?: string
  description?: string
  mechanics?: string
  characteristics?: Characteristic[]
  reference_prompt?: string
  /** "" removes the file from the 3D asset catalog. */
  asset_category?: string
}

export interface AssetCategory {
  id: string
  name: string
  description: string
  asset_count: number
}

export interface Asset {
  id: string
  parent_id: string | null
  category: string
  name: string
  summary: string
  preview: string | null
  has_model: boolean
  path: { id: string; name: string }[]
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init)
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`
    try {
      const body = await res.json()
      if (body?.error) message = body.error
    } catch {
      // non-JSON error body
    }
    throw new Error(message)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// XHR instead of fetch: models are large and fetch cannot report upload progress.
function uploadWithProgress<T>(url: string, method: string, form: FormData, onProgress?: (fraction: number) => void) {
  return new Promise<T>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open(method, url)
    xhr.responseType = 'json'
    xhr.upload.onprogress = (e) => e.lengthComputable && onProgress?.(e.loaded / e.total)
    xhr.onerror = () => reject(new Error('Сетевая ошибка при загрузке'))
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) resolve(xhr.response as T)
      else reject(new Error(xhr.response?.error ?? `${xhr.status} ${xhr.statusText}`))
    }
    xhr.send(form)
  })
}

const json = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(body),
})

export const api = {
  tree: () => request<TreeNode[]>('/api/tree'),
  node: (id: string) => request<NodeDetail>(`/api/nodes/${id}`),
  create: (parentId: string | null, kind: Kind, name: string) =>
    request<NodeDetail>('/api/nodes', json('POST', { parent_id: parentId, kind, name })),
  update: (id: string, patch: NodePatch) => request<NodeDetail>(`/api/nodes/${id}`, json('PATCH', patch)),
  remove: (id: string) => request<void>(`/api/nodes/${id}`, { method: 'DELETE' }),
  upload: (nodeId: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return request<Reference>(`/api/nodes/${nodeId}/references`, { method: 'POST', body: form })
  },
  removeReference: (id: string) => request<void>(`/api/references/${id}`, { method: 'DELETE' }),
  uploadModel: (nodeId: string, file: File, onProgress?: (fraction: number) => void) => {
    const form = new FormData()
    form.append('file', file)
    return uploadWithProgress<Model3D>(`/api/nodes/${nodeId}/model`, 'PUT', form, onProgress)
  },
  removeModel: (nodeId: string) => request<void>(`/api/nodes/${nodeId}/model`, { method: 'DELETE' }),
  assetCategories: () => request<AssetCategory[]>('/api/asset-categories'),
  assets: (category?: string) =>
    request<Asset[]>(`/api/assets${category ? `?category=${encodeURIComponent(category)}` : ''}`),
}

export const isImage = (r: Reference) => r.content_type.startsWith('image/')
export const isVideo = (r: Reference) => r.content_type.startsWith('video/')

export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} Б`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} КБ`
  return `${(bytes / 1024 / 1024).toFixed(1)} МБ`
}

export interface SubtreeStats {
  folders: number
  files: number
  preview: string | null
}

// Counts descendants and finds the first image preview in the subtree.
export function subtreeStats(node: TreeNode): SubtreeStats {
  const stats: SubtreeStats = { folders: 0, files: 0, preview: node.preview }
  const walk = (n: TreeNode) => {
    for (const c of n.children) {
      if (c.kind === 'folder') stats.folders++
      else stats.files++
      if (!stats.preview && c.preview) stats.preview = c.preview
      walk(c)
    }
  }
  walk(node)
  return stats
}

export function findNode(nodes: TreeNode[], id: string): TreeNode | null {
  for (const n of nodes) {
    if (n.id === id) return n
    const found = findNode(n.children, id)
    if (found) return found
  }
  return null
}

export function ancestorsOf(nodes: TreeNode[], id: string): string[] {
  const path: string[] = []
  const walk = (list: TreeNode[]): boolean => {
    for (const n of list) {
      if (n.id === id) return true
      path.push(n.id)
      if (walk(n.children)) return true
      path.pop()
    }
    return false
  }
  return walk(nodes) ? path : []
}

export function plural(n: number, one: string, few: string, many: string): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return one
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return few
  return many
}

export function countLabel(folders: number, files: number): string {
  const parts: string[] = []
  if (folders) parts.push(`${folders} ${plural(folders, 'папка', 'папки', 'папок')}`)
  if (files) parts.push(`${files} ${plural(files, 'файл', 'файла', 'файлов')}`)
  return parts.length ? parts.join(' · ') : 'Пусто'
}
