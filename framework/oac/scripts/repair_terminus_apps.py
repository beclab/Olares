#!/usr/bin/env python3
"""Apply auto-repairs to testdata/terminus-apps/* OlaresManifest.yaml files
based on the lint failures recorded in terminus_apps_report.md.

Fix taxonomy (one app may receive several):

  T  template-syntax    Set olaresManifest.version to '0.11.0' so the
                        legacy helm-render path is used and the {{ }}
                        placeholders parse cleanly. Stops further fixes
                        for that app -- the schema is now legacy.

  W  workloadReplicas   Insert a top-level workloadReplicas map listing
                        every Deployment / StatefulSet observed in the
                        report's resource diagnostic.

  D  olares-dep         Update the options.dependencies[name=olares]
                        version constraint to the value the validator
                        suggested in its error message.

  P  provider-empty     Drop the legacy top-level provider: section.

  O  spec.onlyAdmin     Add (or set) spec.onlyAdmin: true.

  Cs cluster-scope      Set options.appScope.clusterScoped: false.

  Ar app-ref            Empty options.appScope.appRef.

  E  envs/OLARES_USER   Reported but not auto-fixed (requires renaming
                        the env and threading valueFrom). The script
                        prints these so a human can follow up.
"""

from __future__ import annotations

import os
import re
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

REPO_ROOT = Path(__file__).resolve().parent.parent
REPORT = REPO_ROOT / "terminus_apps_report.md"
TESTDATA = REPO_ROOT / "testdata" / "terminus-apps"


def parse_report(content: str) -> Dict[str, dict]:
    """Return {app_name: {"lint": str, "workloads": [str, ...]}}."""
    apps: Dict[str, dict] = {}
    failure_idx = content.find("## Failure details")
    if failure_idx == -1:
        return apps
    failures = content[failure_idx:]
    sections = re.split(r"^### (\S+)\s*$", failures, flags=re.MULTILINE)
    for i in range(1, len(sections), 2):
        name = sections[i].strip()
        body = sections[i + 1] if i + 1 < len(sections) else ""
        m = re.search(r"^- \*\*Lint\*\*:\s*(.+?)$", body, flags=re.MULTILINE)
        if not m:
            continue
        lint = m.group(1).strip()
        workloads: List[str] = []
        seen = set()
        for row in re.finditer(
            r"^\s*\|\s*[^|]+\|\s*(?:Deployment|StatefulSet)\s*\|\s*([^\s|]+)\s*\|",
            body,
            flags=re.MULTILINE,
        ):
            wl = row.group(1)
            if wl and wl != "-" and wl not in seen:
                workloads.append(wl)
                seen.add(wl)
        apps[name] = {"lint": lint, "workloads": workloads}
    return apps


# -- string-level edits ------------------------------------------------------


def set_manifest_version(text: str, value: str) -> str:
    """Replace the value of the top-level olaresManifest.version line."""
    pattern = re.compile(
        r"^(olaresManifest\.version\s*:\s*)['\"]?[^\n'\"]*['\"]?",
        re.MULTILINE,
    )
    if not pattern.search(text):
        return text
    return pattern.sub(rf"\1'{value}'", text, count=1)


def update_olares_dep_version(text: str, target: str) -> str:
    """Update the version constraint on the olares system dependency.

    Handles:
      - name: olares
        type: system
        version: '>=...'

    and the same with name and type lines reordered. Other layouts
    (version before name, etc.) are uncommon and left untouched.
    """
    pat_name_first = re.compile(
        r"(-\s+name\s*:\s*['\"]?olares['\"]?[^\n]*\n"
        r"(?:[ \t]+(?:type|registry|mandatory|description)\s*:[^\n]*\n)*"
        r"[ \t]+version\s*:\s*)['\"]?[^\n'\"]+['\"]?",
        re.MULTILINE,
    )
    new_text, n = pat_name_first.subn(rf"\1'{target}'", text, count=1)
    if n > 0:
        return new_text

    pat_type_then_name = re.compile(
        r"(-\s+type\s*:\s*['\"]?system['\"]?[^\n]*\n"
        r"[ \t]+name\s*:\s*['\"]?olares['\"]?[^\n]*\n"
        r"(?:[ \t]+(?:mandatory|description)\s*:[^\n]*\n)*"
        r"[ \t]+version\s*:\s*)['\"]?[^\n'\"]+['\"]?",
        re.MULTILINE,
    )
    new_text, n = pat_type_then_name.subn(rf"\1'{target}'", text, count=1)
    return new_text if n > 0 else text


