import { useRelease } from '../context/ReleaseContext'
import { assetSize, assetUrl, releasePage } from '../lib/releases'
import { PlatformIcon } from './Icon'
import { MacDownloadButton } from './MacDownloadButton'

export function Downloads() {
  const release = useRelease()
  return <section id="download" aria-labelledby="download-title" className="download-section">
    <div className="text-center"><h2 id="download-title" className="section-title">Download Curb</h2><a href={releasePage} className="mt-3 inline-block text-[13px] text-muted transition-colors hover:text-white" aria-live="polite">{release ? `v${release.version}` : 'Latest release'} · Free &amp; open source</a></div>
    <div className="download-grid">
      <article className="download-platform"><h3><PlatformIcon os="windows" width="22" height="22" />Windows</h3><p>64-bit ZIP{release && ` · ${assetSize('windows-amd64', release)}`}</p><a className="platform-download" aria-label="Download Curb for Windows" href={assetUrl('windows-amd64', release)}>Download</a></article>
      <article className="download-platform"><h3><PlatformIcon os="mac" width="22" height="22" />macOS</h3><p>macOS 15 or later</p><MacDownloadButton /></article>
      <article className="download-platform"><h3><PlatformIcon os="linux" width="22" height="22" />Linux</h3><p>64-bit AppImage{release && ` · ${assetSize('linux-amd64', release)}`}</p><a className="platform-download" aria-label="Download Curb for Linux" href={assetUrl('linux-amd64', release)}>Download</a></article>
    </div>
  </section>
}
