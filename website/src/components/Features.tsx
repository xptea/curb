import { supportedSites } from '../lib/releases'
import { Icon } from './Icon'

const features = [
  { icon: 'audio' as const, title: 'Video or audio', text: 'Keep the video, or save just the audio as an MP3.' },
  { icon: 'quality' as const, title: 'Choose your quality', text: 'Pick a resolution or get the best available.' },
  { icon: 'folder' as const, title: 'Saved where you want', text: 'Set your download folder and Curb remembers it.' },
]

export function Features() {
  return <section id="features" aria-labelledby="features-title" className="section-space">
    <h2 id="features-title" className="section-title text-center">Simple from link to file.</h2>
    <div className="feature-grid">{features.map(feature => <article key={feature.title}><div className="feature-icon"><Icon name={feature.icon} width="22" height="22" /></div><h3 className="mt-5 text-[19px] font-medium tracking-[-.025em]">{feature.title}</h3><p className="mt-3 max-w-[265px] text-[15px] leading-[1.8] text-muted">{feature.text}</p></article>)}</div>
    <div className="mt-10 text-center"><a href={supportedSites} className="inline-flex items-center gap-2 text-[13px] text-muted transition-colors hover:text-white">Browse supported sites<Icon name="external" width="15" height="15" /></a></div>
  </section>
}
