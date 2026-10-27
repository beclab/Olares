# Produce the listing images

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> This file is the material -> delivery pipeline for `spec.featuredImage` and `spec.promoteImage[]`. Composition, typography, and the render contract are in [olares-publish-listing-layout.md](olares-publish-listing-layout.md), its companion. The app icon is a separate asset with its own reference from the parent SKILL.

Target: **JPEG, PNG or WEBP, 1440x900, <=8 MB each.** Neither field is enforced by anything — they cost you installs, not the merge. Plan for **2–6 images**: one hero plus one per feature. That is what a coherent set looks like; the platform ceiling is 8.

| Image | Manifest field |
|---|---|
| Hero — overview copy, richest full-view screenshot | `spec.featuredImage` (exactly one) |
| The whole set, hero first | `spec.promoteImage[]` (>=2 recommended) |

## Step 1: collect the material

### Screenshots — shoot your own instance first

Publishing presupposes the app **already installs and reaches `running` on your Olares**. So open its entrance and capture there. Upstream marketing screenshots are the fallback, not the default: they are often stale, arbitrarily sized, and show a deployment that isn't Olares. Your own instance gives current UI at whatever resolution you want.

Fall back to the official site, then README images, only when a live capture isn't possible. Never pull frames out of a low-res GIF.

Whatever the source, screenshots must be **real UI**. Do not have a model draw a fake interface.

### Upscale to the floor

Screenshots render 700–1000px wide inside the 1440x900 frame, so the bar is "sharp enough", not "as large as possible":

| Source width | Action |
|---|---|
| >= 1000px | Use as-is |
| 500–999px | Real-ESRGAN 2x |
| < 500px | Real-ESRGAN 4x, **and** tell the user detail may be soft and a better source would help |

Real-ESRGAN is a local model (`pip install realesrgan basicsr`, `RealESRGANer` with `RealESRGAN_x4plus`) — no API key, but it pulls in torch.

### Copy — English, lifted not written

- **Headline:** the site's hero slogan, or the README's one-line description. English even for China-facing apps; it typesets better. Never write a paragraph.
- **Features:** 2–6 core points from the features page or README list. **This count decides how many images you make.**
- **Subhead:** optional, distilled from the feature point. Drop it if the frame gets crowded.

### Brand

Pull the primary color, background preference, and overall feel from the site. This feeds the per-app style prompts in the layout reference.

### Icon

Also grab the real icon (SVG > PNG > JPG) into `icon/`. Same rule as everywhere: fetched, never generated. Full procedure in the icon reference from the parent SKILL.

## Step 2: match copy to visuals, both directions

Every image's words and picture must describe the same thing. The failure to avoid is a headline about "real-time collaboration" over a screenshot of a static report.

- **Hero:** platform-level value proposition + the screenshot that shows the most at once.
- **Feature images:** one feature each. The copy names precisely what is visible.
- **Screenshot-led:** when you have the screenshots, write copy from what they actually show. Two images may come from one interface if you crop to different regions — focus one on the comparison pane, another on the output panel and deploy button, then write to each crop.
- **Copy-led:** when screenshots can't cover a feature, build a diagram instead — architecture, flow, node graph, data path. A generated diagram is a legitimate substitute for a screenshot; a generated *screenshot* is not.

## Step 3: compose and render

See [olares-publish-listing-layout.md](olares-publish-listing-layout.md) — wireframe choice, the two-frameworks-maximum rule, typography, and the render contract.

## Step 4: deliver in this structure

```
{app}/
├── 1.png              # hero -> featuredImage
├── 2.png              # feature images, numbered
├── 3.png
├── screenshots/       # real captures used as source
│   ├── screenshot_1.png
│   └── screenshot_2.png
└── icon/
    └── icon.svg       # or icon.png
```

Before delivering, open every file and confirm the frame is exactly 1440x900 and under 8 MB:

```bash
python3 -c "
from PIL import Image; import glob,os,sys
bad=0
for p in sorted(sum((glob.glob(e) for e in ('*.png','*.jpg','*.jpeg','*.webp')), [])):
    im=Image.open(p); mb=os.path.getsize(p)/1048576
    ok = im.size==(1440,900) and mb<=8
    print(('OK  ' if ok else 'BAD '), p, f'{im.size[0]}x{im.size[1]}', f'{mb:.1f} MB')
    bad += not ok
sys.exit(bad)"
```

Also open each file under `screenshots/` — that check is there to catch a screenshot that was described but never actually downloaded.

Ship a short note alongside the images saying where the copy came from and why the layout was chosen. If the user dislikes the result, have them pin down a color, layout, or headline and regenerate rather than re-rolling blind.

## Step 5: wire into the manifest

Upload to [imghost.olares.com](https://imghost.olares.com/) (pick **featured image** / **promotional image**) or host them yourself, then:

```yaml
spec:
  featuredImage: https://<your-host>/myapp-1.png
  promoteImage:
    - https://<your-host>/myapp-1.png
    - https://<your-host>/myapp-2.png
    - https://<your-host>/myapp-3.png
```

URLs must be reachable by the Market CDN. Re-run `olares-cli chart lint ./<app>`; for a listed app, images ship via an `UPDATE` PR with a version bump.

## Stop and ask when

Do not improvise around any of these — pause and put it to the user:

- Official site and GitHub both unreachable
- No screenshot obtainable in any form — UI, architecture diagram, or terminal
- No icon obtainable
- App too unusual to extract normal feature points from (bare CLI tools often land here)
