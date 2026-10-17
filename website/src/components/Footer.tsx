import { repository } from '../lib/releases'
import { Brand } from './Brand'

export function Footer() {
  return <footer className="flex flex-wrap items-center justify-between gap-6 py-12"><Brand /><div className="flex flex-wrap items-center gap-x-6 gap-y-3 text-[13px] text-muted"><span className="inline-flex items-center gap-1.5"><span>© {new Date().getFullYear()}</span><a className="nav-link" href="https://www.seths.app/">www.seths.app</a></span><a className="nav-link" href={repository}>GitHub</a></div></footer>
}
