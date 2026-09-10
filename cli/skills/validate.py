#!/usr/bin/env python3

from __future__ import annotations

import re
import sys
from pathlib import Path
from urllib.parse import unquote

import yaml


ROOT = Path(__file__).resolve().parent
LINK_RE = re.compile(r"(?<!!)\[[^\]]+\]\(([^)]+)\)")
HEADING_RE = re.compile(r"^#{1,6}\s+(.+?)\s*$", re.MULTILINE)
HTML_ANCHOR_RE = re.compile(r'<a\s+(?:name|id)=["\']([^"\']+)["\']', re.IGNORECASE)
SKILL_MAX_LINES = 250
REFERENCE_MAX_LINES = 150
# The per-file ceilings above are satisfiable while the thing an agent pays
# for gets worse: what it reads before its first command is a path across
# files -- the shared front door, the domain SKILL.md, and the references
# this task triggers. A skill states its own common paths in a table; this
# is what one row of it may total.
READ_PATH_MAX_LINES = 250
# The heading that table sits under. Two spellings, because a skill whose
# whole job is diagnosis indexes by what the user reported rather than by
# what the agent means to do, and `## Symptom routing` is the shape the
# README names to copy. Requiring the other name there produced two tables
# routing the same symptoms to the same references, and the shorter one
# was missing a symptom.
FAST_PATHS_HEADINGS = ("Fast paths", "Symptom routing")
# Pointing at a section by quoting its name resolves for nobody and survives
# no rename. Two of the three this suite had were naming headings that no
# longer existed.
PROSE_SECTION_RE = re.compile(r'\b(?:especially|see(?:\s+also)?|section)\s+"([^"\n]{2,80})"')
TABLE_ROW_RE = re.compile(r"^\s*\|(?P<cells>.*)\|\s*$")
BACKTICKED_RE = re.compile(r"`[^`]+`")
# A skill's version is the olares-cli release it ships in, spelled the way npm
# spells it (see .github/workflows/release-cli.yaml, which rejects any other
# shape). The skills are compiled into the binary, so what they document is
# whatever that release's command tree is -- a number of their own would be a
# second thing to bump and a second thing to get wrong.
RELEASE_VERSION_RE = re.compile(r"\d+\.\d+\.\d+-cli\.\d+")
# The one skill whose references every front door is expected to link directly:
# it hosts the platform and app-state models the runtime skills read once.
SHARED_SKILL = "olares-shared"
# Skills with no fast path to declare, and why.
#
# olares-shared is the front door every other path is measured through;
# it is not a destination.
#
# olares-chart and olares-publish describe authoring work -- porting a
# repository into an Olares app, getting a listing through GitBot. Both
# are sequences with a diagnosis at the front, and neither has a first
# command that answers anything, so a fast-path table for them would be
# a fiction written to satisfy this check.
NO_FAST_PATH = {SHARED_SKILL, "olares-chart", "olares-publish"}
REQUIRED_ENTRYPOINT_FACTS = {
    "olares-knowledge/SKILL.md": [
        (
            "knowledge download requires Olares 1.12.7+",
            r"^All verbs require Olares 1\.12\.7\+ because the Settings download edge and provider are both required\. If the version cannot be established, follow the shared profile/auth gate before deciding that an upgrade is needed\.$",
        ),
    ],
    "olares-search/SKILL.md": [
        (
            # Google Drive and Dropbox lost their own verbs to the federated
            # drive search, so the requirement is stated in prose now.
            "cloud sources require Olares 1.12.7+ and a bound integration",
            r"^Google Drive and Dropbox results require Olares 1\.12\.7\+ and a bound integration: ",
        ),
        (
            "knowledge search requires Olares 1.12.7+",
            r"^\| `knowledge` \(`wise`\) \| Wise/Knowledge content search \| requires Olares 1\.12\.7\+; aggregate only \|$",
        ),
    ],
    "olares-cluster/SKILL.md": [
        (
            "pod exec requires Olares 1.12.7+",
            r"^\| `pod` \| `list`, `get`, `yaml`, `events`, `logs`, `delete`, `restart`, `exec` \| `exec` requires Olares 1\.12\.7\+; \[pod operations\]\(references/olares-cluster-pod\.md\); \[exec safety\]\(references/olares-cluster-exec\.md\) \|$",
        ),
        (
            "container exec requires Olares 1.12.7+",
            r"^\| `container` \| `list`, `env`, `logs`, `exec` \| `exec` requires Olares 1\.12\.7\+; \[exec safety\]\(references/olares-cluster-exec\.md\) \|$",
        ),
    ],
    "olares-router/SKILL.md": [
        (
            "router requires Olares 1.12.7+",
            r"^All verbs require Olares 1\.12\.7\+ because Router ships as the `router` Market listing, which asks for that line\. ",
        ),
    ],
    "olares-market/SKILL.md": [
        (
            "canceling resuming/upgrading apps requires Olares 1.12.7+",
            r"^\| lifecycle \| `install`, `upgrade`, `uninstall`, `clone`, `stop`, `resume`, `cancel` \| Canceling `resuming` / `upgrading` requires Olares 1\.12\.7\+; \[lifecycle decisions\]\(references/olares-market-lifecycle\.md\) \|$",
        ),
    ],
}


