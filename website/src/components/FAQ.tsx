import { repository, supportedSites } from '../lib/releases'
import { Icon } from './Icon'

const questions = [
  { question: 'What is Curb?', answer: <>Curb is a native desktop app for downloading videos and MP3 audio. It brings yt-dlp into a simple interface for Windows, macOS, and Linux.</> },
  { question: 'Which sites can I download from?', answer: <>Curb uses yt-dlp’s site support. You can search supported sites inside the app, or browse <a href={supportedSites}>the public list</a>. Availability can change as sites update.</> },
  { question: 'Where do my downloads go?', answer: <>Straight to your computer. Curb starts with your Downloads folder, and you can choose a different default location in Settings.</> },
  { question: 'Is it free?', answer: <>Yes. Curb is free, and its own source code is available under the MIT license. You can <a href={repository}>explore it on GitHub</a>.</> },
  { question: 'How do updates work?', answer: <>When a new release is available, the version in the bottom-left corner becomes an update button. Click it to download the update and restart Curb. Your settings stay in place.</> },
]

export function FAQ() {
  return <section id="faq" aria-labelledby="faq-title" className="section-space mx-auto max-w-[740px]">
    <h2 id="faq-title" className="section-title text-center">A few questions.</h2>
    <div className="mt-9">{questions.map(item => <details className="faq-item group" key={item.question}><summary className="flex cursor-pointer list-none items-center justify-between gap-6 px-5 py-5 text-[15px] text-[#deddd9] transition-colors hover:text-white"><span>{item.question}</span><Icon name="plus" className="shrink-0 text-[#aaa8a4] transition-transform duration-200 group-open:rotate-45" /></summary><div className="faq-answer max-w-[720px] px-5 pr-10 pb-5 text-[14px] leading-[1.8] text-muted">{item.answer}</div></details>)}</div>
  </section>
}
