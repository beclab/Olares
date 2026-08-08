import importlib.util
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("validate.py")
SPEC = importlib.util.spec_from_file_location("skill_validate", MODULE_PATH)
validate = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(validate)


VALID_FRONTMATTER = """---
name: olares-test
version: 1.2.3
description: "Test skill"
compatibility: Requires olares-cli
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---
"""


class ValidatorTests(unittest.TestCase):
    def test_indented_fenced_code_is_ignored(self):
        text = "before\n   ```md\n[bad](missing.md)\n# fake\n   ```\nafter\n"
        stripped = validate.without_fenced_code(text)
        self.assertNotIn("missing.md", stripped)
        self.assertNotIn("# fake", stripped)
        self.assertIn("before", stripped)
        self.assertIn("after", stripped)

    def test_frontmatter_rejects_invalid_yaml(self):
        with tempfile.TemporaryDirectory() as directory:
            skill = Path(directory) / "olares-test" / "SKILL.md"
            skill.parent.mkdir()
            skill.write_text(
                VALID_FRONTMATTER.replace(
                    "compatibility: Requires olares-cli",
                    "compatibility: Requires: broken",
                ),
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = Path(directory)
            try:
                errors = []
                validate.validate_frontmatter(skill, errors)
            finally:
                validate.ROOT = original_root
            self.assertTrue(errors)

    def test_cluster_requires_exec_gate_on_both_rows(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            skill_dir = root / "olares-cluster"
            skill_dir.mkdir()
            (skill_dir / "SKILL.md").write_text(
                """| Noun | Verbs | Read when triggered |
|---|---|---|
| `pod` | `list`, `exec` | `exec` requires Olares 1.12.7+ |
| `container` | `list`, `exec` | no version gate |
""",
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_skill_entrypoint(skill_dir, errors)
            finally:
                validate.ROOT = original_root
            self.assertTrue(any("container exec" in error for error in errors), errors)

    def test_front_door_may_link_the_shared_models_but_no_other_peer_reference(self):
        with tempfile.TemporaryDirectory() as directory:
            # Resolved because the validator resolves every link target before
            # taking it relative to ROOT, and /var is a symlink on macOS.
            root = Path(directory).resolve()
            skill_dir = root / "olares-chart"
            (skill_dir / "references").mkdir(parents=True)
            (root / "olares-shared" / "references").mkdir(parents=True)
            (root / "olares-market" / "references").mkdir(parents=True)
            (skill_dir / "SKILL.md").write_text(
                "[platform model](../olares-shared/references/olares-platform.md)\n"
                "[own reference](references/olares-chart-deploy.md)\n"
                "[market front door](../olares-market/SKILL.md)\n"
                "[market charts](../olares-market/references/olares-market-charts.md#download)\n",
                encoding="utf-8",
            )
            # The carve-out is the front door's alone: a reference reaches the
            # shared models by name, relying on that prerequisite.
            (skill_dir / "references" / "olares-chart-deploy.md").write_text(
                "[platform model](../../olares-shared/references/olares-platform.md)\n",
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_structure(skill_dir, errors)
            finally:
                validate.ROOT = original_root
            self.assertEqual(len(errors), 2, errors)
            joined = "\n".join(errors)
            self.assertIn("olares-chart/SKILL.md: deep-links olares-market/references", joined)
            self.assertIn("olares-chart-deploy.md: deep-links olares-shared/references", joined)


if __name__ == "__main__":
    unittest.main()