def without_fenced_code(text: str) -> str:
    result: list[str] = []
    fence_char = ""
    fence_length = 0
    for line in text.splitlines(keepends=True):
        if not fence_char:
            opener = re.match(r"^ {0,3}(`{3,}|~{3,})", line)
            if opener:
                fence = opener.group(1)
                fence_char = fence[0]
                fence_length = len(fence)
                continue
            result.append(line)
            continue
        if re.fullmatch(rf" {{0,3}}{re.escape(fence_char)}{{{fence_length},}}\s*", line):
            fence_char = ""
            fence_length = 0
    return "".join(result)


def markdown_anchor(heading: str) -> str:
    heading = re.sub(r"<[^>]+>", "", heading)
    heading = re.sub(r"[^\w\s-]", "", heading, flags=re.UNICODE)
    return re.sub(r"\s", "-", heading.strip().lower())


def anchors(path: Path) -> set[str]:
    text = without_fenced_code(path.read_text(encoding="utf-8"))
    result: set[str] = set()
    counts: dict[str, int] = {}
    for heading in HEADING_RE.findall(text):
        base = markdown_anchor(heading)
        count = counts.get(base, 0)
        result.add(base if count == 0 else f"{base}-{count}")
        counts[base] = count + 1
    result.update(HTML_ANCHOR_RE.findall(text))
    return result


def validate_links(path: Path, errors: list[str]) -> None:
    text = without_fenced_code(path.read_text(encoding="utf-8"))
    for raw_target in LINK_RE.findall(text):
        target = raw_target.strip().strip("<>")
        if target.startswith(("http://", "https://", "mailto:", "skill://")):
            continue
        file_part, separator, fragment = target.partition("#")
        target_path = path if not file_part else (path.parent / unquote(file_part)).resolve()
        if not target_path.exists():
            errors.append(f"{path.relative_to(ROOT)}: missing link target {target}")
            continue
        if separator and fragment and target_path.suffix.lower() == ".md":
            if unquote(fragment) not in anchors(target_path):
                errors.append(f"{path.relative_to(ROOT)}: missing anchor {target}")


def validate_prose_section_refs(path: Path, errors: list[str]) -> None:
    """Refuse `(especially "Some Heading")` in favour of an anchor link.

    A quoted heading is a link the anchor validator cannot follow and a
    rename cannot update, and it fails silently in the direction that
    matters: the agent goes looking for a section that is not there.
    """
    text = without_fenced_code(path.read_text(encoding="utf-8"))
    for name in PROSE_SECTION_RE.findall(text):
        errors.append(
            f"{path.relative_to(ROOT)}: names a section in prose ({name!r}); "
            "link it with an anchor so a rename fails the build"
        )


def table_rows(text: str) -> list[list[str]]:
    rows: list[list[str]] = []
    for line in text.splitlines():
        match = TABLE_ROW_RE.match(line)
        if not match:
            continue
        cells = [cell.strip() for cell in match.group("cells").split("|")]
        if all(set(cell) <= set("-: ") for cell in cells):
            continue
        rows.append(cells)
    return rows


