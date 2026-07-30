#!/usr/bin/env python3
"""Composite a fetched project logo into an Olares Market app icon.

Output is 256x256 RGBA, <=512 KB, PNG or WEBP. Runs fully offline: the
source mark must already have been fetched from the project's repo or site.
This script never invents artwork.

    generate_icon.py source_icon.svg MyApp.png
    generate_icon.py source_icon.png MyApp.png --force-bg
    generate_icon.py source_icon.png MyApp.png --force-transparent

Requires Pillow, numpy, scipy; cairosvg additionally for SVG input.
"""

import argparse
import io
import os
import sys

CANVAS = 256
SAFE_MARGIN = 38
GRADIENT_TOP = (0xFB, 0xFB, 0xFB)
GRADIENT_BOTTOM = (0xF1, 0xF1, 0xF1)
MAX_KB = 512
ALPHA_OPAQUE = 128
NEAR_WHITE = 240
CORNER_SAMPLE = 5
RESOLUTION_FLOOR = 200
SVG_RASTER_SIZE = 1024


def die(msg):
    print(f"error: {msg}", file=sys.stderr)
    sys.exit(1)


def load_source(path):
    from PIL import Image

    if not os.path.exists(path):
        die(f"source not found: {path}")

    if path.lower().endswith(".svg"):
        try:
            import cairosvg
        except ImportError:
            die("SVG input needs cairosvg (pip install cairosvg)")
        png = cairosvg.svg2png(
            url=path, output_width=SVG_RASTER_SIZE, output_height=SVG_RASTER_SIZE
        )
        return Image.open(io.BytesIO(png)).convert("RGBA")

    try:
        return Image.open(path).convert("RGBA")
    except Exception as exc:
        die(f"cannot open {path}: {exc}")


def classify_background(img):
    """Average a 5x5 block at each corner and branch on the result."""
    import numpy as np

    arr = np.asarray(img, dtype=np.float64)
    h, w = arr.shape[:2]
    n = min(CORNER_SAMPLE, h, w)
    corners = [
        arr[:n, :n],
        arr[:n, w - n :],
        arr[h - n :, :n],
        arr[h - n :, w - n :],
    ]
    mean = np.mean([c.reshape(-1, 4).mean(axis=0) for c in corners], axis=0)

    if mean[3] < ALPHA_OPAQUE:
        return "transparent"
    if all(mean[i] >= NEAR_WHITE for i in range(3)):
        return "white"
    return "colored"


def remove_white_background(img):
    """Clear near-white pixels connected to the border, keeping white inside the mark.

    Flood-fills inward from the edges rather than clearing every white pixel,
    which is what preserves white counters inside a glyph.
    """
    from collections import deque

    import numpy as np
    from PIL import Image

    arr = np.array(img)
    h, w = arr.shape[:2]
    white = (arr[:, :, :3] >= NEAR_WHITE).all(axis=2) & (arr[:, :, 3] >= ALPHA_OPAQUE)
    if not white.any():
        return img

    seen = np.zeros((h, w), dtype=bool)
    queue = deque()
    border = [(0, x) for x in range(w)] + [(h - 1, x) for x in range(w)]
    border += [(y, 0) for y in range(h)] + [(y, w - 1) for y in range(h)]
    for y, x in border:
        if white[y, x] and not seen[y, x]:
            seen[y, x] = True
            queue.append((y, x))

    while queue:
        y, x = queue.popleft()
        for ny, nx in ((y - 1, x), (y + 1, x), (y, x - 1), (y, x + 1)):
            if 0 <= ny < h and 0 <= nx < w and white[ny, nx] and not seen[ny, nx]:
                seen[ny, nx] = True
                queue.append((ny, nx))

    if not seen.any():
        return img
    arr[seen, 3] = 0
    return Image.fromarray(arr)


def trim_transparent(img):
    bbox = img.getbbox()
    return img.crop(bbox) if bbox else img


def gradient_canvas():
    from PIL import Image

    img = Image.new("RGBA", (CANVAS, CANVAS))
    px = img.load()
    for y in range(CANVAS):
        t = y / (CANVAS - 1)
        row = tuple(
            round(GRADIENT_TOP[i] + (GRADIENT_BOTTOM[i] - GRADIENT_TOP[i]) * t)
            for i in range(3)
        )
        for x in range(CANVAS):
            px[x, y] = row + (255,)
    return img


