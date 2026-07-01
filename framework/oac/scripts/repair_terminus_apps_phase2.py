#!/usr/bin/env python3
"""Phase-2 repairs for testdata/terminus-apps after the manifest pass.

Fixes applied per app (when applicable):

  V  values.yaml workloads.<name>.replicaCount from workloadReplicas
  X  permission.externalData: true when templates reference .Values.sharedlib
  M  middleware workloadReplicas + values.yaml (metadata.name)
  L  legacy spec resource envelope for olaresManifest.version < 0.12.0
"""

from __future__ import annotations

import re
import sys
from io import StringIO
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from ruamel.yaml import YAML
from ruamel.yaml.error import YAMLError

REPO_ROOT = Path(__file__).resolve().parent.parent
TESTDATA = REPO_ROOT / "testdata" / "terminus-apps"

yaml_loader = YAML()
yaml_loader.preserve_quotes = True
yaml_loader.width = 4096

ENV_RENAMES = {
    "OLARES_USER_HUGGINGFACE_TOKEN": "HF_TOKEN",
    "OLARES_USER_HUGGINGFACE_SERVICE": "HF_ENDPOINT",
    "OLARES_USER_TIMEZONE": "APP_TIMEZONE",
}


def try_parse_yaml(text: str) -> Optional[dict]:
    try:
        data = yaml_loader.load(text)
        return data if isinstance(data, dict) else {}
    except YAMLError:
        return None


def dump_yaml(data: dict) -> str:
    buf = StringIO()
    yaml_loader.dump(data, buf)
    return buf.getvalue()


def extract_workload_replicas_regex(text: str) -> Dict[str, int]:
    m = re.search(r"^workloadReplicas:\s*$", text, flags=re.MULTILINE)
    if not m:
        return {}
    out: Dict[str, int] = {}
    for line in text[m.end() :].splitlines():
        if not line.startswith(" ") and line.strip():
            break
        mm = re.match(r"^\s+([A-Za-z0-9_-]+):\s*(\d+)\s*$", line)
        if mm:
            out[mm.group(1)] = int(mm.group(2))
    return out


def extract_workload_replicas(data: Optional[dict], text: str) -> Dict[str, int]:
    if data:
        wr = data.get("workloadReplicas")
        if wr:
            return {str(k): int(v) for k, v in wr.items()}
    return extract_workload_replicas_regex(text)


def ensure_values_workloads(values_path: Path, workloads: Dict[str, int]) -> bool:
    if not workloads:
        return False
    if values_path.exists():
        values = try_parse_yaml(values_path.read_text()) or {}
    else:
        values = {}

    existing = values.get("workloads") or {}
    changed = False
    for name, count in workloads.items():
        entry = existing.get(name) or {}
        if entry.get("replicaCount") != count:
            entry["replicaCount"] = count
            existing[name] = entry
            changed = True

    if not changed and values.get("workloads"):
        return False

    values["workloads"] = existing
    values_path.write_text(dump_yaml(values))
    return True


def uses_sharedlib(app_dir: Path) -> bool:
    for p in app_dir.rglob("*"):
        if not p.is_file() or p.suffix not in {".yaml", ".yml", ".tpl"}:
            continue
        try:
            if ".Values.sharedlib" in p.read_text():
                return True
        except OSError:
            continue
    return False


def ensure_external_data_struct(data: dict) -> bool:
    perm = data.get("permission")
    if not isinstance(perm, dict):
        perm = {}
        data["permission"] = perm
    if perm.get("externalData") is True:
        return False
    perm["externalData"] = True
    return True


def ensure_external_data_regex(text: str) -> Tuple[str, bool]:
    if re.search(r"^\s*externalData:\s*true\s*$", text, flags=re.MULTILINE):
        return text, False
    m = re.search(r"^permission:\s*$", text, flags=re.MULTILINE)
    if not m:
        return text, False
    insert_at = m.end()
    return text[:insert_at] + "\n  externalData: true" + text[insert_at:], True


def manifest_version(text: str, data: Optional[dict]) -> str:
    if data and data.get("olaresManifest.version") is not None:
        return str(data["olaresManifest.version"]).strip("'\"")
    m = re.search(r"^olaresManifest\.version:\s*['\"]?([^'\"\n]+)", text, flags=re.MULTILINE)
    return m.group(1).strip() if m else ""


