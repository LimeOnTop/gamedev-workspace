import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useWorkspace } from '../workspace'
import { AssetCategoryIcon, CubeIcon } from './AssetIcons'

interface Props {
  value: string | null
  onChange: (category: string) => Promise<void>
}

/** Marks a file as a 3D asset of a given category (or removes it from the catalog). */
export default function AssetCategorySelect({ value, onChange }: Props) {
  const { assetCategories } = useWorkspace()
  const [saving, setSaving] = useState(false)
  const current = assetCategories.find((c) => c.id === value)

  return (
    <div className={`asset-select ${value ? 'is-asset' : ''}`}>
      <span className="asset-select-icon">
        {value ? <AssetCategoryIcon category={value} /> : <CubeIcon />}
      </span>
      <label className="asset-select-label" htmlFor="asset-category">3D-ассет</label>
      <select
        id="asset-category"
        value={value ?? ''}
        disabled={saving}
        onChange={async (e) => {
          setSaving(true)
          try {
            await onChange(e.target.value)
          } catch {
            // error already shown by the caller
          } finally {
            setSaving(false)
          }
        }}
      >
        <option value="">Нет (документ)</option>
        {assetCategories.map((c) => (
          <option key={c.id} value={c.id}>{c.name}</option>
        ))}
      </select>
      {current && (
        <Link className="asset-select-link" to={`/assets/${current.id}`}>в каталоге →</Link>
      )}
    </div>
  )
}