def validate_verb_index_rows(skill: Path, errors: list[str]) -> None:
    """A row that only points at `--help` spends a line to say nothing.

    The agent already knows `--help` exists; what a verb index owes it is
    the thing `--help` will not say. A row with nothing else to give is
    better collapsed into a single catch-all row than listed on its own.
    """
    text = without_fenced_code(skill.read_text(encoding="utf-8"))
    for cells in table_rows(text):
        cell = cells[-1]
        if "](" in cell or "--help" not in cell:
            continue
        spans = BACKTICKED_RE.findall(cell)
        if not spans or not all(span.rstrip("`").endswith("--help") for span in spans):
            continue
        residue = BACKTICKED_RE.sub("", cell)
        if set(residue) <= set(";,. \t/"):
            errors.append(
                f"{skill.relative_to(ROOT)}: verb-index row {cells[0]} points only at "
                f"{cell} — give the caller what --help will not say, or collapse the row"
            )


def line_count(path: Path) -> int:
    return len(path.read_text(encoding="utf-8").splitlines())


def validate_fast_paths(skill_dir: Path, errors: list[str]) -> None:
    """Hold a skill's declared read paths to the first-command budget.

    Each row totals the shared front door, this SKILL.md, and every file
    the row links -- which is what the agent reads before it can issue the
    command in that row.

    The block was optional to begin with, which left the budget checking
    only the three skills that had volunteered for it. A skill with no
    block was not under budget; it was unmeasured, and the two paths that
    turned out to be over were both found by writing a block down.

    So it is required, except of the skills in NO_FAST_PATH. Those are
    exempt because a fast path is a claim about a first command, and the
    work they describe has no first command to name -- the exemption is
    the honest answer for them, not a hole to hide a heavy skill in.
    """
    skill = skill_dir / "SKILL.md"
    text = without_fenced_code(skill.read_text(encoding="utf-8"))
    heading = "|".join(re.escape(name) for name in FAST_PATHS_HEADINGS)
    spellings = " or ".join(f"'## {name}'" for name in FAST_PATHS_HEADINGS)
    section = re.split(rf"^#{{1,6}}\s+(?:{heading})\s*$", text, flags=re.MULTILINE)
    if len(section) < 2:
        if skill_dir.name not in NO_FAST_PATH:
            errors.append(
                f"{skill.relative_to(ROOT)}: no {spellings} block — name the tasks an "
                f"agent can act on after one read, so the {READ_PATH_MAX_LINES}-line budget has "
                "something to measure"
            )
        return
    if skill_dir.name in NO_FAST_PATH:
        errors.append(
            f"{skill.relative_to(ROOT)}: declares fast paths but is listed in NO_FAST_PATH; "
            "remove it from that list"
        )
    if len(section) > 2:
        # Both spellings in one file is how doctor ended up routing the same
        # symptoms twice, so it is refused rather than measured.
        errors.append(
            f"{skill.relative_to(ROOT)}: has more than one of {spellings}; one table routes the "
            "agent, a second one competes with it"
        )
    body = re.split(r"^#{1,6}\s+", section[1], flags=re.MULTILINE)[0]
    if not table_rows(body):
        errors.append(f"{skill.relative_to(ROOT)}: its {spellings} block has no rows")
        return

    front_door = ROOT / SHARED_SKILL / "SKILL.md"
    base = line_count(skill)
    if skill_dir.name != SHARED_SKILL:
        base += line_count(front_door)

    for cells in table_rows(body):
        total = base
        seen: set[Path] = set()
        for target in LINK_RE.findall(" ".join(cells)):
            link = unquote(target.strip().strip("<>").partition("#")[0])
            if not link or "://" in link:
                continue
            resolved = (skill.parent / link).resolve()
            if not resolved.exists() or resolved in seen or resolved == skill or resolved == front_door:
                continue
            seen.add(resolved)
            total += line_count(resolved)
        if total > READ_PATH_MAX_LINES:
            errors.append(
                f"{skill.relative_to(ROOT)}: fast path {cells[0]} reads {total} lines, over the "
                f"{READ_PATH_MAX_LINES}-line first-command budget — split what it links, or stop calling it fast"
            )


