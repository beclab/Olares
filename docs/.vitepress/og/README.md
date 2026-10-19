# Document share images

Every rendered page gets a 1200 × 630 PNG using the approved light design:
a softly shaded gray dot background, the black Olares logo, black title,
gray description, documentation URL, and a 12 px black strip along the bottom.
Text is placed directly on the background, without a separate content panel.
There is no category label. The header, text, and footer have 72 px side margins.

`background.svg` contains the background texture and bottom strip. `logo.svg`
is the original black brand logo. `generate.mjs` controls text layout and export.

`transformHead` registers the image and inserts Open Graph and Twitter metadata.
`buildEnd` writes the PNG files into `.vitepress/dist/og/`, including archived
and noindex pages. Deploy the complete build output as usual. No image service,
browser installation, external font request, or manual image upload is needed.

Images are named using a hash of the page path, title, image description,
language, renderer, background, logo, and font files. Generated files are cached
in `.vitepress/cache/og/` and copied
back into each fresh build. The cache and output are already ignored by Git.
Changes to text or design produce new image URLs. Social platforms may still
cache the page's metadata until they recrawl it.

The production URL defaults to `https://www.olares.com/docs/`. A non-root
`BASE_URL` is preserved, including versioned deployments. For a separately
hosted preview, set `DOCS_SITE_URL` to its full public site URL, including the
base path (for example `https://preview.example.com/docs/`).

The page title comes from VitePress (frontmatter title or the first heading).
For long text, set shorter image-only copy in frontmatter:

```yaml
ogTitle: Install and activate Olares
ogDescription: Choose an installation method and set up Olares on your hardware.
```

Descriptions come from the page's `description`, with the localized site tagline
as a fallback. Optional `ogDescription` overrides the image text only (an empty
string hides it). The full page description remains in Open Graph/Twitter metadata.

Titles use Manrope at weight 650 for Latin text and Noto Sans SC at weight 550
for Chinese text. Chinese pages use 0 px title tracking; English pages use
-0.6 px. Descriptions use the same font families at weight 400.

Titles use the full 1056 px text width, with a 184 px height budget, and wrap
into up to three lines, shrinking from 64 px when necessary. Descriptions use
29 px text in a separate area, at least 36 px below the title, with up to three lines. Overlong
image descriptions end with an ellipsis; neither the title nor description can
overlap the footer. Lines fill the available width naturally rather than being
balanced into a narrower column, matching the selected English and Chinese previews.
An unrenderable or excessively long title fails the build instead of publishing
a broken image. Metadata is inserted during the production build; inspect it
with `npm run build` followed by `npm run preview`, not the development server.

Run focused checks with `npm run test:og`, and validate integration with
`npm run build`.

## Fonts and logo

Manrope and Noto Sans SC are unmodified variable fonts from the Google Fonts
repository. Manrope was downloaded on 2026-10-19; Noto Sans SC on 2026-10-16.
Their complete SIL OFL 1.1 licenses and copyright notices are preserved beside the font files in `fonts/`:

- https://github.com/google/fonts/tree/main/ofl/manrope
- https://github.com/google/fonts/tree/main/ofl/notosanssc

The license allows commercial use and embedding. Keep these license files when
redistributing the fonts. The generated images do not need to use the OFL.
Fonts are build inputs only and are not copied to the deployed website.
The black logo is the original SVG supplied by the Olares brand owner.
