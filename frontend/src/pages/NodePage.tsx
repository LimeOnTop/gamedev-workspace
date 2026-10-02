import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, findNode, type NodeDetail } from '../api'
import { EditIcon, FileIcon, FolderIcon, ScrollIcon, TrashIcon } from '../components/Icons'
import NodeCards from '../components/NodeCards'
import EditableText from '../components/EditableText'
import CharacteristicsEditor from '../components/CharacteristicsEditor'
import ReferenceGallery from '../components/ReferenceGallery'
import ModelPanel from '../components/ModelPanel'
import AssetCategorySelect from '../components/AssetCategorySelect'
import PromptPanel from '../components/PromptPanel'
import ScenarioDocument from '../components/ScenarioDocument'
import { useWorkspace } from '../workspace'

export default function NodePage() {
  const { id = '' } = useParams()
  const { tree, version, reload, renameNode, deleteNode, notify, confirm } = useWorkspace()
  const [node, setNode] = useState<NodeDetail | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    api
      .node(id)
      .then((n) => {
        if (cancelled) return
        setNode(n)
        setError(null)
      })
      .catch((e: Error) => !cancelled && setError(e.message))
    return () => {
      cancelled = true
    }
  }, [id, version])

  if (error) {
    return (
      <div className="page">
        <div className="alert">{error === 'not found' ? 'Элемент не найден — возможно, он был удалён.' : error}</div>
        <Link to="/" className="btn ghost">На главную</Link>
      </div>
    )
  }
  if (!node || node.id !== id) return <div className="page muted">Загрузка…</div>

  const save = async (patch: Parameters<typeof api.update>[1]) => {
    try {
      setNode(await api.update(node.id, patch))
      await reload()
    } catch (e) {
      notify((e as Error).message)
      throw e
    }
  }

  const isFolder = node.kind === 'folder'
  const isScenario = node.file_type === 'scenario'
  const treeNode = findNode(tree, node.id)

  // Object fields stay stored while the file is a scenario, so the conversion is reversible.
  const makeScenario = async () => {
    const ok = await confirm(`Сделать «${node.name}» сценарием?`, {
      message:
        'У сценария есть только структурированное описание. Референсы, 3D-модель, характеристики, механики и промпт ' +
        'будут скрыты, но не удалены — они вернутся, если снова сделать файл объектом.' +
        (node.asset_category ? ' Файл будет убран из каталога 3D-ассетов.' : ''),
      action: 'Сделать сценарием',
      danger: false,
    })
    if (ok) await save({ file_type: 'scenario', asset_category: '' }).catch(() => {})
  }

  return (
    <div className="page">
      <nav className="breadcrumbs" aria-label="Путь">
        <Link to="/">Обзор</Link>
        {node.path.map((p) => (
          <span key={p.id}>
            <span className="sep">/</span>
            <Link to={`/node/${p.id}`}>{p.name}</Link>
          </span>
        ))}
      </nav>

      <header className="page-header">
        <div className="title-block">
          <div className="title-row">
            {isFolder ? (
              <FolderIcon size={26} className="ico-folder" />
            ) : isScenario ? (
              <ScrollIcon size={26} className="ico-scenario" />
            ) : (
              <FileIcon size={26} className="ico-file" />
            )}
            <h1>{node.name}</h1>
          </div>
          {!isFolder && !isScenario && (
            <AssetCategorySelect value={node.asset_category} onChange={(asset_category) => save({ asset_category })} />
          )}
        </div>
        <div className="header-actions">
          {!isFolder &&
            (isScenario ? (
              <button className="btn ghost" onClick={() => save({ file_type: 'object' }).catch(() => {})}>
                <FileIcon /> Сделать объектом
              </button>
            ) : (
              <button className="btn ghost" onClick={makeScenario}>
                <ScrollIcon /> Сделать сценарием
              </button>
            ))}
          <button className="btn ghost" onClick={() => renameNode(node.id, node.name)}>
            <EditIcon /> Переименовать
          </button>
          <button className="btn ghost danger" onClick={() => deleteNode(node.id, node.name, node.kind)}>
            <TrashIcon /> Удалить
          </button>
        </div>
      </header>

      {isFolder ? (
        <>
          <section className="panel">
            <EditableText
              label="Описание папки"
              value={node.description}
              placeholder="Коротко: что хранится в этой папке. Показывается на карточке."
              onSave={(description) => save({ description })}
            />
          </section>
          <h2 className="section-title">Содержимое</h2>
          <NodeCards nodes={treeNode?.children ?? []} parentId={node.id} />
        </>
      ) : isScenario ? (
        <ScenarioDocument value={node.description} onSave={(description) => save({ description })} />
      ) : (
        <div className="file-layout">
          <div className="media-row">
            <section className="panel">
              <ReferenceGallery node={node} onChanged={reload} />
            </section>
            <section className="panel">
              <ModelPanel node={node} onChanged={reload} />
            </section>
          </div>
          <div className="file-details">
            <section className="panel">
              <EditableText
                label="Описание"
                value={node.description}
                placeholder="Что это за объект: внешний вид, роль в игре, лор."
                onSave={(description) => save({ description })}
              />
            </section>
            <section className="panel">
              <CharacteristicsEditor
                value={node.characteristics}
                onSave={(characteristics) => save({ characteristics })}
              />
            </section>
            <section className="panel">
              <EditableText
                label="Механики взаимодействия"
                value={node.mechanics}
                placeholder="Как игрок взаимодействует с объектом: действия, условия, последствия."
                onSave={(mechanics) => save({ mechanics })}
              />
            </section>
            <section className="panel">
              <PromptPanel node={node} onSave={(reference_prompt) => save({ reference_prompt })} />
            </section>
          </div>
        </div>
      )}
    </div>
  )
}
