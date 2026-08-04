## Summary

Prepare the VitePress docs site for sub-path deployment at `https://www.olares.com/docs` (migrating from `docs.olares.com` at the domain root).

`base` is already driven by `BASE_URL` at build time; this PR fixes the few places that bypass VitePress base handling and updates the canonical sitemap host for the new URL.

## Changes

- **`docs/.vitepress/config.mts`**: sitemap `hostname` → `https://www.olares.com/docs/`
- **`docs/use-cases/stable-diffusion.md`** & **`docs/zh/use-cases/stable-diffusion.md`**: replace raw `<img src="/images/...">` in HTML tables with `:src="withBase('/images/...')"` so images work when `BASE_URL=/docs/` (VitePress does not rewrite bare HTML absolute paths)

## Build / deploy

Built by `release-docs` with:

```bash
BASE_URL=/docs/ npm run build
```

Paired with `release-docs` `path_prefix=/docs` and outer nginx preserving `/docs` (no trailing slash on `proxy_pass`).

## Test plan

- [ ] `BASE_URL=/docs/ npm run build` in `docs/` succeeds
- [ ] Built HTML references assets as `/docs/assets/*` (not `/assets/*`)
- [ ] Stable Diffusion gallery images load at `/docs/use-cases/stable-diffusion`
- [ ] `sitemap.xml` URLs use `https://www.olares.com/docs/...`

## Related

- `beclab/release-docs` PR: `feat/docs-subpath` (build script + docker image layout)