def is_legacy_version(version: str) -> bool:
    m = re.match(r"^(\d+)\.(\d+)\.(\d+)", version)
    if not m:
        return False
    return tuple(map(int, m.groups())) < (0, 12, 0)


def first_accelerator_fields(text: str, data: Optional[dict]) -> Dict[str, str]:
    if data:
        spec = data.get("spec") or {}
        acc = spec.get("accelerator")
        if isinstance(acc, list) and acc and isinstance(acc[0], dict):
            return {k: str(v) for k, v in acc[0].items() if k != "mode"}

    # Regex fallback: first list item under spec.accelerator.
    m = re.search(r"^\s*accelerator:\s*$", text, flags=re.MULTILINE)
    if not m:
        return {}
    fields: Dict[str, str] = {}
    for line in text[m.end() :].splitlines():
        if line.startswith("  - mode:"):
            continue
        mm = re.match(r"^\s+([A-Za-z]+):\s*(.+?)\s*$", line)
        if not mm:
            if line.startswith("  - ") or (line and not line.startswith(" ")):
                break
            continue
        key, val = mm.group(1), mm.group(2).strip("'\"")
        if key != "mode":
            fields[key] = val
    return fields


def legacy_envelope_missing(text: str, data: Optional[dict]) -> bool:
    keys = ("requiredCpu", "limitedCpu", "requiredMemory", "limitedMemory", "requiredDisk")
    if data:
        spec = data.get("spec") or {}
        return not all(spec.get(k) for k in keys)
    for k in keys:
        if re.search(rf"^\s{re.escape(k)}:\s+\S", text, flags=re.MULTILINE):
            return False
    return True


def ensure_legacy_resources_struct(data: dict) -> bool:
    spec = data.get("spec")
    if not isinstance(spec, dict):
        return False
    acc = (spec.get("accelerator") or [None])[0]
    if not isinstance(acc, dict):
        return False
    mapping = (
        "requiredCpu",
        "limitedCpu",
        "requiredMemory",
        "limitedMemory",
        "requiredDisk",
        "limitedDisk",
    )
    changed = False
    for key in mapping:
        if acc.get(key) and not spec.get(key):
            spec[key] = acc[key]
            changed = True
    return changed


def ensure_legacy_resources_regex(text: str) -> Tuple[str, bool]:
    fields = first_accelerator_fields(text, None)
    if not fields:
        return text, False
    changed = False
    out = text
    insert_lines: List[str] = []
    for src in ("requiredCpu", "limitedCpu", "requiredMemory", "limitedMemory", "requiredDisk", "limitedDisk"):
        if not fields.get(src):
            continue
        if re.search(rf"^\s{re.escape(src)}:\s+\S", out, flags=re.MULTILINE):
            continue
        insert_lines.append(f"  {src}: {fields[src]}")
        changed = True
    if not changed:
        return text, False
    m = re.search(r"^(\s*)supportArch:\s*$", out, flags=re.MULTILINE)
    if m:
        pos = m.start()
        block = "\n".join(insert_lines) + "\n"
        out = out[:pos] + block + out[pos:]
    return out, True


def fix_olares_user_env_names_struct(data: dict) -> bool:
    envs = data.get("envs")
    if not isinstance(envs, list):
        return False
    changed = False
    for entry in envs:
        if not isinstance(entry, dict):
            continue
        name = entry.get("envName")
        if not isinstance(name, str) or not name.startswith("OLARES_USER_"):
            continue
        new_name = ENV_RENAMES.get(name, name.removeprefix("OLARES_USER_"))
        if entry.get("envName") != new_name:
            entry["envName"] = new_name
            changed = True
        vf = entry.get("valueFrom")
        if isinstance(vf, dict) and not vf.get("envName"):
            vf["envName"] = name
            changed = True
    return changed


def fix_olares_user_env_names_regex(text: str) -> Tuple[str, bool]:
    lines = text.splitlines()
    new_lines: List[str] = []
    changed = False
    for line in lines:
        m = re.match(r"^(\s*)- envName:\s*(OLARES_USER_[A-Z0-9_]+)\s*$", line)
        if m:
            indent, old = m.group(1), m.group(2)
            new = ENV_RENAMES.get(old, old.removeprefix("OLARES_USER_"))
            new_lines.append(f"{indent}- envName: {new}")
            changed = True
            continue
        new_lines.append(line)
    out = "\n".join(new_lines) + ("\n" if text.endswith("\n") else "")
    return out, changed


