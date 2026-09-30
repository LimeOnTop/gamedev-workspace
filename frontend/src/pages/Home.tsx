import { Link } from 'react-router-dom'
import { countLabel, plural, subtreeStats } from '../api'
import { AssetCategoryIcon } from '../components/AssetIcons'
import NodeCards from '../components/NodeCards'
import { useWorkspace } from '../workspace'

export default function Home() {
  const { tree, assetCategories, loading, error } = useWorkspace()

  const totals = tree.reduce(
    (acc, node) => {
      const s = subtreeStats(node)
      acc.folders += s.folders + (node.kind === 'folder' ? 1 : 0)
      acc.files += s.files + (node.kind === 'file' ? 1 : 0)
      return acc
    },
    { folders: 0, files: 0 },
  )
  const assetTotal = assetCategories.reduce((sum, c) => sum + c.asset_count, 0)

  return (
    <div className="page">
      <header className="page-header">
        <div>
          <h1>Обзор</h1>
          <p className="muted">
            {loading
              ? 'Загрузка…'
              : `${countLabel(totals.folders, totals.files)} · ${assetTotal} ${plural(assetTotal, '3D-ассет', '3D-ассета', '3D-ассетов')}`}
          </p>
        </div>
      </header>

      {error ? (
        <div className="alert">Не удалось загрузить данные: {error}</div>
      ) : (
        <>
          {assetCategories.length > 0 && (
            <>
              <h2 className="section-title">
                <Link to="/assets">3D-ассеты</Link>
              </h2>
              <div className="asset-tiles">
                {assetCategories.map((c) => (
                  <Link key={c.id} to={`/assets/${c.id}`} className="asset-tile" title={c.description}>
                    <span className="asset-tile-icon">
                      <AssetCategoryIcon category={c.id} size={22} />
                    </span>
                    <span className="asset-tile-name">{c.name}</span>
                    <span className="asset-tile-count">{c.asset_count}</span>
                  </Link>
                ))}
              </div>
            </>
          )}
          <h2 className="section-title">Структура</h2>
          <NodeCards nodes={tree} parentId={null} />
        </>
      )}
    </div>
  )
}
