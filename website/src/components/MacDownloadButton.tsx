import { useRelease } from '../context/ReleaseContext'
import { useEffect, useRef } from 'react'
import { assetSize, assetUrl } from '../lib/releases'

export function MacDownloadButton() {
  const release = useRelease()
  const menu = useRef<HTMLDetailsElement>(null)

  useEffect(() => {
    const outside = (event: PointerEvent) => {
      if (menu.current && !menu.current.contains(event.target as Node)) menu.current.open = false
    }
    const escape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && menu.current?.open) {
        menu.current.open = false
        menu.current.querySelector('summary')?.focus()
      }
    }
    document.addEventListener('pointerdown', outside)
    document.addEventListener('keydown', escape)
    return () => {
      document.removeEventListener('pointerdown', outside)
      document.removeEventListener('keydown', escape)
    }
  }, [])

  return <details ref={menu} className="mac-download">
    <summary className="platform-download" aria-label="Choose macOS architecture">Download</summary>
    <div className="mac-download-menu">
      <a href={assetUrl('darwin-arm64', release)} onClick={() => { if (menu.current) menu.current.open = false }}><span>Apple silicon</span><span>{assetSize('darwin-arm64', release)}</span></a>
      <a href={assetUrl('darwin-amd64', release)} onClick={() => { if (menu.current) menu.current.open = false }}><span>Intel</span><span>{assetSize('darwin-amd64', release)}</span></a>
    </div>
  </details>
}
