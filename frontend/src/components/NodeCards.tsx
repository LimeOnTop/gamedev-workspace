import { Link } from 'react-router-dom'
import { countLabel, subtreeStats, type TreeNode } from '../api'
import { useWorkspace } from '../workspace'
import { FileIcon, FilePlusIcon, FolderIcon, FolderPlusIcon, ScrollIcon, ScrollPlusIcon } from './Icons'

type AddKind = 'folder' | 'file' | 'scenario'

const ADD_CARDS: Record<AddKind, { label: string; Icon: typeof FileIcon }> = {
  folder: { label: 'Новая папка', Icon: FolderPlusIcon },
  file: { label: 'Новый файл', Icon: FilePlusIcon },
  scenario: { label: 'Новый сценарий', Icon: ScrollPlusIcon },
}

// Scenario summaries are raw Markdown; strip the markup for the card.
function plainSummary(markdown: string): string {
  return markdown
    .split('\n')
    .filter((line) => !/^\s*\|?\s*:?-{3,}/.test(line))
    .join(' ')
    .replace(/[#*_`>|]/g, ' ')
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1')
    .replace(/\s+/g, ' ')
    .trim()
}

export default function NodeCards({ nodes, parentId }: { nodes: TreeNode[]; parentId: string | null }) {
  const { createNode } = useWorkspace()

  return (
    <div className="cards">
      {nodes.map((node) => (
        <NodeCard key={node.id} node={node} />
      ))}
      <AddCard kind="folder" onClick={() => createNode(parentId, 'folder')} />
      <AddCard kind="file" onClick={() => createNode(parentId, 'file')} />
      <AddCard kind="scenario" onClick={() => createNode(parentId, 'file', 'scenario')} />
    </div>
  )
}

function NodeCard({ node }: { node: TreeNode }) {
  const stats = subtreeStats(node)
  const isFolder = node.kind === 'folder'
  const isScenario = node.file_type === 'scenario'
  const summary = isScenario ? plainSummary(node.summary) : node.summary

  return (
    <Link to={`/node/${node.id}`} className="card">
      <div className="card-cover">
        {stats.preview ? (
          <img src={stats.preview} alt="" loading="lazy" />
        ) : (
          <div className="card-cover-placeholder">
            {isFolder ? <FolderIcon size={40} /> : isScenario ? <ScrollIcon size={40} /> : <FileIcon size={40} />}
          </div>
        )}
        <span className={`card-badge ${isScenario ? 'scenario' : node.kind}`}>
          {isFolder ? 'Папка' : isScenario ? 'Сценарий' : 'Файл'}
        </span>
      </div>
      <div className="card-body">
        <h3 className="card-title">{node.name}</h3>
        {summary && <p className="card-summary">{summary}</p>}
        {isFolder && <div className="card-meta">{countLabel(stats.folders, stats.files)}</div>}
      </div>
    </Link>
  )
}

function AddCard({ kind, onClick }: { kind: AddKind; onClick: () => void }) {
  const { label, Icon } = ADD_CARDS[kind]
  return (
    <button className="card card-add" onClick={onClick}>
      <Icon size={28} />
      <span>{label}</span>
    </button>
  )
}
