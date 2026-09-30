import { Link } from 'react-router-dom'
import { countLabel, subtreeStats, type Kind, type TreeNode } from '../api'
import { useWorkspace } from '../workspace'
import { FileIcon, FilePlusIcon, FolderIcon, FolderPlusIcon } from './Icons'

export default function NodeCards({ nodes, parentId }: { nodes: TreeNode[]; parentId: string | null }) {
  const { createNode } = useWorkspace()

  return (
    <div className="cards">
      {nodes.map((node) => (
        <NodeCard key={node.id} node={node} />
      ))}
      <AddCard kind="folder" onClick={() => createNode(parentId, 'folder')} />
      <AddCard kind="file" onClick={() => createNode(parentId, 'file')} />
    </div>
  )
}

function NodeCard({ node }: { node: TreeNode }) {
  const stats = subtreeStats(node)
  const isFolder = node.kind === 'folder'

  return (
    <Link to={`/node/${node.id}`} className="card">
      <div className="card-cover">
        {stats.preview ? (
          <img src={stats.preview} alt="" loading="lazy" />
        ) : (
          <div className="card-cover-placeholder">
            {isFolder ? <FolderIcon size={40} /> : <FileIcon size={40} />}
          </div>
        )}
        <span className={`card-badge ${node.kind}`}>{isFolder ? 'Папка' : 'Файл'}</span>
      </div>
      <div className="card-body">
        <h3 className="card-title">{node.name}</h3>
        {node.summary && <p className="card-summary">{node.summary}</p>}
        {isFolder && <div className="card-meta">{countLabel(stats.folders, stats.files)}</div>}
      </div>
    </Link>
  )
}

function AddCard({ kind, onClick }: { kind: Kind; onClick: () => void }) {
  return (
    <button className="card card-add" onClick={onClick}>
      {kind === 'folder' ? <FolderPlusIcon size={28} /> : <FilePlusIcon size={28} />}
      <span>{kind === 'folder' ? 'Новая папка' : 'Новый файл'}</span>
    </button>
  )
}
