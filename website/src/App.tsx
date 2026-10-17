import { Header } from './components/Header'
import { Hero } from './components/Hero'
import { Features } from './components/Features'
import { FAQ } from './components/FAQ'
import { Downloads } from './components/Downloads'
import { Footer } from './components/Footer'
import { ReleaseProvider } from './context/ReleaseContext'

export default function App() {
  return <ReleaseProvider><div id="top"><a className="skip-link" href="#main">Skip to content</a><Header /><div className="mx-auto max-w-[1152px] px-6 sm:px-8"><main id="main"><Hero /><Features /><FAQ /><Downloads /></main><Footer /></div></div></ReleaseProvider>
}
