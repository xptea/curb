import { readFile, writeFile, rm, copyFile } from 'node:fs/promises'
import { render } from '../.prerender/entry-server.js'

const index = new URL('../dist/index.html', import.meta.url)
const template = await readFile(index, 'utf8')
if (!template.includes('<!--app-html-->')) throw new Error('Missing static HTML placeholder')
await writeFile(index, template.replace('<!--app-html-->', () => render()))
await copyFile(new URL('../public/404.html', import.meta.url), new URL('../dist/404.html', import.meta.url))
await rm(new URL('../.prerender', import.meta.url), { recursive: true, force: true })
console.log('Pre-rendered static HTML. No server-side runtime is required.')
