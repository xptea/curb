import { repository } from '../lib/releases'
import { DownloadButton } from './DownloadButton'
import { Icon, PlatformIcon } from './Icon'

export function Hero() {
  return <>
    <section id="home" aria-labelledby="hero-title" className="hero-section">
      <h1 id="hero-title" className="hero-title">Save a video.<br /><span>Or just the audio.</span></h1>
      <p className="hero-subtitle">Paste a link, choose your quality, and download to your computer.</p>
      <div className="hero-actions"><DownloadButton /><a href={repository} className="button-secondary">GitHub<Icon name="external" width="16" height="16" /></a></div>
      <div className="hero-platforms" aria-label="Available for Windows, macOS, and Linux"><span><PlatformIcon os="windows" width="13" height="13" />Windows</span><span><PlatformIcon os="mac" width="13" height="13" />macOS</span><span><PlatformIcon os="linux" width="14" height="14" />Linux</span></div>
    </section>
    <figure id="preview" className="app-preview"><div className="screenshot-frame"><img src="/curb-screenshot.png" width="2230" height="1348" alt="Curb desktop app in light mode, with a link input and video, audio, and quality controls." fetchPriority="high" className="block h-auto w-full" /></div></figure>
  </>
}
