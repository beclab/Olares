# Produce the app icon

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This file is *how to make* `metadata.icon`. Where it sits among the requirement tiers is in the market-ready targets from the parent SKILL.

`metadata.icon` is the only listing asset the platform **requires**: `chart lint` fails without it. Note that `lint` only checks the URL is well-formed `http(s)` — a typo'd host passes validation and shows a broken image in the Market, so confirm the URL actually resolves before submitting.

Target: **PNG or WEBP, 256x256, <=512 KB.** The default CDN icon is fine for local deploy and wrong for a public listing.

Runs entirely offline after fetching: no API key, no hosted model. [`../scripts/generate_icon.py`](../scripts/generate_icon.py) does the compositing.

## Rule zero: never generate the icon

The icon must be the project's **real** mark, fetched from its GitHub repo or official site. An invented or AI-drawn icon misrepresents the upstream project in a public catalog. If no icon can be found, **stop and ask the user** — do not substitute one.

## Step 1: fetch the source

Search in this order, taking the first usable hit:

1. Logo/icon files in the GitHub repo — usually `/`, `assets/`, `docs/`, `.github/`
2. Logo images referenced from the README
3. Official site favicon (SVG favicon first)
4. Official site `og:image` or `apple-touch-icon`
5. Logo element in the page markup

Format preference: **SVG > PNG > JPG > ICO**. Save as `source_icon.svg` / `source_icon.png` before processing.

### Keep the mark, drop the wordmark

The icon must be a **single graphic with no text**.

| What you fetched | What to do |
|---|---|
| Pure graphic mark | Use directly |
| Graphic + text side by side (wordmark) | Crop to the graphic only — find the graphic's bounding box, crop, re-center |
| Graphic and text overlapping (can't crop) | Re-fetch from `favicon` / `apple-touch-icon` — those are almost always the pure mark |
| Text-only logo, no graphic at all | **Ask the user** whether to use it; if yes, treat as a colored background (fill 256x256) |

### Resolution floor: 200px

Check resolution immediately after fetching. Anything **< 200px is unusable** — scaling it up to the 180px safe zone shows visible blur. Climb this ladder instead:

1. Favicon too small (e.g. 48px) -> look for the project's Flutter / mobile repo, which carries 192–512px launcher icons
2. Main repo has nothing high-res -> try `android/app/src/main/res/mipmap-xxxhdpi/`, `assets/images/logo.png`, `public/logo512.png`
3. Still nothing -> check Docker Hub, npm, or Flathub pages (all require a high-res icon at publish time)
4. Genuinely low-res only -> upscale with Real-ESRGAN to >=200px, then composite
5. Still blurry after upscaling, or source < 64px -> report back to the user and ask for a better source

## Step 2: classify the background

Everything downstream branches on this. The script samples a 5x5 block at each of the four corners and averages it:

| Corner reading | Type | Treatment |
|---|---|---|
| alpha < 128 | Transparent | Default gradient + centered in the safe zone |
| alpha >= 128, RGB all >= 240 | White / near-white | **Remove the white**, then gradient + safe zone |
| alpha >= 128, RGB not white | Colored | Icon carries its own background — scale to fill 256x256, safe zone does not apply |

White backgrounds must be removed, never kept — a white square on the Launchpad reads as a rendering bug. Removal uses **edge flood-fill**: only white regions connected to the border go transparent, so white *inside* the mark survives.

Two shapes that look transparent but behave as colored: a circular or rounded-square icon (macOS style), and a transparent PNG whose mark sits on its own rounded-rect plate. Both fill 256x256.

### Glow icons are the exception

Some icons report transparent corners but are designed against dark: outward glow, edge bloom, warm gradients, a light mark ringed by colored halo, or the project's own site always showing it on dark. On the light gradient those go invisible and look like they're floating.

Composite them onto solid `#000000` first, then run the script with `--force-bg` so it fills 256x256 and the glow survives.

## Step 3: composite

Constants — canvas 256x256, safe-zone margin 38px on each side (so a 180x180 centered box), vertical gradient `#FBFBFB` (top) to `#F1F1F1` (bottom). The gradient and the safe zone apply **only** to the transparent/white branch; colored icons ignore both.

```bash
python3 ../scripts/generate_icon.py source_icon.svg MyApp.png
python3 ../scripts/generate_icon.py source_icon.png MyApp.png --force-bg          # glow / known colored
python3 ../scripts/generate_icon.py source_icon.png MyApp.png --force-transparent # override misdetection
```

Name the output after the app (`MongoDB.png`, `sglang.png`). Needs `Pillow` and `numpy`, plus `cairosvg` for SVG input.

## Step 4: verify before delivery

The script self-checks and exits non-zero on failure, so a clean exit already means 256x256 / RGBA / <=512 KB. To re-check a file you did not just generate:

```bash
python3 -c "
from PIL import Image; import os,sys
p='MyApp.png'; img=Image.open(p); kb=os.path.getsize(p)/1024
assert img.size==(256,256), img.size
assert img.mode=='RGBA', img.mode
assert kb<=512, f'{kb:.0f} KB'
print(f'OK {p}: {img.size[0]}x{img.size[1]}, {img.mode}, {kb:.0f} KB')"
```

A 256x256 PNG rarely approaches 512 KB. If one does, re-export as WEBP — the platform accepts it for icons and it cuts size sharply. The check above is the PNG one; a fully opaque WEBP legitimately reports `RGB`, because the encoder drops an all-255 alpha channel. Only treat a missing alpha channel as a failure when the icon actually has transparent pixels.

## Step 5: wire it into the manifest

Upload to [imghost.olares.com](https://imghost.olares.com/) (pick **app icon**) or host it yourself, then put the URL in `metadata.icon` **and** in each `entrances[].icon` — the same URL is fine unless entrances genuinely differ:

```yaml
metadata:
  icon: https://<your-host>/myapp-icon.png
entrances:
  - name: myapp
    icon: https://<your-host>/myapp-icon.png
```

Re-run `olares-cli chart lint ./<app>` afterwards. For an app already in the Market, shipping a new icon needs an `UPDATE` PR with a version bump.

## Batch mode

For several apps at once: fetch, classify, and composite each one separately — background type is per-icon and a batch default will be wrong for some. Write all outputs to one directory, run the check on every file, then zip.
