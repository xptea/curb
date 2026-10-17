import { strict as assert } from 'node:assert'
import { test } from 'node:test'
import { assetSize, assetUrl, fetchRelease, parseRelease, platforms, releaseFeed, releasePage } from './releases.ts'

function feed(version = '9.8.7') {
  return {
    schema: 1,
    version,
    platforms: Object.fromEntries(platforms.map(({ id }) => [id, {
      name: `Curb-${version}-${id}.zip`, size: 10 * 1024 * 1024, sha256: 'a'.repeat(64), format: 'zip',
    }])),
  }
}

test('a newly published version supplies every download URL and size', () => {
  const release = parseRelease(feed())
  assert.equal(release.version, '9.8.7')
  for (const { id } of platforms) {
    assert.equal(assetUrl(id, release), `https://github.com/xptea/curb/releases/download/v9.8.7/Curb-9.8.7-${id}.zip`)
    assert.equal(assetSize(id, release), '10.0 MB')
  }
})

test('before a successful check, downloads use the latest-release page', () => {
  for (const { id } of platforms) {
    assert.equal(assetUrl(id, null), releasePage)
    assert.equal(assetSize(id, null), '')
  }
})

test('malformed, incomplete, or mixed-version feeds cannot update the page', () => {
  assert.throws(() => parseRelease(null))
  assert.throws(() => parseRelease({ ...feed(), schema: 2 }))
  assert.throws(() => parseRelease({ ...feed(), platforms: {} }))
  for (const name of ['../outside.zip', 'Curb-0.0.1-windows.zip']) {
    const malformed = feed()
    malformed.platforms['windows-amd64'].name = name
    assert.throws(() => parseRelease(malformed))
  }
  const malformed = feed()
  malformed.platforms['linux-amd64'].size = -1
  assert.throws(() => parseRelease(malformed))
})

test('the browser loader reads the live feed without using a cached build version', async t => {
  const signal = new AbortController().signal
  t.mock.method(globalThis, 'fetch', async (url: string, options: RequestInit) => {
    assert.equal(url, releaseFeed)
    assert.equal(options.cache, 'no-store')
    assert.equal(options.signal, signal)
    return new Response(JSON.stringify(feed('2.3.4')))
  })
  assert.equal((await fetchRelease(signal)).version, '2.3.4')
})

test('HTTP failures and invalid JSON reject instead of creating broken downloads', async t => {
  const mocked = t.mock.method(globalThis, 'fetch', async () => new Response('Unavailable', { status: 503 }))
  await assert.rejects(fetchRelease(), /503/)
  mocked.mock.mockImplementation(async () => new Response('not JSON'))
  await assert.rejects(fetchRelease())
})
