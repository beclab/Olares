#!/usr/bin/env python3
"""Render a 1440x900 Olares Market listing image around a real screenshot.

Three stages, per the render contract in
references/olares-publish-listing-layout.md:

  1. pre-composite  Pillow places the real screenshot into the frame's
                    screenshot region, producing the layout reference.
  2. render         an image model generates backdrop, typography and
                    decoration around that region (skipped without a key).
  3. re-composite   Pillow pastes the real screenshot back at the same box.

Stage 3 is why the shipped UI pixels are authentic: the model regenerates the
whole canvas, so its version of the screenshot is a reconstruction and must
not be delivered. This script owns the bounding box so stages 1 and 3 cannot
disagree.

    render_promo.py --screenshot shot.png --out 1.png --frame A \
        --headline "Self-hosted search that respects you" --highlight search

Requires Pillow. Uses OPENAI_API_KEY when set; without it, emits the
deterministic tier and says so.
"""

import argparse
import base64
import io
import json
import os
import sys
import urllib.error
import urllib.request
import uuid

W, H = 1440, 900
RENDER_W, RENDER_H = 1536, 960
MAX_MB = 8
CORNER_RADIUS = 18
SHADOW_BLUR = 26
SHADOW_OFFSET = 18
REFERENCE_MAX_BYTES = 1_500_000
API_URL = "https://api.openai.com/v1/images/edits"
DEFAULT_MODEL = "gpt-image-2"

# Boxes are (x, y, w, h) on the 1440x900 frame. A box may extend past the
# bottom edge, which is the intended bleed for frame A.
FRAMES = {
    "A": {"shot": (160, 330, 1120, 640), "text": (120, 92, 1200, 200), "align": "center"},
    "B": {"shot": (250, 320, 940, 560), "text": (120, 80, 1200, 190), "align": "center"},
    "C": {"shot": (700, 210, 640, 480), "text": (100, 280, 520, 340), "align": "left"},
    "D": {"shot": (740, 120, 580, 660), "text": (100, 250, 540, 400), "align": "left"},
    "E": {"shot": (170, 350, 1100, 550), "text": (110, 90, 820, 200), "align": "left"},
    "F": {"shot": (95, 320, 610, 500), "text": (120, 80, 1200, 180), "align": "center"},
    "G": {"shot": (0, 0, 1440, 900), "text": (120, 620, 1200, 200), "align": "center"},
}

FONT_CANDIDATES = [
    "/System/Library/Fonts/Supplemental/Helvetica.ttc",
    "/System/Library/Fonts/HelveticaNeue.ttc",
    "/System/Library/Fonts/SFNS.ttf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
    "/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
]


def die(msg):
    print(f"error: {msg}", file=sys.stderr)
    sys.exit(1)


def note(msg):
    print(f"note: {msg}", file=sys.stderr)


def load_font(path, size):
    from PIL import ImageFont

    candidates = [path] if path else []
    candidates += FONT_CANDIDATES
    for p in candidates:
        if p and os.path.exists(p):
            try:
                return ImageFont.truetype(p, size)
            except OSError:
                continue
    return ImageFont.load_default()