def validate_skill_entrypoint(skill_dir: Path, errors: list[str]) -> None:
    skill = skill_dir / "SKILL.md"
    text = without_fenced_code(skill.read_text(encoding="utf-8"))
    relative_skill = str(skill.relative_to(ROOT))
    for description, pattern in REQUIRED_ENTRYPOINT_FACTS.get(relative_skill, []):
        if not re.search(pattern, text, re.MULTILINE):
            errors.append(f"{relative_skill}: missing entrypoint fact: {description}")

    linked_files = {
        unquote(raw_target.strip().strip("<>").partition("#")[0])
        for raw_target in LINK_RE.findall(text)
    }
    for reference in sorted((skill_dir / "references").glob("*.md")) if (skill_dir / "references").exists() else []:
        expected = f"references/{reference.name}"
        if expected not in linked_files:
            errors.append(f"{skill.relative_to(ROOT)}: reference is not one hop from SKILL.md: {expected}")

    forbidden = {
        "CRITICAL — before doing anything": "use the thin shared prerequisite wording",
        "Anything outside this scope": "route through the shared suite map without repeating it",
    }
    for phrase, guidance in forbidden.items():
        if phrase in text:
            errors.append(f"{skill.relative_to(ROOT)}: forbidden phrase {phrase!r}; {guidance}")

    source_citation = re.search(r"`?cli/(?:cmd|pkg|internal)/[^`\s]+\.go(?::\d+)?`?", text)
    if source_citation:
        errors.append(
            f"{skill.relative_to(ROOT)}: Go source citation {source_citation.group(0)!r} belongs in verification, not the shipped skill"
        )


def validate_frontmatter(skill: Path, errors: list[str]) -> None:
    lines = skill.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "---":
        errors.append(f"{skill.relative_to(ROOT)}: frontmatter must start with ---")
        return
    try:
        end = lines.index("---", 1)
    except ValueError:
        errors.append(f"{skill.relative_to(ROOT)}: frontmatter is missing its closing ---")
        return

    class UniqueKeyLoader(yaml.SafeLoader):
        pass

    def construct_mapping(loader, node, deep=False):
        mapping = {}
        for key_node, value_node in node.value:
            key = loader.construct_object(key_node, deep=deep)
            if key in mapping:
                raise yaml.constructor.ConstructorError(
                    "while constructing a mapping",
                    node.start_mark,
                    f"found duplicate key {key!r}",
                    key_node.start_mark,
                )
            mapping[key] = loader.construct_object(value_node, deep=deep)
        return mapping

    UniqueKeyLoader.add_constructor(
        yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
        construct_mapping,
    )
    try:
        root = yaml.load("\n".join(lines[1:end]), Loader=UniqueKeyLoader)
    except yaml.YAMLError as error:
        errors.append(f"{skill.relative_to(ROOT)}: invalid YAML frontmatter: {error}")
        return
    if not isinstance(root, dict):
        errors.append(f"{skill.relative_to(ROOT)}: frontmatter must be a mapping")
        return

    allowed_keys = {"name", "version", "description", "compatibility", "metadata"}
    unknown_keys = sorted(set(root) - allowed_keys)
    if unknown_keys:
        errors.append(f"{skill.relative_to(ROOT)}: unknown frontmatter keys: {', '.join(unknown_keys)}")

    expected_name = skill.parent.name
    if root.get("name") != expected_name:
        errors.append(f"{skill.relative_to(ROOT)}: frontmatter name must be {expected_name!r}")
    if not isinstance(root.get("version"), str) or not RELEASE_VERSION_RE.fullmatch(root["version"]):
        errors.append(
            f"{skill.relative_to(ROOT)}: frontmatter version must be the CLI release "
            "it ships in, x.y.z-cli.n"
        )
    for key in ("description", "compatibility"):
        if not isinstance(root.get(key), str) or not root[key].strip():
            errors.append(f"{skill.relative_to(ROOT)}: frontmatter {key!r} is required")

    if root.get("metadata") != {
        "openclaw": {
            "requires": {
                "bins": ["olares-cli"],
            },
        },
    }:
        errors.append(
            f"{skill.relative_to(ROOT)}: frontmatter metadata.openclaw.requires.bins must include olares-cli"
        )