def ensure_middleware_workloads_struct(data: dict) -> Tuple[Dict[str, int], bool]:
    if str(data.get("olaresManifest.type", "")).strip("'\"") != "middleware":
        return {}, False
    name = (data.get("metadata") or {}).get("name")
    if not name:
        return {}, False
    name = str(name)
    wr = data.get("workloadReplicas")
    if wr and name in wr:
        return {name: 1}, False
    if wr is None:
        data["workloadReplicas"] = {name: 1}
    else:
        wr[name] = 1
    return {name: 1}, True


def ensure_middleware_workloads_regex(text: str) -> Tuple[str, Dict[str, int], bool]:
    if not re.search(r"^olaresManifest\.type:\s*['\"]?middleware", text, flags=re.MULTILINE):
        return text, {}, False
    m = re.search(r"^metadata:\s*$", text, flags=re.MULTILINE)
    if not m:
        return text, {}, False
    name = None
    for line in text[m.end() :].splitlines():
        if line and not line.startswith(" "):
            break
        mm = re.match(r"^\s+name:\s*(\S+)", line)
        if mm:
            name = mm.group(1)
            break
    if not name:
        return text, {}, False
    if re.search(rf"^\s{re.escape(name)}:\s*\d+\s*$", text, flags=re.MULTILINE):
        return text, {name: 1}, False
    block = f"\nworkloadReplicas:\n  {name}: 1\n"
    m2 = re.search(r"^apiVersion\s*:[^\n]*\n", text, flags=re.MULTILINE)
    if m2:
        pos = m2.end()
        return text[:pos] + block + text[pos:], {name: 1}, True
    return text + block, {name: 1}, True


def repair_app(app_dir: Path) -> List[str]:
    manifest_path = app_dir / "OlaresManifest.yaml"
    if not manifest_path.exists():
        return []

    text = manifest_path.read_text()
    data = try_parse_yaml(text)
    applied: List[str] = []
    manifest_changed = False
    mw_workloads: Dict[str, int] = {}

    if data is not None:
        mw_workloads, mw_changed = ensure_middleware_workloads_struct(data)
        if mw_changed:
            applied.append("M")
            manifest_changed = True
    else:
        text, mw_workloads, mw_changed = ensure_middleware_workloads_regex(text)
        if mw_changed:
            applied.append("M")
            manifest_changed = True

    workloads = extract_workload_replicas(data, text)
    workloads = {**workloads, **mw_workloads}

    values_path = app_dir / "values.yaml"
    if workloads and ensure_values_workloads(values_path, workloads):
        applied.append("V")

    if uses_sharedlib(app_dir):
        version = manifest_version(text, data)
        if not is_legacy_version(version):
            if data is not None:
                if ensure_external_data_struct(data):
                    applied.append("X")
                    manifest_changed = True
            else:
                text, ok = ensure_external_data_regex(text)
                if ok:
                    applied.append("X")
                    manifest_changed = True

    version = manifest_version(text, data)
    if is_legacy_version(version) and legacy_envelope_missing(text, data):
        if data is not None:
            if ensure_legacy_resources_struct(data):
                applied.append("L")
                manifest_changed = True
        else:
            text, ok = ensure_legacy_resources_regex(text)
            if ok:
                applied.append("L")
                manifest_changed = True

    if manifest_changed:
        if data is not None:
            manifest_path.write_text(dump_yaml(data))
        else:
            manifest_path.write_text(text)

    return applied


def main() -> int:
    if not TESTDATA.is_dir():
        print(f"missing testdata dir: {TESTDATA}", file=sys.stderr)
        return 1

    summary: Dict[str, int] = {}
    for app_dir in sorted(TESTDATA.iterdir()):
        if not app_dir.is_dir() or app_dir.name.startswith("."):
            continue
        applied = repair_app(app_dir)
        if not applied:
            continue
        for tag in applied:
            summary[tag] = summary.get(tag, 0) + 1
        print(f"  {app_dir.name}: {','.join(applied)}")

    print("\nfix summary:")
    for tag, n in sorted(summary.items()):
        print(f"  {tag}: {n}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
