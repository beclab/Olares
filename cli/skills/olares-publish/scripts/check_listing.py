#!/usr/bin/env python3
"""Check Manifest text and listing image URLs before Market submission.

Requires Python 3 and PyYAML. Does not modify the chart.
"""

import argparse
from pathlib import Path
import re
import sys
from urllib.parse import urlsplit
from urllib.request import Request, urlopen

import yaml


IMAGE_FIELDS = {"icon", "featuredImage", "promoteImage"}
# Strong signals only: ordinary accented Latin, Chinese and Japanese are valid.
SUSPECT = re.compile("\ufffd|\u00ef\u00bf\u00bd|\u00e2\u20ac[\u201c\u201d\u2122]|[\x00-\x08\x0b\x0c\x0e-\x1f]")


def image_urls(value, path="", in_image=False):
    """Collect image fields anywhere in a Manifest, including entrances."""
    if isinstance(value, dict):
        for key, child in value.items():
            yield from image_urls(child, f"{path}.{key}".lstrip("."), in_image or key in IMAGE_FIELDS)
    elif isinstance(value, list):
        for index, child in enumerate(value):
            yield from image_urls(child, f"{path}[{index}]", in_image)
    elif in_image and isinstance(value, str) and value:
        yield path, value


def check_url(url, timeout):
    parsed = urlsplit(url)
    if parsed.scheme not in ("http", "https") or not parsed.netloc:
        raise ValueError("expected an absolute http(s) image URL")
    # GET avoids false failures on CDNs that reject HEAD; read only a small prefix.
    request = Request(url, headers={"User-Agent": "Olares-listing-check/1.0"})
    with urlopen(request, timeout=timeout) as response:
        if not 200 <= response.status < 300:
            raise ValueError(f"HTTP {response.status}")
        kind = response.headers.get_content_type()
        if not kind.startswith("image/"):
            raise ValueError(f"expected image content, received {kind}")
        if not response.read(512):
            raise ValueError("empty image response")


def check_chart(chart, timeout=15):
    root = chart / "OlaresManifest.yaml"
    files = [root, *sorted((chart / "i18n").rglob("OlaresManifest.yaml"))]
    errors = []
    urls = {}
    for file in files:
        try:
            text = file.read_text(encoding="utf-8", errors="strict")
            suspect_lines = set()
            for number, line in enumerate(text.splitlines(), 1):
                if SUSPECT.search(line):
                    suspect_lines.add(number)
                    errors.append(f"{file}:{number}: suspected garbled text; review against the source")
            data = yaml.safe_load(text)
            if not isinstance(data, dict):
                raise ValueError("expected a Manifest mapping")
            # Also catches replacement characters represented by YAML escapes.
            for token in yaml.scan(text):
                if (isinstance(token, yaml.tokens.ScalarToken) and SUSPECT.search(token.value)
                        and not any(token.start_mark.line + 1 <= line <= token.end_mark.line + 1
                                    for line in suspect_lines)):
                    issue = f"{file}:{token.start_mark.line + 1}: suspected garbled YAML scalar"
                    errors.append(issue)
            for field, url in image_urls(data):
                urls.setdefault(url, []).append(f"{file}:{field}")
        except (OSError, UnicodeError, yaml.YAMLError, ValueError) as exc:
            errors.append(f"{file}: {exc}")
    for url, locations in urls.items():
        try:
            check_url(url, timeout)
            print(f"OK image: {url}")
        except Exception as exc:
            errors.append(f"{', '.join(locations)}: {url}: {exc}")
    for error in errors:
        print(error, file=sys.stderr)
    print(f"Checked {len(files)} Manifests and {len(urls)} unique image URLs; {len(errors)} issues")
    return 1 if errors else 0


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("chart", type=Path)
    parser.add_argument("--timeout", type=float, default=15, help="HTTP timeout in seconds (default: 15)")
    args = parser.parse_args()
    if args.timeout <= 0:
        parser.error("--timeout must be positive")
    return check_chart(args.chart, args.timeout)


if __name__ == "__main__":
    sys.exit(main())