def insert_workload_replicas(text: str, workloads: List[str]) -> str:
    """Insert a top-level workloadReplicas: map after the apiVersion line.

    No-op if the manifest already declares workloadReplicas: at the top
    level, or if workloads is empty.
    """
    if not workloads:
        return text
    if re.search(r"^workloadReplicas\s*:", text, flags=re.MULTILINE):
        return text
    block = "workloadReplicas:\n" + "\n".join(f"  {w}: 1" for w in workloads)
    pattern = re.compile(r"(^apiVersion\s*:[^\n]*\n)", flags=re.MULTILINE)
    if pattern.search(text):
        return pattern.sub(rf"\1\n{block}\n", text, count=1)
    return text + "\n" + block + "\n"


def remove_top_level_provider(text: str) -> str:
    """Drop the top-level provider: block until the next column-0 key.

    The block runs from a line that starts with `provider:` (no leading
    whitespace) up to but not including the next non-indented non-blank
    line. Lines that begin with a `#` comment are also treated as
    column-0 boundaries so we don't gobble adjacent comments.
    """
    lines = text.split("\n")
    out: List[str] = []
    skipping = False
    for line in lines:
        stripped = line.lstrip()
        is_top_level = (
            line
            and not line.startswith((" ", "\t"))
            and not stripped.startswith("#")
            and ":" in line
        )
        if skipping:
            if is_top_level:
                skipping = False
                out.append(line)
            continue
        if is_top_level and re.match(r"provider\s*:", line):
            skipping = True
            continue
        out.append(line)
    return "\n".join(out)


def add_or_set_spec_only_admin(text: str) -> str:
    """Ensure spec.onlyAdmin: true appears under the top-level spec: block.

    If a spec.onlyAdmin: false line already exists it is flipped to
    true; otherwise we insert `  onlyAdmin: true` right after the
    `spec:` header line.
    """
    spec_match = re.search(r"^spec\s*:\s*$", text, flags=re.MULTILINE)
    if not spec_match:
        return text
    spec_block, span = _slice_top_level_block(text, "spec")
    if spec_block is None:
        return text

    flipped = re.sub(
        r"^([ \t]+onlyAdmin\s*:\s*)(?:false|False|FALSE)\s*$",
        r"\1true",
        spec_block,
        flags=re.MULTILINE,
    )
    if flipped != spec_block:
        return text[: span[0]] + flipped + text[span[1]:]

    if re.search(r"^[ \t]+onlyAdmin\s*:", spec_block, flags=re.MULTILINE):
        return text  # already has onlyAdmin (presumably true)

    insert = "  onlyAdmin: true\n"
    new_spec = re.sub(
        r"(^spec\s*:\s*\n)",
        rf"\1{insert}",
        spec_block,
        count=1,
        flags=re.MULTILINE,
    )
    return text[: span[0]] + new_spec + text[span[1]:]


def fix_app_scope(text: str, *, want_cluster_scoped_false: bool, want_app_ref_empty: bool) -> str:
    """Patch options.appScope under the top-level options: block."""
    options_block, span = _slice_top_level_block(text, "options")
    if options_block is None:
        return text
    patched = options_block

    if want_cluster_scoped_false:
        patched = re.sub(
            r"^(\s+clusterScoped\s*:\s*)(?:true|True|TRUE)\s*$",
            r"\1false",
            patched,
            flags=re.MULTILINE,
        )

    if want_app_ref_empty:
        patched = _empty_app_ref_block(patched)

    if patched != options_block:
        return text[: span[0]] + patched + text[span[1]:]
    return text


def _empty_app_ref_block(options_block: str) -> str:
    """Replace `appRef:` plus its list items with `appRef: []`."""
    lines = options_block.split("\n")
    out: List[str] = []
    skipping_indent: Optional[int] = None
    i = 0
    while i < len(lines):
        line = lines[i]
        if skipping_indent is not None:
            stripped = line.lstrip()
            indent = len(line) - len(stripped)
            if not stripped:
                out.append(line)
                i += 1
                continue
            if indent > skipping_indent:
                i += 1
                continue
            skipping_indent = None
        m = re.match(r"^([ \t]+)appRef\s*:\s*$", line)
        if m:
            indent = len(m.group(1).expandtabs())
            out.append(f"{m.group(1)}appRef: []")
            skipping_indent = indent
            i += 1
            continue
        out.append(line)
        i += 1
    return "\n".join(out)


def _slice_top_level_block(text: str, key: str) -> Tuple[Optional[str], Tuple[int, int]]:
    """Return the (text, (start, end)) for a top-level YAML block.

    The block runs from the line `<key>:` (column 0) through the line
    just before the next column-0 non-blank, non-comment line. start
    is the byte offset of `<key>` in text; end is one past the block's
    last byte (so text[start:end] == block).
    """
    pat = re.compile(rf"^({re.escape(key)}\s*:.*?)$", flags=re.MULTILINE)
    m = pat.search(text)
    if not m:
        return None, (-1, -1)
    start = m.start()
    end = len(text)
    for nm in re.finditer(r"^[A-Za-z_][\w.]*\s*:", text[m.end() :], flags=re.MULTILINE):
        end = m.end() + nm.start()
        break
    return text[start:end], (start, end)


