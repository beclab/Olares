#!/usr/bin/env python3
"""Phase-3 cleanup for remaining terminus-apps lint failures."""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import List, Optional, Tuple

REPO_ROOT = Path(__file__).resolve().parent.parent
TESTDATA = REPO_ROOT / "testdata" / "terminus-apps"

# Apps that need template OLARES_USER -> olaresEnv renames after envName fixes.
TEMPLATE_ENV_FIXES = {
    "fireflyiii": [
        (r"\.Values\.olaresEnv\.OLARES_USER_TIMEZONE", ".Values.olaresEnv.APP_TIMEZONE"),
    ],
    "freshrss": [
        (r"\.Values\.olaresEnv\.OLARES_USER_TIMEZONE", ".Values.olaresEnv.APP_TIMEZONE"),
    ],
    "ntfy": [
        (r"\.Values\.olaresEnv\.OLARES_USER_TIMEZONE", ".Values.olaresEnv.APP_TIMEZONE"),
    ],
    "openwebui": [
        (r"\.Values\.olaresEnv\.OLARES_USER_HUGGINGFACE_SERVICE", ".Values.olaresEnv.HF_ENDPOINT"),
        (r"\.Values\.olaresEnv\.OLARES_USER_HUGGINGFACE_TOKEN", ".Values.olaresEnv.HF_TOKEN"),
    ],
}


def manifest_version(text: str) -> str:
    m = re.search(r"^olaresManifest\.version:\s*['\"]?([^'\"\n]+)", text, flags=re.MULTILINE)
    return m.group(1).strip() if m else ""


def is_legacy_version(version: str) -> bool:
    m = re.match(r"^(\d+)\.(\d+)\.(\d+)", version)
    if not m:
        return False
    return tuple(map(int, m.groups())) < (0, 12, 0)


def remove_external_data(text: str) -> Tuple[str, bool]:
    new = re.sub(r"^(\s*)externalData:\s*true\s*\n", "", text, flags=re.MULTILINE)
    return new, new != text


def remove_accelerator_block(text: str) -> Tuple[str, bool]:
    m = re.search(r"^(\s*)accelerator:\s*$", text, flags=re.MULTILINE)
    if not m:
        return text, False
    indent = len(m.group(1).expandtabs())
    lines = text.splitlines(keepends=True)
    line_no = text[: m.start()].count("\n")
    end_line = line_no + 1
    while end_line < len(lines):
        line = lines[end_line]
        if not line.strip():
            end_line += 1
            continue
        stripped = line.lstrip()
        cur_indent = len(line) - len(stripped)
        if cur_indent > indent:
            end_line += 1
            continue
        if cur_indent == indent and (stripped.startswith("- ") or stripped.startswith("#")):
            end_line += 1
            continue
        break
    new = "".join(lines[:line_no] + lines[end_line:])
    return new, True


def update_olares_dep(text: str, target: str = ">=1.12.6-0") -> Tuple[str, bool]:
    pat = re.compile(
        r"(-\s+name\s*:\s*['\"]?olares['\"]?[^\n]*\n"
        r"(?:[ \t]+(?:type|registry|mandatory|description)\s*:[^\n]*\n)*"
        r"[ \t]+version\s*:\s*)['\"]?[^\n'\"]+['\"]?",
        re.MULTILINE,
    )
    new, n = pat.subn(rf"\1'{target}'", text, count=1)
    return new, n > 0


def add_spec_only_admin(text: str) -> Tuple[str, bool]:
    if re.search(r"^\s+onlyAdmin:\s*true\s*$", text, flags=re.MULTILINE):
        return text, False
    m = re.search(r"^spec:\s*$", text, flags=re.MULTILINE)
    if not m:
        return text, False
    insert_at = m.end()
    return text[:insert_at] + "\n  onlyAdmin: true" + text[insert_at:], True


def fix_templates(app_dir: Path, app_name: str) -> bool:
    fixes = TEMPLATE_ENV_FIXES.get(app_name)
    if not fixes:
        return False
    changed = False
    for p in app_dir.rglob("*"):
        if not p.is_file() or p.suffix not in {".yaml", ".yml", ".tpl"}:
            continue
        text = p.read_text()
        new = text
        for old, new_val in fixes:
            new = re.sub(old, new_val, new)
        if new != text:
            p.write_text(new)
            changed = True
    return changed


def fix_nofx_deployment(app_dir: Path) -> bool:
    dep = app_dir / "templates" / "nofx" / "deployment.yaml"
    if not dep.exists():
        return False
    text = dep.read_text()
    new = re.sub(
        r"^(metadata:\s*\n\s*name:\s*)nofx\s*$",
        r'\1"{{ .Release.Name }}"',
        text,
        count=1,
        flags=re.MULTILINE,
    )
    if new != text:
        dep.write_text(new)
        return True
    return False


def repair_app(app_dir: Path) -> List[str]:
    manifest = app_dir / "OlaresManifest.yaml"
    if not manifest.exists():
        return []
    text = manifest.read_text()
    applied: List[str] = []
    version = manifest_version(text)

    if is_legacy_version(version) and re.search(r"^\s*externalData:\s*true", text, flags=re.MULTILINE):
        text, ok = remove_external_data(text)
        if ok:
            applied.append("X-")

    if is_legacy_version(version) and re.search(r"^\s*accelerator:", text, flags=re.MULTILINE):
        has_legacy = any(
            re.search(rf"^\s+{re.escape(k)}:\s+\S", text, flags=re.MULTILINE)
            for k in ("requiredCpu", "limitedCpu", "requiredMemory", "limitedMemory", "requiredDisk")
        )
        if has_legacy:
            text, ok = remove_accelerator_block(text)
            if ok:
                applied.append("A-")

    if app_dir.name in {"testappempty", "testappv1"}:
        text, ok = update_olares_dep(text)
        if ok:
            applied.append("D")

    if app_dir.name == "teatappintelv3":
        text, ok = add_spec_only_admin(text)
        if ok:
            applied.append("O")

    if applied and any(t in applied for t in ("X-", "A-", "D", "O")):
        manifest.write_text(text)

    if fix_templates(app_dir, app_dir.name):
        applied.append("T")

    if app_dir.name == "nofx" and fix_nofx_deployment(app_dir):
        applied.append("N")

    return applied


def main() -> int:
    for app_dir in sorted(TESTDATA.iterdir()):
        if not app_dir.is_dir() or app_dir.name.startswith("."):
            continue
        applied = repair_app(app_dir)
        if applied:
            print(f"  {app_dir.name}: {','.join(applied)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
