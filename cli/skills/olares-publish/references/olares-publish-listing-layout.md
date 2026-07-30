# Listing image composition and render contract

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first.
> Material collection, copy, delivery structure, and manifest wiring are in [olares-publish-listing-images.md](olares-publish-listing-images.md). This file covers how the 1440x900 frame is laid out and rendered.

## How far you can get depends on your tools

Split the work at the point where a hosted image model becomes necessary. Only the last stage needs one.

| Tier | Work | Needs |
|---|---|---|
| **Deterministic** | Collect, upscale, pre-composite, resize to exactly 1440x900, verify | Pillow only. No key. |
| **Rendered** | Backdrop, typography, decoration generated around the screenshot region, with the real capture composited back in | An image model |

The deterministic tier already delivers something shippable: real screenshot with rounded corners and a soft shadow, brand-colored gradient backdrop, headline set in a local sans with the key word highlighted. That satisfies most of the layout rules below. The rendered tier buys polish — organic lighting, 3D accents, designed decoration.

Reach the rendered tier by, in order: the host agent's own image tool, then [`../scripts/render_promo.py`](../scripts/render_promo.py) if `OPENAI_API_KEY` is set. If neither is available, **say so and hand over the deterministic output** — don't quietly ship a lesser result as if it were the intended one.

## Wireframes

Seven frames distilled from real app-store listings:

| Frame | Composition | Use when |
|---|---|---|
| **A** | Centered headline, full-width screenshot along the bottom | Default. Screenshot is content-rich. |
| **B** | Headline on top, main screenshot, floating feature cards | Showing layered functionality |
| **C** | Text left, angled screenshot right, badge | Squarish screenshot, or you want depth |
| **D** | Text left, upright screenshot right, vertically centered | Longer copy |
| **E** | Headline on top, mascot/brand element upper right, screenshot below | Brand has a character |
| **F** | Centered headline, two screenshots side by side | Comparisons — dark/light, two views |
| **G** | Full-bleed screenshot, text overlaid on top | UI is visually striking on its own |

**At most two frames per app.** One frame means hero and all feature images share it. Two means the hero takes one and *every* feature image takes the other. More than that and the set stops reading as a set.

## Layout rules

Frame is 1440x900 (8:5).

**Top-text / bottom-image is the default** — highest tolerance for varied content. Text sits in the upper third to half, headline very large, 1–3 lines, centered or edge-aligned. The screenshot occupies the lower area and may bleed off the bottom edge for a sense of continuation. Works best with wide screenshots such as dashboards.

**Side-by-side** suits phone screenshots and tall panels: text on one side at roughly a third of the width, screenshot on the other, either one tall capture or several overlapping.

Alignment and whether to show the icon are both judgment calls — pick what looks right. But **however you present screenshots, present them the same way across one app's set.**

### Background

Never flat black or flat white. Take the base tone from the site's primary color and add texture — fine grid, dot matrix, faint gradient bloom, or geometric shapes. Light or dark follows the brand and legibility. Backgrounds may vary across a set (per-feature tinting is fine) as long as the family resemblance holds.

### Type

- Dark background, white headline. Light background, near-black — not pure black.
- **Highlight one or two key words in the brand accent color.** Cheap to do, and it is what separates a designed frame from a plain one.
- One sans family (Inter, SF Pro, Roboto) across every image of the app, no exceptions.
- Subhead optional: smaller, lower contrast — grey, or white at partial opacity.

### Screenshots in frame

Rounded corners and a soft drop shadow, always, so they separate from the backdrop. A macOS window chrome or a minimal device shell adds realism. Full-screen captures are not required — crop to the region that matters, or scatter several panels as overlapping cards. Blurring a full screenshot into a backdrop and floating sharp detail cards on top is a good trick when one capture has to carry two jobs.

### Decoration

Allowed: icons, geometric patterns, waves, scattered dots, 3D shapes, floating cards, grids, glows, gradient overlays.

**Never any text** — no labels, no badges with words, no fake code. Decoration carries zero readable content.

One or two elements, placed where the frame is empty. They must not crowd the headline or cover the screenshot.

## Render contract

Three stages. **The shipped screenshot pixels are placed by Pillow, not by the model.**

This is not a stylistic preference. `POST /v1/images/edits` does not composite — it regenerates the entire canvas, so whatever the model returns in the screenshot area is its *reconstruction* of the UI, no matter how firmly the prompt says otherwise. Shipping that would violate the hard rule in the parent SKILL against fabricated UI. So the model builds the frame, and the real capture goes in afterwards.

