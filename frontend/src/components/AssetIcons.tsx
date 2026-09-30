import type { ReactElement, SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement> & { size?: number }

function base({ size = 16, ...props }: IconProps) {
  return {
    width: size,
    height: size,
    viewBox: '0 0 24 24',
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: 1.8,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
    'aria-hidden': true,
    ...props,
  }
}

/** Generic 3D asset marker (cube). */
export const CubeIcon = (p: IconProps) => (
  <svg {...base(p)}><path d="m12 3 8 4.5v9L12 21l-8-4.5v-9z" /><path d="m4 7.5 8 4.5 8-4.5M12 12v9" /></svg>
)

const WeaponIcon = (p: IconProps) => (
  <svg {...base(p)}><path d="M14.5 4H20v5.5L9 20.5 3.5 15z" /><path d="m7 17-3 3M5.5 13.5l5 5" /></svg>
)
const SiegeIcon = (p: IconProps) => (
  <svg {...base(p)}><circle cx="7" cy="18" r="2.5" /><circle cx="17" cy="18" r="2.5" /><path d="M4 15h16M9.5 15 17 5M17 5l2.5 1.5M6 15l3-6" /></svg>
)
const MapIcon = (p: IconProps) => (
  <svg {...base(p)}><path d="M4 21V9l4-3 4 3 4-3 4 3v12" /><path d="M4 13h16M9 21v-4h6v4M8 6V3M16 6V3" /></svg>
)
const CharacterIcon = (p: IconProps) => (
  <svg {...base(p)}><circle cx="12" cy="7" r="3.5" /><path d="M5 21a7 7 0 0 1 14 0" /></svg>
)
const ItemIcon = (p: IconProps) => (
  <svg {...base(p)}><path d="M4 10h16v9a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1z" /><path d="M4 10V8a4 4 0 0 1 4-4h8a4 4 0 0 1 4 4v2M11 13h2v3h-2z" /></svg>
)
const PropIcon = (p: IconProps) => (
  <svg {...base(p)}><ellipse cx="12" cy="5" rx="6" ry="2" /><path d="M6 5v14c0 1.1 2.7 2 6 2s6-.9 6-2V5M6 12c0 1.1 2.7 2 6 2s6-.9 6-2" /></svg>
)
const GearIcon = (p: IconProps) => (
  <svg {...base(p)}><path d="M5 11a7 7 0 0 1 14 0v6l-3 3H8l-3-3z" /><path d="M12 4v16M8 12h8" /></svg>
)

const ICONS: Record<string, (p: IconProps) => ReactElement> = {
  weapon: WeaponIcon,
  siege: SiegeIcon,
  map: MapIcon,
  character: CharacterIcon,
  item: ItemIcon,
  prop: PropIcon,
  gear: GearIcon,
}

export function AssetCategoryIcon({ category, ...props }: IconProps & { category: string }) {
  const Icon = ICONS[category] ?? CubeIcon
  return <Icon {...props} />
}