def fit_safe_zone(mark):
    """Centre the mark inside the 180x180 safe zone over the standard gradient."""
    from PIL import Image

    canvas = gradient_canvas()
    box = CANVAS - 2 * SAFE_MARGIN
    mark = trim_transparent(mark)
    scale = min(box / mark.width, box / mark.height)
    size = (max(1, round(mark.width * scale)), max(1, round(mark.height * scale)))
    mark = mark.resize(size, Image.LANCZOS)
    canvas.alpha_composite(
        mark, ((CANVAS - size[0]) // 2, (CANVAS - size[1]) // 2)
    )
    return canvas


def scale_to_fill(mark):
    """Cover the full 256x256 canvas; used when the icon carries its own background."""
    from PIL import Image

    scale = max(CANVAS / mark.width, CANVAS / mark.height)
    size = (max(1, round(mark.width * scale)), max(1, round(mark.height * scale)))
    mark = mark.resize(size, Image.LANCZOS)
    left = (size[0] - CANVAS) // 2
    top = (size[1] - CANVAS) // 2
    return mark.crop((left, top, left + CANVAS, top + CANVAS))


def save(img, path):
    fmt = "WEBP" if path.lower().endswith(".webp") else "PNG"
    if fmt == "WEBP":
        img.save(path, "WEBP", quality=90, method=6)
    else:
        img.save(path, "PNG", optimize=True)

    return os.path.getsize(path) / 1024


def verify(path, needs_alpha):
    from PIL import Image

    img = Image.open(path)
    kb = os.path.getsize(path) / 1024
    problems = []
    if img.size != (CANVAS, CANVAS):
        problems.append(f"size {img.size[0]}x{img.size[1]}, want {CANVAS}x{CANVAS}")
    # A fully opaque icon may legitimately come back as RGB from the WEBP
    # encoder, which drops an all-255 alpha channel. Only insist on alpha
    # when the composite actually has transparent pixels to lose.
    if needs_alpha and img.mode != "RGBA":
        problems.append(f"mode {img.mode}, want RGBA — transparency was lost")
    elif img.mode not in ("RGBA", "RGB"):
        problems.append(f"mode {img.mode}, want RGBA or RGB")
    if kb > MAX_KB:
        problems.append(f"{kb:.0f} KB, want <={MAX_KB} KB — re-export as .webp")
    if problems:
        die("output rejected: " + "; ".join(problems))
    print(f"OK {path}: {img.size[0]}x{img.size[1]}, {img.mode}, {kb:.0f} KB")


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("source", help="fetched logo: .svg, .png, .jpg, .ico")
    ap.add_argument("output", help="destination .png or .webp, named after the app")
    mode = ap.add_mutually_exclusive_group()
    mode.add_argument(
        "--force-bg",
        action="store_true",
        help="treat as colored: fill 256x256, no safe zone (glow icons, rounded plates)",
    )
    mode.add_argument(
        "--force-transparent",
        action="store_true",
        help="treat as transparent: gradient backdrop plus safe zone",
    )
    args = ap.parse_args()

    src = load_source(args.source)

    if max(src.size) < RESOLUTION_FLOOR:
        print(
            f"warning: source is {src.width}x{src.height}, below the {RESOLUTION_FLOOR}px "
            "floor — the result will look blurry. Find a higher-res source or upscale "
            "with Real-ESRGAN first.",
            file=sys.stderr,
        )

    if args.force_bg:
        kind = "colored"
    elif args.force_transparent:
        kind = "transparent"
    else:
        kind = classify_background(src)

    if kind == "colored":
        out = scale_to_fill(src)
    else:
        if kind == "white":
            src = remove_white_background(src)
        out = fit_safe_zone(src)

    print(f"background: {kind}")
    needs_alpha = out.getchannel("A").getextrema()[0] < 255
    save(out, args.output)
    verify(args.output, needs_alpha)


if __name__ == "__main__":
    main()
