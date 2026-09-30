import type { NodeDetail } from '../api'
import { useWorkspace } from '../workspace'
import EditableText from './EditableText'
import { CopyIcon, SparkIcon } from './Icons'

const MAX_DESCRIPTION = 400

/**
 * Builds a starting prompt from the object's card. Image models understand
 * English best, so the frame is English while the card details are kept as-is.
 */
export function buildPromptDraft(node: NodeDetail): string {
  const description = node.description.replace(/\s+/g, ' ').trim()
  const details = node.characteristics
    .filter((c) => c.key && c.value)
    .slice(0, 8)
    .map((c) => `${c.key}: ${c.value}`)
    .join('; ')

  return [
    `Concept art reference of "${node.name}" for a medieval castle-siege game, single object, game-ready 3D asset reference.`,
    description && `Description: ${description.length > MAX_DESCRIPTION ? `${description.slice(0, MAX_DESCRIPTION)}…` : description}`,
    details && `Key details: ${details}.`,
    'Stylized semi-realistic look, readable silhouette, clear materials (wood, iron, stone, leather), subtle wear.',
    'Three-quarter view, neutral grey background, soft even studio lighting, full object in frame, no text, no watermark, high detail.',
  ]
    .filter(Boolean)
    .join('\n')
}

export default function PromptPanel({ node, onSave }: { node: NodeDetail; onSave: (prompt: string) => Promise<void> }) {
  const { notify } = useWorkspace()

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(node.reference_prompt)
      notify('Промпт скопирован')
    } catch {
      notify('Не удалось скопировать — выделите текст вручную')
    }
  }

  return (
    <EditableText
      className="prompt-panel"
      label={<><SparkIcon className="ico-accent" /> Промпт для генерации референсов</>}
      value={node.reference_prompt}
      placeholder="Опишите объект для нейросети (Midjourney, SDXL, DALL·E…): форма, материалы, цвета, стиль, ракурс. Лучше на английском."
      hint={!node.reference_prompt ? 'Можно собрать черновик из описания и характеристик или попросить Claude написать промпт через MCP.' : undefined}
      draft={{ label: node.reference_prompt ? 'Пересобрать' : 'Собрать черновик', build: () => buildPromptDraft(node) }}
      actions={
        node.reference_prompt ? (
          <button className="btn ghost small" onClick={copy} title="Скопировать промпт">
            <CopyIcon /> Копировать
          </button>
        ) : null
      }
      onSave={onSave}
    />
  )
}
