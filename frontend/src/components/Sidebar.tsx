import { useEffect, useRef, useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { ancestorsOf, type TreeNode } from '../api'
import { useWorkspace } from '../workspace'
import { AssetCategoryIcon, CubeIcon } from './AssetIcons'
import {
  ChevronIcon, EditIcon, FileIcon, FilePlusIcon, FolderIcon, FolderPlusIcon, HomeIcon, TrashIcon, UploadIcon,
} from './Icons'

const STORAGE_KEY = 'workspace.expanded'

function loadExpanded(): Set<string> {
  try {
    return new Set(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]'))
  } catch {
    return new Set()
  }
}

export default function Sidebar() {
  const { tree, assetCategories, loading, error, createNode } = useWorkspace()
  const location = useLocation()
  const activeId = location.pathname.match(/^\/node\/([^/]+)/)?.[1] ?? null
  const [expanded, setExpanded] = useState<Set<string>>(loadExpanded)

  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify([...expanded]))
    } catch {
      // storage unavailable
    }
  }, [expanded])

  // Reveal the active node when navigating to it from elsewhere.
  useEffect(() => {
    if (!activeId) return
    const ancestors = ancestorsOf(tree, activeId)
    if (ancestors.some((id) => !expanded.has(id))) {
      setExpanded((prev) => new Set([...prev, ...ancestors]))
    }
  }, [activeId, tree]) // expanded is intentionally omitted: collapsing manually must stick

  const toggle = (id: string) =>
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <Link to="/" className="brand">
          <span className="brand-mark">◆</span> Gamedev Workspace
        </Link>
      </div>

      <nav className="sidebar-nav">
        <Link to="/" className={`nav-home ${location.pathname === '/' ? 'active' : ''}`}>
          <HomeIcon /> Обзор
        </Link>
      </nav>

      <div className="sidebar-scroll">
      <div className="sidebar-section">
        <Link to="/assets" className="section-link">3D-ассеты</Link>
      </div>
      <nav className="asset-nav" aria-label="Каталог 3D-ассетов">
        {assetCategories.map((c) => (
          <Link
            key={c.id}
            to={`/assets/${c.id}`}
            className={`asset-nav-item ${location.pathname === `/assets/${c.id}` ? 'active' : ''}`}
            title={c.description}
          >
            <AssetCategoryIcon category={c.id} className="ico-asset" />
            <span className="tree-name">{c.name}</span>
            <span className="asset-count">{c.asset_count || ''}</span>
          </Link>
        ))}
      </nav>

      <div className="sidebar-section">
        <span>Структура</span>
        <div className="row-actions visible">
          <button className="icon-btn" title="Новая папка" onClick={() => createNode(null, 'folder')}>
            <FolderPlusIcon />
          </button>
          <button className="icon-btn" title="Новый файл" onClick={() => createNode(null, 'file')}>
            <FilePlusIcon />
          </button>
        </div>
      </div>

      <div className="tree" role="tree">
        {loading && <div className="tree-empty">Загрузка…</div>}
        {error && <div className="tree-empty error">Ошибка: {error}</div>}
        {!loading && !error && tree.length === 0 && (
          <div className="tree-empty">Пока пусто — создайте первую папку.</div>
        )}
        {tree.map((node) => (
          <TreeItem key={node.id} node={node} depth={0} expanded={expanded} toggle={toggle} activeId={activeId} />
        ))}
      </div>
      </div>
    </aside>
  )
}

interface TreeItemProps {
  node: TreeNode
  depth: number
  expanded: Set<string>
  toggle: (id: string) => void
  activeId: string | null
}

function TreeItem({ node, depth, expanded, toggle, activeId }: TreeItemProps) {
  const { createNode, renameNode, deleteNode, uploadFiles } = useWorkspace()
  const fileInput = useRef<HTMLInputElement>(null)
  const isFolder = node.kind === 'folder'
  const isOpen = expanded.has(node.id)

  const openFolder = () => {
    if (!isOpen) toggle(node.id)
  }

  return (
    <div role="treeitem" aria-expanded={isFolder ? isOpen : undefined}>
      <div className={`tree-row ${activeId === node.id ? 'active' : ''}`} style={{ paddingLeft: 8 + depth * 14 }}>
        {isFolder ? (
          <button
            className={`chevron ${isOpen ? 'open' : ''}`}
            onClick={() => toggle(node.id)}
            aria-label={isOpen ? 'Свернуть' : 'Развернуть'}
          >
            <ChevronIcon size={14} />
          </button>
        ) : (
          <span className="chevron-spacer" />
        )}
        <Link to={`/node/${node.id}`} className="tree-link" onClick={isFolder ? openFolder : undefined}>
          {isFolder ? (
            <FolderIcon className="ico-folder" />
          ) : node.asset_category ? (
            <CubeIcon className="ico-asset" aria-label="3D-ассет" />
          ) : (
            <FileIcon className="ico-file" />
          )}
          <span className="tree-name">{node.name}</span>
        </Link>
        <div className="row-actions">
          {isFolder ? (
            <>
              <button className="icon-btn" title="Новая папка" onClick={() => { openFolder(); createNode(node.id, 'folder') }}>
                <FolderPlusIcon />
              </button>
              <button className="icon-btn" title="Новый файл" onClick={() => { openFolder(); createNode(node.id, 'file') }}>
                <FilePlusIcon />
              </button>
            </>
          ) : (
            <>
              <button className="icon-btn" title="Загрузить референс" onClick={() => fileInput.current?.click()}>
                <UploadIcon />
              </button>
              <input
                ref={fileInput}
                type="file"
                multiple
                hidden
                accept="image/*,video/*,application/pdf"
                onChange={(e) => {
                  if (e.target.files?.length) uploadFiles(node.id, e.target.files)
                  e.target.value = ''
                }}
              />
            </>
          )}
          <button className="icon-btn" title="Переименовать" onClick={() => renameNode(node.id, node.name)}>
            <EditIcon />
          </button>
          <button className="icon-btn danger" title="Удалить" onClick={() => deleteNode(node.id, node.name, node.kind)}>
            <TrashIcon />
          </button>
        </div>
      </div>
      {isFolder && isOpen && (
        <div role="group">
          {node.children.length === 0 && (
            <div className="tree-empty nested" style={{ paddingLeft: 30 + depth * 14 }}>пусто</div>
          )}
          {node.children.map((child) => (
            <TreeItem key={child.id} node={child} depth={depth + 1} expanded={expanded} toggle={toggle} activeId={activeId} />
          ))}
        </div>
      )}
    </div>
  )
}
