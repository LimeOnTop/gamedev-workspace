export interface Theme {
  id: string
  name: string
  /** Background, panel and accent colours used for the picker swatch. */
  swatch: [string, string, string]
}

// Keep in sync with src/themes.css.
export const THEMES: Theme[] = [
  { id: 'ember', name: 'Угли', swatch: ['#120e0c', '#2c211b', '#ff8a3d'] },
  { id: 'abyss', name: 'Бездна', swatch: ['#070d14', '#172638', '#22d3ee'] },
  { id: 'moss', name: 'Мох', swatch: ['#0b100d', '#1c2820', '#4ade80'] },
  { id: 'amethyst', name: 'Аметист', swatch: ['#0f0b16', '#261c36', '#c084fc'] },
  { id: 'crimson', name: 'Пепел и кровь', swatch: ['#0b0b0c', '#232326', '#ef4444'] },
  { id: 'graphite', name: 'Графит', swatch: ['#0f1115', '#232835', '#7c9cff'] },
]

export const DEFAULT_THEME = 'ember'
// index.html reads the same key before first paint to avoid a colour flash.
const STORAGE_KEY = 'workspace.theme'

export function loadTheme(): string {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved && THEMES.some((t) => t.id === saved)) return saved
  } catch {
    // storage unavailable
  }
  return DEFAULT_THEME
}

export function applyTheme(id: string) {
  document.documentElement.dataset.theme = id
  const meta = document.querySelector('meta[name="theme-color"]')
  const theme = THEMES.find((t) => t.id === id)
  if (meta && theme) meta.setAttribute('content', theme.swatch[0])
  try {
    localStorage.setItem(STORAGE_KEY, id)
  } catch {
    // storage unavailable
  }
}