def dominant_accent(img):
    """Pick the most saturated frequent color, used when --accent is omitted."""
    import colorsys

    small = img.convert("RGB").resize((64, 64))
    counts = {}
    for px in small.getdata():
        r, g, b = (c / 255 for c in px)
        h_, l_, s_ = colorsys.rgb_to_hls(r, g, b)
        if s_ < 0.25 or l_ < 0.15 or l_ > 0.9:
            continue
        key = tuple(c // 24 for c in px)
        counts.setdefault(key, [0, px])
        counts[key][0] += 1
    if not counts:
        return (64, 110, 220)
    return max(counts.values(), key=lambda v: v[0])[1]


def parse_color(text, fallback):
    if not text:
        return fallback
    t = text.strip().lstrip("#")
    if len(t) != 6:
        die(f"--accent must be #RRGGBB, got {text}")
    return tuple(int(t[i : i + 2], 16) for i in (0, 2, 4))


def mix(a, b, t):
    return tuple(round(a[i] + (b[i] - a[i]) * t) for i in range(3))


def backdrop(accent):
    """Dark brand-tinted gradient with a faint dot grid: never flat color."""
    from PIL import Image, ImageDraw

    top = mix(accent, (10, 12, 20), 0.72)
    bottom = mix(accent, (4, 5, 12), 0.88)
    img = Image.new("RGB", (W, H))
    draw = ImageDraw.Draw(img)
    for y in range(H):
        draw.line([(0, y), (W, y)], fill=mix(top, bottom, y / (H - 1)))

    dots = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(dots)
    tint = mix(accent, (255, 255, 255), 0.5)
    for y in range(0, H, 40):
        for x in range(0, W, 40):
            d.ellipse([x, y, x + 2, y + 2], fill=tint + (26,))
    img = Image.alpha_composite(img.convert("RGBA"), dots)

    glow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    ImageDraw.Draw(glow).ellipse(
        [W * 0.55, -H * 0.35, W * 1.25, H * 0.75], fill=accent + (46,)
    )
    from PIL import ImageFilter

    glow = glow.filter(ImageFilter.GaussianBlur(160))
    return Image.alpha_composite(img, glow)


def rounded(img, radius):
    from PIL import Image, ImageDraw

    mask = Image.new("L", img.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle([0, 0, img.width - 1, img.height - 1], radius, fill=255)
    out = img.convert("RGBA")
    out.putalpha(mask)
    return out


def cover(img, box_w, box_h):
    from PIL import Image

    scale = max(box_w / img.width, box_h / img.height)
    size = (max(1, round(img.width * scale)), max(1, round(img.height * scale)))
    img = img.resize(size, Image.LANCZOS)
    left = (size[0] - box_w) // 2
    top = (size[1] - box_h) // 2
    return img.crop((left, top, left + box_w, top + box_h))


def place_screenshot(canvas, shot, box, radius=CORNER_RADIUS, shadow=True):
    """Paste the real capture at box with rounded corners and a soft shadow."""
    from PIL import Image, ImageDraw, ImageFilter

    x, y, bw, bh = box
    full_bleed = (x, y, bw, bh) == (0, 0, W, H)
    panel = cover(shot, bw, bh)
    panel = rounded(panel, 0 if full_bleed else radius)

    if shadow and not full_bleed:
        layer = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
        ImageDraw.Draw(layer).rounded_rectangle(
            [x, y + SHADOW_OFFSET, x + bw, y + bh + SHADOW_OFFSET], radius, fill=(0, 0, 0, 150)
        )
        canvas.alpha_composite(layer.filter(ImageFilter.GaussianBlur(SHADOW_BLUR)))

    canvas.alpha_composite(panel, (x, y))
    return canvas


def text_scrim(canvas, textbox):
    """Darken behind overlaid text; without it, copy on a busy capture is unreadable."""
    from PIL import Image, ImageDraw, ImageFilter

    x, y, bw, bh = textbox
    pad = 90
    layer = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    ImageDraw.Draw(layer).rectangle(
        [x - pad, y - pad, x + bw + pad, y + bh + pad], fill=(6, 8, 14, 205)
    )
    canvas.alpha_composite(layer.filter(ImageFilter.GaussianBlur(60)))
    return canvas


def draw_headline(canvas, frame, headline, highlight, accent, font_path, scrim=False):
    """Wrap the headline into the text box, tinting highlighted words."""
    from PIL import ImageDraw

    if not headline:
        return canvas

    x, y, bw, bh = frame["text"]
    if scrim:
        canvas = text_scrim(canvas, frame["text"])
    align = frame["align"]
    draw = ImageDraw.Draw(canvas)
    words = headline.split()
    marks = {w.lower().strip(".,:;!?") for w in (highlight or "").split(",") if w.strip()}

    size = 76
    while size > 28:
        font = load_font(font_path, size)
        space = draw.textlength(" ", font=font)
        lines, cur, cur_w = [], [], 0.0
        for word in words:
            ww = draw.textlength(word, font=font)
            if cur and cur_w + space + ww > bw:
                lines.append((cur, cur_w))
                cur, cur_w = [word], ww
            else:
                cur_w = cur_w + space + ww if cur else ww
                cur.append(word)
        if cur:
            lines.append((cur, cur_w))
        line_h = size * 1.22
        if len(lines) <= 3 and len(lines) * line_h <= bh:
            break
        size -= 4

    for i, (line, line_w) in enumerate(lines):
        ly = y + i * line_h
        lx = x + (bw - line_w) / 2 if align == "center" else x
        for word in line:
            color = accent if word.lower().strip(".,:;!?") in marks else (255, 255, 255)
            draw.text((lx, ly), word, font=font, fill=color + (255,))
            lx += draw.textlength(word, font=font) + draw.textlength(" ", font=font)
    return canvas


def boxes_overlap(a, b):
    ax, ay, aw, ah = a
    bx, by, bw, bh = b
    return ax < bx + bw and bx < ax + aw and ay < by + bh and by < ay + ah


def to_png_bytes(img, max_bytes=REFERENCE_MAX_BYTES):
    from PIL import Image

    buf = io.BytesIO()
    img.convert("RGB").save(buf, "PNG", optimize=True)
    if buf.tell() <= max_bytes:
        return buf.getvalue()
    scaled = img.convert("RGB")
    while buf.tell() > max_bytes and scaled.width > 400:
        scaled = scaled.resize((int(scaled.width * 0.8), int(scaled.height * 0.8)), Image.LANCZOS)
        buf = io.BytesIO()
        scaled.save(buf, "PNG", optimize=True)
    return buf.getvalue()


def build_prompt(args, frame_letter):
    typeface = args.typeface
    parts = [
        f"Design a polished {W}x{H} app-store listing image, 8:5, wireframe {frame_letter}.",
        "Use image 1 ONLY for spatial layout and composition — ignore its colors entirely.",
        "Embed the UI from image 2 faithfully in the same region; DO NOT redraw or restyle it.",
    ]
    if args.reference:
        parts.append(
            "Image 3 is the first image of this set: match its typeface, palette and overall tone exactly."
        )
    if args.headline:
        parts.append(f'Headline text, rendered exactly: "{args.headline}".')
    if args.highlight:
        parts.append(f"Render these words in the brand accent color: {args.highlight}.")
    parts += [
        f"Typeface: {typeface}. Maintain strict font aspect ratio. "
        "DO NOT compress or stretch text horizontally or vertically to fit space.",
        f"Brand accent color {args.accent_hex}. Backdrop must carry texture — "
        "fine grid, dot matrix or gradient bloom — never flat black or flat white.",
        "Add one or two decorative elements (geometric shapes, glow, floating cards) "
        "in the empty area of the frame, in this same generation. "
        "Decoration must contain NO readable text of any kind.",
    ]
    if args.style:
        parts.append(args.style)
    return " ".join(parts)


def call_model(api_key, model, prompt, images):
    boundary = uuid.uuid4().hex
    body = bytearray()

    def field(name, value):
        body.extend(f"--{boundary}\r\n".encode())
        body.extend(f'Content-Disposition: form-data; name="{name}"\r\n\r\n'.encode())
        body.extend(f"{value}\r\n".encode())

    field("model", model)
    field("prompt", prompt)
    field("size", f"{RENDER_W}x{RENDER_H}")
    field("n", "1")

    for i, payload in enumerate(images):
        body.extend(f"--{boundary}\r\n".encode())
        body.extend(
            f'Content-Disposition: form-data; name="image[]"; filename="ref{i}.png"\r\n'.encode()
        )
        body.extend(b"Content-Type: image/png\r\n\r\n")
        body.extend(payload)
        body.extend(b"\r\n")
    body.extend(f"--{boundary}--\r\n".encode())

    req = urllib.request.Request(
        API_URL,
        data=bytes(body),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": f"multipart/form-data; boundary={boundary}",
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=300) as resp:
            payload = json.loads(resp.read())
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode(errors="replace")[:600]
        if exc.code == 400 and "model" in detail:
            die(
                "the API rejected the model. GPT Image models require organization "
                f"verification on the OpenAI account — this is a configuration problem, "
                f"not a transient one. Response: {detail}"
            )
        die(f"images/edits returned HTTP {exc.code}: {detail}")
    except urllib.error.URLError as exc:
        die(f"cannot reach the API: {exc.reason}")

    try:
        return base64.b64decode(payload["data"][0]["b64_json"])
    except (KeyError, IndexError) as exc:
        die(f"unexpected response shape: {exc}")


def main():
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--screenshot", required=True, help="real capture, never a mockup")
    ap.add_argument("--out", required=True, help="destination .png / .jpg / .webp")
    ap.add_argument("--frame", default="A", choices=sorted(FRAMES), help="wireframe, default A")
    ap.add_argument("--headline", default="", help="1-3 lines, lifted from the hero slogan")
    ap.add_argument("--highlight", default="", help="comma-separated words to tint with the accent")
    ap.add_argument("--accent", default="", help="#RRGGBB; sampled from the screenshot if omitted")
    ap.add_argument("--style", default="", help="extra style prompt: material, lighting, spatial feel")
    ap.add_argument("--reference", default="", help="first finished image of the set, for consistency")
    ap.add_argument("--font", default="", help="path to the sans used across the whole set")
    ap.add_argument("--typeface", default="Inter", help="typeface name passed to the model")
    ap.add_argument("--model", default=DEFAULT_MODEL)
    ap.add_argument(
        "--deterministic",
        action="store_true",
        help="skip the model even if a key is present",
    )
    args = ap.parse_args()

    from PIL import Image

    if not os.path.exists(args.screenshot):
        die(f"screenshot not found: {args.screenshot}")
    shot = Image.open(args.screenshot).convert("RGBA")

    frame = FRAMES[args.frame]
    box = frame["shot"]
    accent = parse_color(args.accent, dominant_accent(shot))
    args.accent_hex = "#%02X%02X%02X" % accent

    # Stage 1: pre-composite. The box recorded here is the one stage 3 reuses.
    overlaid = boxes_overlap(box, frame["text"])
    reference = place_screenshot(backdrop(accent), shot, box)
    reference = draw_headline(
        reference, frame, args.headline, args.highlight, accent, args.font, scrim=overlaid
    )

    api_key = os.environ.get("OPENAI_API_KEY", "").strip()
    rendered = None

    if args.deterministic:
        note("--deterministic set: skipping the model.")
    elif not api_key:
        note(
            "OPENAI_API_KEY is not set, so the rendered tier is unavailable. "
            "Delivering the deterministic composition — real screenshot, brand gradient, "
            "typeset headline. Say so when handing this over; do not present it as the "
            "fully rendered result."
        )
    else:
        images = [to_png_bytes(reference), to_png_bytes(shot)]
        if args.reference:
            if not os.path.exists(args.reference):
                die(f"--reference not found: {args.reference}")
            images.append(to_png_bytes(Image.open(args.reference)))
        raw = call_model(api_key, args.model, build_prompt(args, args.frame), images)
        rendered = Image.open(io.BytesIO(raw)).convert("RGBA")

    if rendered is None:
        final = reference
    else:
        # Stage 3: downsample, then put the authentic pixels back.
        final = rendered.resize((W, H), Image.LANCZOS)
        raw_path = os.path.splitext(args.out)[0] + ".render.png"
        rendered.save(raw_path)
        note(f"raw model output kept at {raw_path} — compare it against {args.out}.")
        final = place_screenshot(final, shot, box)
        if overlaid:
            final = draw_headline(
                final, frame, args.headline, args.highlight, accent, args.font, scrim=True
            )

    fmt = os.path.splitext(args.out)[1].lower()
    if fmt in (".jpg", ".jpeg"):
        final.convert("RGB").save(args.out, quality=92, optimize=True)
    elif fmt == ".webp":
        final.save(args.out, "WEBP", quality=92, method=6)
    else:
        final.convert("RGB").save(args.out, "PNG", optimize=True)

    check = Image.open(args.out)
    mb = os.path.getsize(args.out) / 1048576
    if check.size != (W, H):
        die(f"output is {check.size[0]}x{check.size[1]}, want {W}x{H}")
    if mb > MAX_MB:
        die(f"output is {mb:.1f} MB, want <={MAX_MB} MB")
    print(f"OK {args.out}: {check.size[0]}x{check.size[1]}, {mb:.1f} MB, frame {args.frame}, accent {args.accent_hex}")
    if rendered is not None:
        print(
            "Check that the model's screenshot area matches the box; if it drifted, "
            "re-render rather than shipping a paste over a mismatch."
        )


if __name__ == "__main__":
    main()
