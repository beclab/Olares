#!/usr/bin/env python3

"""Write the release being built into every skill's frontmatter.

The suite is compiled into the olares-cli binary, so the release it ships in
is a fact about the build, not something to maintain by hand in twelve files.
It used to be maintained by hand, and the failure had no floor: an OS-line
release whose skills still named the previous one shipped new instructions
under a label that matched what the user already had, so nothing anywhere
said the copy on disk was out of date.

The release job runs this before compiling. What is committed is a
placeholder, which is why nothing here reads the current value: there is
nothing in the working tree worth preserving.

Standard library only, deliberately -- this runs in a job that has Go and
nothing else installed, and adding a pip step to stamp a version is a step
that can fail on a release that is otherwise ready.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent
FENCE = "---"
VERSION_KEY = "version:"
# The shape release-cli.yaml enforces for an npm version, which is the only
# number the suite is allowed to carry. Kept in step with validate.py's
# RELEASE_VERSION_RE and with embed_test.go's releaseVersion.
RELEASE_VERSION_RE = re.compile(r"^\d+\.\d+\.\d+-cli\.\d+$")
# What is committed. It satisfies the shape above, so validate.py needs no
# carve-out for it, and it names no release anybody could mistake for one.
# publish.sh refuses it; embed_test.go pins it.
PLACEHOLDER_VERSION = "0.0.0-cli.0"


class StampError(Exception):
    pass


def stamp_file(skill: Path, version: str) -> None:
    """Rewrite the version in one SKILL.md's frontmatter.

    Only inside the frontmatter: a body that documents a `version:` field of
    somebody else's YAML is prose, and rewriting it would corrupt an example
    an agent is meant to copy.
    """
    lines = skill.read_text(encoding="utf-8").split("\n")
    if not lines or lines[0].strip() != FENCE:
        raise StampError(f"{skill}: does not open with a {FENCE!r} frontmatter fence")
    for index, line in enumerate(lines[1:], start=1):
        if line.strip() == FENCE:
            raise StampError(f"{skill}: no {VERSION_KEY} line in the frontmatter")
        if line.startswith(VERSION_KEY):
            lines[index] = f"{VERSION_KEY} {version}"
            skill.write_text("\n".join(lines), encoding="utf-8")
            return
    raise StampError(f"{skill}: frontmatter fence is never closed")


def stamp(root: Path, version: str) -> list[Path]:
    """Stamp every skill under root and return the files written."""
    if not RELEASE_VERSION_RE.match(version):
        raise StampError(
            f"{version!r} is not an olares-cli release (x.y.z-cli.n); "
            "the suite carries the release it ships in and nothing else"
        )
    skills = sorted(root.glob("olares-*/SKILL.md"))
    if not skills:
        raise StampError(f"no skills under {root}")
    for skill in skills:
        stamp_file(skill, version)
    return skills


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {Path(argv[0]).name} <x.y.z-cli.n>", file=sys.stderr)
        return 2
    try:
        stamped = stamp(ROOT, argv[1])
    except StampError as error:
        print(f"stamp failed: {error}", file=sys.stderr)
        return 1
    print(f"stamped {len(stamped)} skills at {argv[1]}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
