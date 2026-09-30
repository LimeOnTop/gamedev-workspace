import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, plural, type Asset } from '../api'
import { AssetCategoryIcon, CubeIcon } from '../components/AssetIcons'
import { useWorkspace } from '../workspace'

export default function AssetsPage() {
  const { category = '' } = useParams()
  const { assetCategories, version } = useWorkspace()
  const [assets, setAssets] = useState<Asset[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')

  useEffect(() => {
    let cancelled = false
    setError(null)
    api
      .assets(category)
      .then((list) => !cancelled && setAssets(list))
      .catch((e: Error) => !cancelled && setError(e.message))
    return () => {
      cancelled = true
    }
  }, [category, version])

  useEffect(() => setQuery(''), [category])

  const meta = assetCategories.find((c) => c.id === category)
  const title = category ? meta?.name ?? category : 'Все 3D-ассеты'
  const needle = query.trim().toLowerCase()
  const visible = (assets ?? []).filter(
    (a) => !needle || a.name.toLowerCase().includes(needle) || a.path.some((p) => p.name.toLowerCase().includes(needle)),
  )
  const namesById = new Map(assetCategories.map((c) => [c.id, c.name]))

  return (
    <div className="page">
      <nav className="breadcrumbs" aria-label="Путь">
        <Link to="/">Обзор</Link>
        <span className="sep">/</span>
        <Link to="/assets">3D-ассеты</Link>
      </nav>

      <header className="page-header">
        <div>
          <div className="title-row">
            {category ? <AssetCategoryIcon category={category} size={26} className="ico-asset" /> : <CubeIcon size={26} className="ico-asset" />}
            <h1>{title}</h1>
          </div>
          <p className="muted">
            {meta?.description ?? 'Объекты, для которых будет сделана 3D-модель.'}
            {assets && ` · ${assets.length} ${plural(assets.length, 'объект', 'объекта', 'объектов')}`}
          </p>
        </div>
        {assets && assets.length > 6 && (
          <input
            className="asset-filter"
            type="search"
            placeholder="Фильтр по названию или папке"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        )}
      </header>

      {error ? (
        <div className="alert">{error === 'not found' ? 'Такой категории нет.' : error}</div>
      ) : !assets ? (
        <p className="muted">Загрузка…</p>
      ) : assets.length === 0 ? (
        <div className="empty-state">
          <CubeIcon size={40} />
          <strong>В этой категории пока нет объектов</strong>
          <span>
            Откройте файл объекта в структуре и выберите категорию в поле «3D-ассет» под заголовком — он появится
            здесь, оставаясь в своей папке.
          </span>
        </div>
      ) : (
        <div className="cards">
          {visible.map((asset) => (
            <Link key={asset.id} to={`/node/${asset.id}`} className="card">
              <div className="card-cover">
                {asset.preview ? (
                  <img src={asset.preview} alt="" loading="lazy" />
                ) : (
                  <div className="card-cover-placeholder">
                    <AssetCategoryIcon category={asset.category} size={40} />
                  </div>
                )}
                {!category && <span className="card-badge asset">{namesById.get(asset.category) ?? asset.category}</span>}
              </div>
              <div className="card-body">
                <h3 className="card-title">{asset.name}</h3>
                {asset.summary && <p className="card-summary">{asset.summary}</p>}
                {asset.path.length > 0 && (
                  <div className="card-meta">{asset.path.map((p) => p.name).join(' / ')}</div>
                )}
              </div>
            </Link>
          ))}
          {visible.length === 0 && <p className="muted">Ничего не найдено.</p>}
        </div>
      )}
    </div>
  )
}