# -- per-app driver ----------------------------------------------------------


def classify_lint(lint: str) -> dict:
    info = {
        "template": False,
        "workload_replicas_missing": False,
        "olares_dep_target": None,  # str or None
        "provider_must_be_empty": False,
        "spec_only_admin": False,
        "cluster_scoped_false": False,
        "app_ref_empty": False,
        "envs_olares_user": False,
    }
    if "is template syntax outside any YAML string value" in lint:
        info["template"] = True
        return info

    if "workloadReplicas is required" in lint:
        info["workload_replicas_missing"] = True

    m = re.search(
        r'options\.dependencies\[name=olares\]\.version "[^"]+" must restrict the Olares system version to "([^"]+)"',
        lint,
    )
    if m:
        info["olares_dep_target"] = m.group(1)

    if "provider must be empty" in lint:
        info["provider_must_be_empty"] = True

    if "spec.onlyAdmin must be true when options.shared=true" in lint:
        info["spec_only_admin"] = True

    if "options.appScope.clusterScoped must be false when options.shared=true" in lint:
        info["cluster_scoped_false"] = True

    if "options.appScope.appRef must be empty when options.shared=true" in lint:
        info["app_ref_empty"] = True

    if 'envs[' in lint and 'must not start with "OLARES_USER"' in lint:
        info["envs_olares_user"] = True

    return info


def fix_app(app_dir: Path, info: dict, lint_info: dict) -> List[str]:
    """Apply all fixes that lint_info dictates. Returns list of applied tags."""
    manifest_path = app_dir / "OlaresManifest.yaml"
    if not manifest_path.exists():
        return []
    text = manifest_path.read_text()
    original = text
    applied: List[str] = []

    if lint_info["template"]:
        text = set_manifest_version(text, "0.11.0")
        if text != original:
            applied.append("T")
        manifest_path.write_text(text)
        return applied

    if lint_info["olares_dep_target"]:
        new_text = update_olares_dep_version(text, lint_info["olares_dep_target"])
        if new_text != text:
            applied.append("D")
        text = new_text

    if lint_info["workload_replicas_missing"]:
        new_text = insert_workload_replicas(text, info["workloads"])
        if new_text != text:
            applied.append("W")
        text = new_text

    if lint_info["provider_must_be_empty"]:
        new_text = remove_top_level_provider(text)
        if new_text != text:
            applied.append("P")
        text = new_text

    if lint_info["spec_only_admin"]:
        new_text = add_or_set_spec_only_admin(text)
        if new_text != text:
            applied.append("O")
        text = new_text

    if lint_info["cluster_scoped_false"] or lint_info["app_ref_empty"]:
        new_text = fix_app_scope(
            text,
            want_cluster_scoped_false=lint_info["cluster_scoped_false"],
            want_app_ref_empty=lint_info["app_ref_empty"],
        )
        if new_text != text:
            if lint_info["cluster_scoped_false"]:
                applied.append("Cs")
            if lint_info["app_ref_empty"]:
                applied.append("Ar")
        text = new_text

    if lint_info["envs_olares_user"]:
        applied.append("E?")  # flagged, not auto-fixed

    if text != original:
        manifest_path.write_text(text)
    return applied


def main() -> int:
    if not REPORT.exists():
        print(f"missing report: {REPORT}", file=sys.stderr)
        return 1

    apps = parse_report(REPORT.read_text())
    if not apps:
        print("no failure entries found in report", file=sys.stderr)
        return 1

    print(f"parsed {len(apps)} failing app(s)")
    summary: Dict[str, int] = {}
    env_followups: List[str] = []
    untouched: List[str] = []

    for name, info in sorted(apps.items()):
        app_dir = TESTDATA / name
        if not app_dir.exists():
            print(f"  ! missing app dir: {name}")
            continue
        lint_info = classify_lint(info["lint"])
        applied = fix_app(app_dir, info, lint_info)
        if not applied:
            untouched.append(name)
            continue
        for tag in applied:
            summary[tag] = summary.get(tag, 0) + 1
        if "E?" in applied:
            env_followups.append(name)
        print(f"  {name}: {','.join(applied)}")

    print("\nfix summary:")
    for tag, n in sorted(summary.items()):
        print(f"  {tag}: {n}")
    if env_followups:
        print(f"\nenvs[N] OLARES_USER (manual follow-up): {len(env_followups)} app(s)")
        for n in env_followups:
            print(f"  - {n}")
    if untouched:
        print(f"\nuntouched (no fix produced): {len(untouched)} app(s)")
        for n in untouched[:20]:
            print(f"  - {n}")
        if len(untouched) > 20:
            print(f"  ... and {len(untouched) - 20} more")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
