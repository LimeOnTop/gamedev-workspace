import type { DetailedHTMLProps, HTMLAttributes } from 'react'

// <model-viewer> attributes used by ModelPanel; values are passed as HTML attributes.
type ModelViewerAttributes = DetailedHTMLProps<HTMLAttributes<HTMLElement>, HTMLElement> & {
  src?: string
  alt?: string
  'camera-controls'?: string
  'camera-orbit'?: string
  'touch-action'?: string
  'interaction-prompt'?: string
  'environment-image'?: string
  'shadow-intensity'?: string
  'shadow-softness'?: string
  exposure?: string
  'auto-rotate'?: string
  'auto-rotate-delay'?: string
  'rotation-per-second'?: string
}

declare module 'react' {
  namespace JSX {
    interface IntrinsicElements {
      'model-viewer': ModelViewerAttributes
    }
  }
}
