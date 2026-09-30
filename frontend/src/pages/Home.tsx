import { countLabel, subtreeStats } from '../api'
import NodeCards from '../components/NodeCards'
import { useWorkspace } from '../workspace'

export default function Home() {
  const { tree, loading, error } = useWorkspace()

  const totals = tree.reduce(
    (acc, node) => {
      const s = subtreeStats(node)
      acc.folders += s.folders + (node.kind === 'folder' ? 1 : 0)
      acc.files += s.files + (node.kind === 'file' ? 1 : 0)
      return acc
    },
    { folders: 0, files: 0 },
  )

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Обзор</h1>
          <p className="muted">{loading ? 'Загрузка…' : countLabel(totals.folders, totals.files)}</p>
        </div>
      </header>
      {error ? <div className="alert">Не удалось загрузить данные: {error}</div> : <NodeCards nodes={tree} parentId={null} />}
    </div>
  )
}
