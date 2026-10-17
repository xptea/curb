export type Platform = 'windows-amd64' | 'darwin-arm64' | 'darwin-amd64' | 'linux-amd64'
export type OperatingSystem = 'windows' | 'mac' | 'linux' | 'other'
type Asset = { name: string; size: number; sha256: string; format: string }
export type Release = { schema: 1; version: string; platforms: Record<Platform, Asset> }

export const repository = 'https://github.com/xptea/curb'
export const releasePage = `${repository}/releases/latest`
export const releaseFeed = 'https://raw.githubusercontent.com/xptea/curb/main/version.json'
export const supportedSites = 'https://github.com/yt-dlp/yt-dlp/blob/master/supportedsites.md'

export const platforms: { id: Platform; os: OperatingSystem; name: string; description: string }[] = [
  { id: 'windows-amd64', os: 'windows', name: 'Windows', description: '64-bit · Portable ZIP' },
  { id: 'darwin-arm64', os: 'mac', name: 'macOS · Apple silicon', description: 'macOS 15 or later' },
  { id: 'darwin-amd64', os: 'mac', name: 'macOS · Intel', description: 'macOS 15 or later' },
  { id: 'linux-amd64', os: 'linux', name: 'Linux', description: '64-bit · AppImage' },
]

export function parseRelease(value: unknown): Release {
  if (!value || typeof value !== 'object') throw new Error('Invalid release feed')
  const data = value as Record<string, unknown>
  if (data.schema !== 1 || typeof data.version !== 'string' || !/^\d+\.\d+\.\d+(?:-[\da-z.-]+)?(?:\+[\da-z.-]+)?$/i.test(data.version) || !data.platforms || typeof data.platforms !== 'object') throw new Error('Invalid release feed')
  const assets = data.platforms as Record<string, unknown>
  const checked = {} as Record<Platform, Asset>
  for (const { id } of platforms) {
    const value = assets[id]
    if (!value || typeof value !== 'object') throw new Error(`Missing platform: ${id}`)
    const asset = value as Record<string, unknown>
    if (typeof asset.name !== 'string' || !asset.name.startsWith(`Curb-${data.version}-`) || !/^[\da-z._+-]+$/i.test(asset.name) || typeof asset.size !== 'number' || !Number.isSafeInteger(asset.size) || asset.size <= 0 || typeof asset.sha256 !== 'string' || !/^[\da-f]{64}$/i.test(asset.sha256) || typeof asset.format !== 'string') throw new Error(`Invalid platform: ${id}`)
    checked[id] = { name: asset.name, size: asset.size, sha256: asset.sha256, format: asset.format }
  }
  return { schema: 1, version: data.version, platforms: checked }
}

export async function fetchRelease(signal?: AbortSignal): Promise<Release> {
  const response = await fetch(releaseFeed, { cache: 'no-store', signal })
  if (!response.ok) throw new Error(`Release feed returned ${response.status}`)
  return parseRelease(await response.json())
}

export function assetUrl(platform: Platform, release: Release | null) {
  const asset = release?.platforms[platform]
  return release && asset ? `${repository}/releases/download/v${release.version}/${encodeURIComponent(asset.name)}` : releasePage
}

export function assetSize(platform: Platform, release: Release | null) {
  const asset = release?.platforms[platform]
  return asset ? `${(asset.size / 1024 / 1024).toFixed(1)} MB` : ''
}

export function detectOS(): OperatingSystem {
  const platform = navigator.platform.toLowerCase()
  if (platform.includes('win')) return 'windows'
  if (platform.includes('mac')) return 'mac'
  if (platform.includes('linux') && !/android/i.test(navigator.userAgent)) return 'linux'
  return 'other'
}