**Stage 1 — pre-composite.** With Pillow, scale and crop the real screenshot into the chosen wireframe's screenshot region. Two outputs: a pre-composited reference that fixes the spatial layout, and **the region's bounding box**, which stage 3 needs. Keep the box.

**Stage 2 — render once.** Pass the references in this exact order — order is what the prompt refers to as "image 1 / image 2 / image 3":

1. **Pre-composited reference** — spatial layout only. State explicitly that its greys are to be ignored.
2. **The full-resolution screenshot** — the UI the frame is being built around. The model's rendering of it is scaffolding that stage 3 overwrites; asking for fidelity here still helps, because it keeps the surrounding lighting and perspective consistent with what finally lands there.
3. **The first finished image** — from image 2 of the set onward, as the lock for typeface and overall tone.

**Stage 3 — re-composite.** Downsample the output to exactly 1440x900, then paste the real screenshot back at the stage-1 bounding box, applying the rounded corners and drop shadow deterministically. Now every UI pixel in the deliverable is authentic.

Both stages that touch the box must use the same numbers, which is why [`../scripts/render_promo.py`](../scripts/render_promo.py) owns it rather than leaving it to be re-derived.

> If the model's rendered screenshot area drifts noticeably from the box — different angle, different proportions — **re-render instead of pasting over it.** A paste onto a mismatched area leaves a visible double edge, which looks worse than the reconstruction it was meant to fix.

The prompt must carry:

- The app's own style prompts — material, lighting, spatial feel — derived from its brand, applied identically across the whole set.
- The typeface family, plus: *"Maintain strict font aspect ratio. DO NOT compress or stretch text horizontally or vertically to fit space."*
- The decoration request, **inside this same single generation**. Do not run a second image-to-image pass to add decoration — that is what produces the fibrous, linty background texture.
- Explicit reference by index, e.g. *"use image 1 ONLY for spatial layout and composition, ignore its colors; embed the UI from image 2 faithfully, DO NOT redraw it."*

### Running it

```bash
# hero image; accent is sampled from the screenshot when omitted
python3 ../scripts/render_promo.py --screenshot screenshots/dashboard.png --out 1.png \
  --frame A --headline "Self-hosted search that respects your data" --highlight search

# every later image of the set locks onto the first one
python3 ../scripts/render_promo.py --screenshot screenshots/detail.png --out 2.png \
  --frame C --headline "Index everything you own" --highlight everything \
  --accent "#3B7CE6" --reference 1.png --style "soft studio lighting, matte surfaces"

# force the deterministic tier even where a key exists
python3 ../scripts/render_promo.py --screenshot shot.png --out 1.png --deterministic
```

Needs `Pillow`; the rendered tier additionally needs `OPENAI_API_KEY`. Pass `--font` to pin the exact typeface file across a set. When the model runs, the untouched output is also written to `<out>.render.png` so you can compare it against the delivered file and catch drift.

### Model constraints worth knowing

These bite when calling `POST /v1/images/edits` directly (the script handles them):

- **Omit `input_fidelity`.** `gpt-image-2` is high-fidelity by default and rejects the parameter. Only the older `gpt-image-1` family wants an explicit `high`.
- **You cannot request 1440x900.** Both edges must be multiples of 16, and 900 is not. Ask for **1536x960** — exactly 8:5, both edges divisible by 16, within the pixel bounds — then downsample to 1440x900. Downsampling only, no stretch.
- Reference order maps to "image 1/2/3" in the prompt. Up to 16 references; keep each around 1.5 MB.
- Responses come back base64 (`b64_json`).
- GPT Image models require organization verification on the OpenAI account. A 400 on model validation is a configuration problem — report it, don't retry into a timeout.

## What good sets do

From reviewing 80+ images across 18 published listings:

Copy is short and lifted straight from the hero slogan or README one-liner — 1–3 lines at large size, with about half adding a brief subhead. Bullet lists are rare.

The image area is not required to be a UI screenshot. Infrastructure and backend tools are better served by an architecture diagram; developer tools by a terminal or code window with syntax highlighting.

Background color tracks the brand primary closely, and nearly every set uses texture or geometric decoration to break up flat color. Two-color headline highlighting shows up constantly and is the single highest-leverage detail.

Layout can shift within one set — a platform overview in side-by-side, a specific feature in top-text/bottom-image — as long as it stays within the two-frame budget.
