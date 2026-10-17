import { repository } from '../lib/releases'
import { Brand } from './Brand'
import { DownloadButton } from './DownloadButton'

export function Header() {
  return <header className="site-header">
    <Brand />
    <div className="flex items-center gap-7 sm:gap-8">
      <nav aria-label="Main navigation" className="header-nav"><a href="#features">Features</a><a href="#faq">FAQ</a><a href={repository} className="hidden sm:block">GitHub</a></nav>
      <DownloadButton compact />
    </div>
  </header>
}
