import { useRelease } from '../context/ReleaseContext'
import { useEffect, useRef, useState } from 'react'
import { assetSize, assetUrl, detectOS, platforms, type OperatingSystem } from '../lib/releases'
import { Icon, PlatformIcon } from './Icon'

export function DownloadButton({ compact = false }: { compact?: boolean }) {
  const release = useRelease()
  const [os, setOS] = useState<OperatingSystem>('other')
  const menu = useRef<HTMLDetailsElement>(null)

  useEffect(() => {
    setOS(detectOS())
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

  // Browsers cannot reliably distinguish Intel Macs from Apple silicon.
  // Mac visitors choose their architecture instead of receiving a guessed build.
  const primary = os === 'windows' ? assetUrl('windows-amd64', release) : os === 'linux' ? assetUrl('linux-amd64', release) : '#download'
  const label = os === 'windows' ? 'Download for Windows' : os === 'linux' ? 'Download for Linux' : os === 'mac' ? 'Download for macOS' : 'Download Curb'

  return <div className={`download-control ${compact ? 'is-compact' : ''}`}>
    <a href={primary} className="download-primary" aria-label={label}>
      <PlatformIcon os={os} className="shrink-0" />
      <span>{compact ? 'Download' : label}</span>
    </a>
    <details className="download-options" ref={menu}>
      <summary aria-label="Choose download platform"><Icon name="chevron" width="16" height="16" /></summary>
      <div className="platform-menu">
        <p className="px-3 pt-2 pb-2.5 text-[11px] tracking-[.1em] text-muted uppercase">Choose your platform</p>
        {platforms.map(platform => <a key={platform.id} href={assetUrl(platform.id, release)} aria-label={`Download Curb for ${platform.name}`} className="platform-option" onClick={() => { if (menu.current) menu.current.open = false }}>
          <PlatformIcon os={platform.os} />
          <span className="flex-1"><span className="block text-[13px] text-white">{platform.name}</span><span className="block pt-0.5 text-[11px] text-muted">{platform.description}</span></span>
          <span className="text-[11px] text-muted">{assetSize(platform.id, release)}</span>
        </a>)}
      </div>
    </details>
  </div>
}
