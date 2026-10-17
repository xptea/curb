import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { fetchRelease, type Release } from '../lib/releases'

const ReleaseContext = createContext<Release | null>(null)

export function ReleaseProvider({ children }: { children: ReactNode }) {
  const [release, setRelease] = useState<Release | null>(null)

  useEffect(() => {
    let disposed = false
    let pending = false
    let lastAttempt = 0
    let controller: AbortController | undefined
    const refresh = async () => {
      if (document.visibilityState === 'hidden' || pending || Date.now() - lastAttempt < 60_000) return
      pending = true
      lastAttempt = Date.now()
      controller = new AbortController()
      const timeout = window.setTimeout(() => controller?.abort(), 10_000)
      try {
        const latest = await fetchRelease(controller.signal)
        if (!disposed) setRelease(latest)
      } catch {
        // Keep a previously verified release, or use GitHub's latest-release page.
      } finally {
        window.clearTimeout(timeout)
        pending = false
      }
    }
    void refresh()
    const timer = window.setInterval(() => { void refresh() }, 5 * 60_000)
    const onReturn = () => { void refresh() }
    window.addEventListener('focus', onReturn)
    document.addEventListener('visibilitychange', onReturn)
    return () => {
      disposed = true
      controller?.abort()
      window.clearInterval(timer)
      window.removeEventListener('focus', onReturn)
      document.removeEventListener('visibilitychange', onReturn)
    }
  }, [])

  return <ReleaseContext.Provider value={release}>{children}</ReleaseContext.Provider>
}

export function useRelease() { return useContext(ReleaseContext) }
