#!/usr/bin/env python3
"""Wire Deployment/StatefulSet spec.replicas to .Values.workloads.<name>.replicaCount.

For every app under testdata/terminus-apps that declares workloadReplicas in
OlaresManifest.yaml, find matching workloads in chart templates and replace
hard-coded replica counts with the values.yaml indirection used by jellyfin
and acestep15v3:

  workloads:
    jellyfin:
      replicaCount: 1

  spec:
    replicas: {{ .Values.workloads.jellyfin.replicaCount }}

A workload matches when its Deployment/StatefulSet metadata.name equals a
workloadReplicas key, or when metadata.name is templated on
{{ .Release.Name }} and the app directory name is listed in workloadReplicas.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

REPO_ROOT = Path(__file__).resolve().parent.parent
TESTDATA = REPO_ROOT / "testdata" / "terminus-apps"

RELEASE_NAME_RES = [
    re.compile(r"^\s*name:\s*\{\{\s*\.Release\.Name\s*\}\}\s*$"),
    re.compile(r'^\s*name:\s*["\']?\{\{\s*\.Release\.Name\s*\}\}["\']?\s*$'),
]

REPLICAS_ALREADY = re.compile(r"\.Values\.workloads\.")


def parse_workload_replicas(manifest: Path) -> Dict[str, int]:
    text = manifest.read_text()
    if not re.search(r"^workloadReplicas:\s*$", text, flags=re.MULTILINE):
        return {}
    out: Dict[str, int] = {}
    m = re.search(r"^workloadReplicas:\s*$", text, flags=re.MULTILINE)
    for line in text[m.end() :].splitlines():
        if line.strip() and not line.startswith(" "):
            break
        mm = re.match(r"^\s+([A-Za-z0-9_-]+):\s*(\d+)\s*$", line)
        if mm:
            out[mm.group(1)] = int(mm.group(2))
    return out


def split_yaml_docs(content: str) -> List[str]:
    docs: List[str] = []
    current: List[str] = []
    for line in content.splitlines(keepends=True):
        if line.strip() == "---":
            if current:
                docs.append("".join(current))
                current = []
            continue
        current.append(line)
    if current:
        docs.append("".join(current))
    return docs if docs else [content]


def doc_kind(doc: str) -> Optional[str]:
    m = re.search(r"^kind:\s*(\S+)\s*$", doc, flags=re.MULTILINE)
    return m.group(1) if m else None


def doc_metadata_name(doc: str) -> Optional[str]:
    # metadata.name may appear anywhere under metadata:
    m = re.search(r"^metadata:\s*$", doc, flags=re.MULTILINE)
    if not m:
        return None
    # scan until next top-level key (no leading whitespace) after metadata block
    lines = doc[m.end() :].splitlines()
    for line in lines:
        if line and not line.startswith(" ") and not line.startswith("\t"):
            break
        mm = re.match(r"^\s+name:\s*(.+?)\s*$", line)
        if mm:
            val = mm.group(1).strip().strip("'\"")
            return val
    return None


def is_release_name_template(name_line_value: str, raw_line: str) -> bool:
    if "Release.Name" in raw_line or "Release.Name" in name_line_value:
        return True
    return False


def resolve_workload_key(
    doc: str, name: Optional[str], workloads: Dict[str, int], app_name: str
) -> Optional[str]:
    if not name:
        return None
    # Find the raw metadata name line for template detection.
    raw = None
    for line in doc.splitlines():
        mm = re.match(r"^\s+name:\s*(.+?)\s*$", line)
        if mm and mm.group(1).strip().strip("'\"") == name:
            raw = line
            break
        if mm and "Release.Name" in line:
            raw = line
            break

    if raw and is_release_name_template(name, raw):
        if app_name in workloads:
            return app_name
        return None

    if name in workloads:
        return name

    # Strip quotes from templated literal names.
    if name in workloads:
        return name

    return None


def replace_spec_replicas(doc: str, workload_key: str) -> Tuple[str, bool]:
    target = f".Values.workloads.{workload_key}.replicaCount"
    if target in doc:
        return doc, False

    # Replace the first replicas: under spec: (Deployment/StatefulSet level).
    lines = doc.splitlines(keepends=True)
    in_spec = False
    spec_indent: Optional[int] = None
    for i, line in enumerate(lines):
        stripped = line.lstrip()
        indent = len(line) - len(stripped)
        if re.match(r"^spec:\s*$", stripped):
            in_spec = True
            spec_indent = indent
            continue
        if not in_spec:
            continue
        # Left spec block when dedented past spec.
        if stripped and indent <= (spec_indent or 0):
            break
        mm = re.match(r"^(\s*)replicas:\s*(.+?)\s*$", line)
        if mm and indent == (spec_indent or 0) + 2:
            repl = f'{mm.group(1)}replicas: {{ {{ .Values.workloads.{workload_key}.replicaCount }} }}\n'
            # fix double braces from f-string
            repl_indent = mm.group(1)
            if "-" in workload_key:
                repl = f'{repl_indent}replicas: {{{{ (index .Values.workloads "{workload_key}").replicaCount }}}}\n'
            else:
                repl = f"{repl_indent}replicas: {{{{ .Values.workloads.{workload_key}.replicaCount }}}}\n"
            if lines[i] != repl:
                lines[i] = repl
                return "".join(lines), True
            return doc, False
    return doc, False


def patch_template_file(path: Path, workloads: Dict[str, int], app_name: str) -> List[str]:
    content = path.read_text()
    docs = split_yaml_docs(content)
    changed_keys: List[str] = []
    new_docs: List[str] = []
    file_changed = False

    for doc in docs:
        kind = doc_kind(doc)
        if kind not in {"Deployment", "StatefulSet"}:
            new_docs.append(doc)
            continue
        name = doc_metadata_name(doc)
        key = resolve_workload_key(doc, name, workloads, app_name)
        if not key:
            new_docs.append(doc)
            continue
        patched, ok = replace_spec_replicas(doc, key)
        new_docs.append(patched)
        if ok:
            changed_keys.append(key)
            file_changed = True

    if not file_changed:
        return []

    # Reassemble with --- separators preserved.
    out_parts: List[str] = []
    for idx, doc in enumerate(new_docs):
        if idx > 0 and not doc.startswith("---"):
            # Original had --- between docs; rebuild conservatively.
            out_parts.append("---\n")
        out_parts.append(doc)
    path.write_text("".join(out_parts))
    return changed_keys


def ensure_values_workloads(values_path: Path, workloads: Dict[str, int]) -> bool:
    text = values_path.read_text() if values_path.exists() else ""
    changed = False
    lines = text.splitlines()
    existing: Dict[str, bool] = {}
    for w in workloads:
        pat = re.compile(rf"^\s{re.escape(w)}:\s*$")
        if any(pat.match(l) for l in lines if "workloads:" in text):
            # deeper check below
            pass

    # Simple append-if-missing using ruamel-free text edit.
    if "workloads:" not in text:
        block = "workloads:\n" + "\n".join(
            f"  {k}:\n    replicaCount: {v}" for k, v in sorted(workloads.items())
        )
        text = text.rstrip() + "\n\n" + block + "\n"
        values_path.write_text(text)
        return True

    for w, count in workloads.items():
        if re.search(rf"^\s{re.escape(w)}:\s*$", text, flags=re.MULTILINE):
            if not re.search(
                rf"^\s{re.escape(w)}:\s*\n\s+replicaCount:\s*{count}\s*$",
                text,
                flags=re.MULTILINE,
            ):
                # workload key exists; leave as-is unless replicaCount missing entirely
                if not re.search(
                    rf"^\s{re.escape(w)}:\s*\n\s+replicaCount:",
                    text,
                    flags=re.MULTILINE,
                ):
                    text = re.sub(
                        rf"(^\s{re.escape(w)}:\s*\n)",
                        rf"\1    replicaCount: {count}\n",
                        text,
                        count=1,
                        flags=re.MULTILINE,
                    )
                    changed = True
        else:
            insert = f"  {w}:\n    replicaCount: {count}\n"
            text = re.sub(
                r"(^workloads:\s*\n)",
                r"\1" + insert,
                text,
                count=1,
                flags=re.MULTILINE,
            )
            changed = True

    if changed:
        values_path.write_text(text)
    return changed


def process_app(app_dir: Path) -> Tuple[List[str], List[str]]:
    manifest = app_dir / "OlaresManifest.yaml"
    if not manifest.exists():
        return [], []
    workloads = parse_workload_replicas(manifest)
    if not workloads:
        return [], []

    values = app_dir / "values.yaml"
    ensure_values_workloads(values, workloads)

    templates_dir = app_dir / "templates"
    if not templates_dir.is_dir():
        return [], list(workloads.keys())

    changed_files: List[str] = []
    missing: List[str] = list(workloads.keys())

    for tpl in sorted(templates_dir.rglob("*")):
        if not tpl.is_file():
            continue
        if tpl.suffix not in {".yaml", ".yml", ".tpl"}:
            continue
        # Treat workloads already wired via .Values.workloads as matched.
        try:
            tpl_text = tpl.read_text()
        except OSError:
            continue
        for w in list(missing):
            if f".Values.workloads.{w}.replicaCount" in tpl_text or f'workloads "{w}"' in tpl_text:
                missing.remove(w)
            elif f'(index .Values.workloads "{w}")' in tpl_text:
                missing.remove(w)
        keys = patch_template_file(tpl, workloads, app_dir.name)
        if keys:
            changed_files.append(str(tpl.relative_to(app_dir)))
            for k in keys:
                if k in missing:
                    missing.remove(k)

    return changed_files, missing


def main() -> int:
    summary = {"apps": 0, "files": 0, "missing": 0}
    for app_dir in sorted(TESTDATA.iterdir()):
        if not app_dir.is_dir() or app_dir.name.startswith("."):
            continue
        changed, missing = process_app(app_dir)
        if not changed and not missing:
            continue
        if changed:
            summary["apps"] += 1
            summary["files"] += len(changed)
            print(f"{app_dir.name}: {', '.join(changed)}")
        if missing:
            summary["missing"] += len(missing)
            print(f"  ! unmatched workloadReplicas keys: {', '.join(missing)}")

    print(
        f"\nupdated {summary['apps']} app(s), {summary['files']} template file(s); "
        f"{summary['missing']} workload key(s) still unmatched"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
