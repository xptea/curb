# Curb landing page

A React + TypeScript + Tailwind CSS page inspired by Tella, using a softer charcoal palette, rounded controls, and borderless sections. The supplied app screenshot and Curb icon are bundled locally, along with the Geist font. There are no analytics, external fonts, or backend services. The browser loads public release metadata from GitHub Raw.

## Develop

Use Node.js 22.12 or newer:

```sh
cd website
npm ci
npm run dev
```

The page includes a platform download selector, responsive layouts, native FAQ disclosures, and keyboard navigation. Mac visitors explicitly choose Apple silicon or Intel; browsers cannot reliably detect that distinction.

## Static production build

```sh
npm run build
npm run preview
```

The build pre-renders React into `dist/index.html`, then removes the temporary server build. The output is only HTML, CSS, browser JavaScript, fonts, and images. Without JavaScript, the download controls link to GitHub's latest-release page. React loads the current version, file sizes, and direct platform links in the browser, and adds platform detection and dropdown dismissal.

Download URLs, file sizes, and the displayed version load dynamically from [the public version feed](https://raw.githubusercontent.com/xptea/curb/main/version.json), which the release workflow updates after publishing. **No website rebuild or redeploy is needed for new desktop releases.** A shared browser request loads the feed when the page opens; visible pages recheck every five minutes and when the tab regains focus (at most once per minute). GitHub's CDN may cache the feed for up to five minutes.

The feed is validated before any controls update, so all links and version text change together. If the initial check fails or times out, the page shows “Latest release” and links to GitHub's latest-release page. If a later check fails, the last verified release stays available. Private-repository credentials are never embedded in the site. `public/_headers` permits release requests only to GitHub Raw in addition to the site's own origin.

Run `npm test` for release parsing, future-version download links, and failure handling checks.

## Cloudflare Workers

`wrangler.jsonc` configures a Worker with **only static assets**, no `main` script, APIs, databases, or server rendering. Cloudflare currently serves static asset requests for free without a request quota; build/deployment platform limits still apply. See [Cloudflare's static asset billing documentation](https://developers.cloudflare.com/workers/static-assets/billing-and-limitations/).

From this folder:

```sh
npx wrangler login
npm run deploy
```

For Cloudflare's Git integration, use:

| Setting | Value |
| --- | --- |
| Root directory | `website` |
| Build command | `npm ci && npm run build` |
| Deploy command | `npx wrangler deploy` |

Add a custom domain to the `curb-website` Worker in Cloudflare when ready. Only changes under `website/**` need to rebuild the landing page; release metadata is loaded dynamically. Do not configure a catch-all Worker function; unmatched URLs use the included static 404 page.

`public/_headers` adds asset caching and basic browser security headers. Hashed build assets are cached for a year; the home page revalidates. App downloads go directly to GitHub, keeping release binaries out of the Worker.

## Files

```text
src/components/     Header, hero, platform selector, features, FAQ, downloads, footer
src/lib/releases.ts Public release feed validation and download links
src/context/        Shared live release state and refresh schedule
src/styles.css      Tailwind and shared visual styles
scripts/            Build-time HTML pre-rendering
public/             Supplied screenshot, app icons, headers, and 404 page
wrangler.jsonc      Static Cloudflare deployment configuration
```
