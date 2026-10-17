import type { ReactNode, SVGProps } from 'react'
import type { OperatingSystem } from '../lib/releases'

type IconName = 'arrow-right' | 'arrow-up' | 'arrow-down' | 'external' | 'chevron' | 'plus' | 'audio' | 'quality' | 'folder' | 'github' | 'question' | 'home' | 'info'
const paths: Record<IconName, ReactNode> = {
  home: <><path d="m3 10 9-7 9 7v11h-6v-7H9v7H3Z" /></>,
  info: <><circle cx="12" cy="12" r="9" /><path d="M12 11v6m0-10h.01" /></>,
  github: <><path d="M9 19c-4 1-4-2-6-2m12 5v-4a3.5 3.5 0 0 0-1-2.7c3.3-.4 6.7-1.6 6.7-7.3A5.7 5.7 0 0 0 19.2 4c.2-1 .1-2-.3-3 0 0-1.2-.4-4 1.5a13.5 13.5 0 0 0-7 0C5.2.6 4 1 4 1c-.4 1-.5 2-.3 3A5.7 5.7 0 0 0 2.2 8c0 5.7 3.4 6.9 6.7 7.3A3.5 3.5 0 0 0 8 18v4" transform="translate(1 1) scale(.9)" /></>,
  question: <><circle cx="12" cy="12" r="9" /><path d="M9.5 9a2.5 2.5 0 0 1 5 .5c0 1.5-2.5 2-2.5 3.5m0 3h.01" /></>,
  'arrow-right': <><path d="M4 12h16m-6-6 6 6-6 6" /></>,
  'arrow-up': <><path d="M12 20V4m-6 6 6-6 6 6" /></>,
  'arrow-down': <><path d="M12 4v16m-6-6 6 6 6-6" /></>,
  external: <><path d="M7 17 17 7M7 7h10v10" /></>,
  chevron: <path d="m7 10 5 5 5-5" />,
  plus: <path d="M5 12h14M12 5v14" />,
  audio: <><path d="M9 18V5l12-2v13M9 9l12-2" /><ellipse cx="6" cy="18" rx="3" ry="3" /><ellipse cx="18" cy="16" rx="3" ry="3" /></>,
  quality: <><path d="M4 6h16M4 12h16M4 18h16" /><circle cx="8" cy="6" r="2" /><circle cx="16" cy="12" r="2" /><circle cx="10" cy="18" r="2" /></>,
  folder: <path d="M3 7a2 2 0 0 1 2-2h5l2 3h7a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z" />,
}

export function Icon({ name, ...props }: SVGProps<SVGSVGElement> & { name: IconName }) {
  return <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>{paths[name]}</svg>
}

export function PlatformIcon({ os, ...props }: SVGProps<SVGSVGElement> & { os: OperatingSystem }) {
  if (os === 'windows') return <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" {...props}><path fill="#47a8f5" d="M2 4 11 2.8v8.4H2Zm10.5-1.4L22 1v10.2h-9.5ZM2 12.7h9V21L2 19.8Zm10.5 0H22V23l-9.5-1.6Z" /></svg>
  // Apple silhouette from Simple Icons (CC0): github.com/simple-icons/simple-icons.
  if (os === 'mac') return <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true" {...props}><path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701" /></svg>
  if (os === 'linux') return <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" {...props}><path fill="#e8e8e8" d="M8 8C7 3 9 1 12 1s5 2 4 7c3 3 4 7 2 11H6C4 15 5 11 8 8Z" /><path fill="#252525" d="M8 9c1-2 1-3 1-4 0-2 1-3 3-3s3 1 3 3c0 1 0 2 1 4 1 2 2 5 1 8l-2 2H9l-2-2c-1-3 0-6 1-8Z" /><ellipse cx="12" cy="14" rx="4" ry="6" fill="#f4f4f4" /><ellipse cx="10.4" cy="6" rx="1" ry="1.5" fill="white" /><ellipse cx="13.6" cy="6" rx="1" ry="1.5" fill="white" /><circle cx="10.7" cy="6.3" r=".45" /><circle cx="13.3" cy="6.3" r=".45" /><path fill="#edbf45" d="m9 8 3-1 3 1-3 2ZM4 19l4-2 3 4-6 1-2-1Zm16 0-4-2-3 4 6 1 2-1Z" /></svg>
  return <Icon {...props} name="arrow-down" />
}