def validate_structure(skill_dir: Path, errors: list[str]) -> None:
    """Enforce the README's size limit and its ban on cross-skill deep links.

    One-hop reachability is already checked against the SKILL.md's own links,
    above; this covers the other direction. A pointer between two references
    of the same skill is allowed on purpose -- both ends are already one hop
    from the front door -- but pointing into a peer skill's references lands
    the agent on a file whose prerequisites it has not loaded.

    The SKILL.md is checked too, with one carve-out the README requires: the
    shared platform and app-state models are linked one hop from every front
    door, because every runtime skill loads them as prerequisites. Any other
    peer skill has to be reached through its own SKILL.md.
    """
    skill = skill_dir / "SKILL.md"
    references = sorted(skill_dir.glob("references/*.md"))

    for path, limit in [(skill, SKILL_MAX_LINES)] + [(ref, REFERENCE_MAX_LINES) for ref in references]:
        count = len(path.read_text(encoding="utf-8").splitlines())
        if count > limit:
            errors.append(f"{path.relative_to(ROOT)}: {count} lines exceeds the {limit}-line limit")

    for path in [skill] + references:
        for target in LINK_RE.findall(without_fenced_code(path.read_text(encoding="utf-8"))):
            link = unquote(target.split("#", 1)[0]).strip()
            if not link or "://" in link:
                continue
            try:
                rel = (path.parent / link).resolve().relative_to(ROOT)
            except ValueError:
                continue
            if len(rel.parts) < 2 or rel.parts[1] != "references" or rel.parts[0] == skill_dir.name:
                continue
            if path == skill and rel.parts[0] == SHARED_SKILL:
                continue
            errors.append(
                f"{path.relative_to(ROOT)}: deep-links {rel} — reach a peer skill through "
                "its SKILL.md, not by pointing into its references"
            )


def validate_one_version(skill_dirs: list[Path], errors: list[str]) -> None:
    """The suite ships as one artifact, so it carries one version.

    Every skill is compiled into the same binary and installed by the same
    command, and `olares-cli` says so on startup when what is installed came
    from a different release. Twelve numbers moving independently made that
    sentence unsayable, and made "which release is this skill from" a question
    with twelve answers.
    """
    declared: dict[str, list[str]] = {}
    for skill_dir in skill_dirs:
        version = read_frontmatter_version(skill_dir / "SKILL.md")
        if version is not None:
            declared.setdefault(version, []).append(skill_dir.name)
    if len(declared) > 1:
        spelled = "; ".join(
            f"{version}: {', '.join(sorted(skills))}" for version, skills in sorted(declared.items())
        )
        errors.append(
            f"the suite declares {len(declared)} versions but ships as one artifact — {spelled}"
        )


def read_frontmatter_version(skill: Path) -> str | None:
    for line in skill.read_text(encoding="utf-8").splitlines()[1:]:
        if line == "---":
            return None
        if line.startswith("version:"):
            return line.split(":", 1)[1].strip().strip('"')
    return None


def main() -> int:
    errors: list[str] = []
    skill_dirs = sorted(path.parent for path in ROOT.glob("olares-*/SKILL.md"))
    for path in sorted(ROOT.glob("olares-*/**/*.md")):
        validate_links(path, errors)
        validate_prose_section_refs(path, errors)
    for skill_dir in skill_dirs:
        validate_frontmatter(skill_dir / "SKILL.md", errors)
        validate_skill_entrypoint(skill_dir, errors)
        validate_structure(skill_dir, errors)
        validate_verb_index_rows(skill_dir / "SKILL.md", errors)
        validate_fast_paths(skill_dir, errors)
    validate_one_version(skill_dirs, errors)

    if errors:
        print("Skill validation failed:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"Skill validation OK ({len(skill_dirs)} skills)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
