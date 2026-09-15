import importlib.util
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("validate.py")
SPEC = importlib.util.spec_from_file_location("skill_validate", MODULE_PATH)
validate = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(validate)

STAMP_PATH = Path(__file__).with_name("stamp.py")
STAMP_SPEC = importlib.util.spec_from_file_location("skill_stamp", STAMP_PATH)
stamp = importlib.util.module_from_spec(STAMP_SPEC)
assert STAMP_SPEC.loader is not None
STAMP_SPEC.loader.exec_module(stamp)


VALID_FRONTMATTER = """---
name: olares-test
version: 1.12.7-cli.4
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

    def test_version_must_name_the_cli_release(self):
        # A bare x.y.z is a number of the skill's own, which is what the suite
        # stopped having: it ships inside one binary, so it carries that
        # binary's release or nothing anybody can act on.
        for version, valid in [("1.12.7-cli.4", True), ("1.12.7", False), ("4.18.0", False)]:
            with tempfile.TemporaryDirectory() as directory:
                skill = Path(directory) / "olares-test" / "SKILL.md"
                skill.parent.mkdir()
                skill.write_text(
                    VALID_FRONTMATTER.replace("version: 1.12.7-cli.4", f"version: {version}"),
                    encoding="utf-8",
                )
                original_root = validate.ROOT
                validate.ROOT = Path(directory)
                try:
                    errors = []
                    validate.validate_frontmatter(skill, errors)
                finally:
                    validate.ROOT = original_root
                self.assertEqual(not errors, valid, f"{version}: {errors}")

    def test_the_suite_declares_one_version(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            for name, version in [("olares-a", "1.12.7-cli.4"), ("olares-b", "1.12.7-cli.3")]:
                skill_dir = root / name
                skill_dir.mkdir()
                (skill_dir / "SKILL.md").write_text(
                    VALID_FRONTMATTER.replace("version: 1.12.7-cli.4", f"version: {version}"),
                    encoding="utf-8",
                )
            errors = []
            validate.validate_one_version(sorted(root.iterdir()), errors)
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("ships as one artifact", errors[0])
            self.assertIn("1.12.7-cli.3: olares-b", errors[0])

    def test_publish_script_lists_every_skill(self):
        """publish.sh names its slugs by hand, so a new skill is invisible to it.

        Nothing fails when that happens: the script publishes the eleven it
        knows about and reports success, and the twelfth is missing from the
        registry until somebody notices. The binary is unaffected -- it embeds
        by pattern -- which is exactly why this needs a test rather than a
        reader.
        """
        script = (MODULE_PATH.parent / "publish.sh").read_text(encoding="utf-8")
        # Closed on a line of its own; the display names contain parentheses.
        block = script.split("SKILLS=(\n", 1)[1].split("\n)", 1)[0]
        listed = {line.strip().strip('"').split("|", 1)[0] for line in block.splitlines() if "|" in line}
        on_disk = {path.parent.name for path in MODULE_PATH.parent.glob("olares-*/SKILL.md")}
        self.assertEqual(listed, on_disk)

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

    def test_a_section_named_in_prose_is_refused(self):
        """The form the anchor validator cannot follow is the form that rots.

        Both of the ones this suite shipped were naming headings that had
        been renamed out from under them, and nothing said so.
        """
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            reference = root / "olares-test" / "references" / "r.md"
            reference.parent.mkdir(parents=True)
            reference.write_text(
                'Read the parent (especially "OpType vs State") first.\n',
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_prose_section_refs(reference, errors)
            finally:
                validate.ROOT = original_root
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("OpType vs State", errors[0])

    def test_an_anchor_link_to_the_same_section_is_accepted(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            reference = root / "olares-test" / "references" / "r.md"
            reference.parent.mkdir(parents=True)
            reference.write_text(
                "Read [OpType vs State](../SKILL.md#optype-vs-state) first.\n",
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_prose_section_refs(reference, errors)
            finally:
                validate.ROOT = original_root
            self.assertEqual(errors, [])

    def test_a_verb_row_that_only_points_at_help_is_refused(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            skill = root / "olares-test" / "SKILL.md"
            skill.parent.mkdir(parents=True)
            skill.write_text(
                "| Area | Verbs | Read when triggered |\n"
                "|---|---|---|\n"
                "| `video` | `config get` | `settings video --help` |\n"
                "| `gpu` | list | `settings gpu --help`; `settings compute --help` |\n"
                "| `users` | list | [user lifecycle](references/u.md) |\n"
                "| `pod` | `exec` | `exec` needs 1.12.7+; `cluster pod --help` |\n",
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_verb_index_rows(skill, errors)
            finally:
                validate.ROOT = original_root
            # The linked row and the one carrying a version gate both earn
            # their line; only the two bare pointers do not.
            self.assertEqual(len(errors), 2, errors)
            self.assertIn("`video`", errors[0])
            self.assertIn("`gpu`", errors[1])

    def test_a_fast_path_over_the_budget_is_refused(self):
        """Per-file ceilings pass while the path an agent reads does not."""
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            (root / validate.SHARED_SKILL).mkdir(parents=True)
            (root / validate.SHARED_SKILL / "SKILL.md").write_text(
                "front door\n" * 100, encoding="utf-8"
            )
            skill_dir = root / "olares-test"
            (skill_dir / "references").mkdir(parents=True)
            (skill_dir / "references" / "big.md").write_text("x\n" * 149, encoding="utf-8")
            (skill_dir / "references" / "small.md").write_text("x\n" * 40, encoding="utf-8")
            (skill_dir / "SKILL.md").write_text(
                "## Fast paths\n\n"
                "| Task | Read | First command |\n"
                "|---|---|---|\n"
                "| install | [big](references/big.md) | `olares-cli market install` |\n"
                "| list | [small](references/small.md) | `olares-cli market list` |\n",
                encoding="utf-8",
            )
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_fast_paths(skill_dir, errors)
            finally:
                validate.ROOT = original_root
            self.assertEqual(len(errors), 1, errors)
            self.assertIn("fast path install", errors[0])
            self.assertIn("first-command budget", errors[0])

    def test_a_skill_without_fast_paths_is_refused(self):
        """An undeclared path is unmeasured, not under budget.

        While the block was optional the budget above checked the three
        skills that had volunteered for it, and both paths that turned
        out to be over were found by writing a block down.
        """
        errors = self.fast_path_errors_for("olares-test", "## Verb index\n")
        self.assertEqual(len(errors), 1, errors)
        self.assertIn("'## Fast paths'", errors[0])

    def test_a_symptom_table_counts_as_the_declaration(self):
        """A diagnosis skill indexes by what the user reported.

        Requiring the other spelling of the heading is what produced two
        tables in olares-doctor routing the same symptoms to the same
        references, one of them written a symptom short.
        """
        errors = self.fast_path_errors_for(
            "olares-test",
            "## Symptom routing\n\n| Symptom | Reference |\n|---|---|\n"
            "| an app will not start | [stuck](references/stuck.md) |\n",
        )
        self.assertEqual(errors, [])

    def test_declaring_both_spellings_is_refused(self):
        errors = self.fast_path_errors_for(
            "olares-test",
            "## Fast paths\n\n| Task | Read | First command |\n|---|---|---|\n"
            "| start it | this file | `olares-cli market install` |\n\n"
            "## Symptom routing\n\n| Symptom | Reference |\n|---|---|\n"
            "| it will not start | [stuck](references/stuck.md) |\n",
        )
        self.assertEqual(len(errors), 1, errors)
        self.assertIn("more than one", errors[0])

    def test_the_skills_with_no_first_command_are_exempt(self):
        for name in sorted(validate.NO_FAST_PATH):
            with self.subTest(skill=name):
                self.assertEqual(self.fast_path_errors_for(name, "## Verb index\n"), [])

    def test_an_exempt_skill_that_declares_fast_paths_is_told_to_leave_the_list(self):
        errors = self.fast_path_errors_for(
            sorted(validate.NO_FAST_PATH)[0],
            "## Fast paths\n\n| Task | Read | First command |\n|---|---|---|\n"
            "| list | this file | `olares-cli market list` |\n",
        )
        self.assertEqual(len(errors), 1, errors)
        self.assertIn("NO_FAST_PATH", errors[0])

    def test_an_empty_fast_paths_block_is_refused(self):
        errors = self.fast_path_errors_for("olares-test", "## Fast paths\n\nSoon.\n")
        self.assertEqual(len(errors), 1, errors)
        self.assertIn("no rows", errors[0])

    def fast_path_errors_for(self, skill_name, body):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            (root / validate.SHARED_SKILL).mkdir(parents=True, exist_ok=True)
            (root / validate.SHARED_SKILL / "SKILL.md").write_text("front door\n", encoding="utf-8")
            skill_dir = root / skill_name
            skill_dir.mkdir(parents=True, exist_ok=True)
            (skill_dir / "SKILL.md").write_text(body, encoding="utf-8")
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_fast_paths(skill_dir, errors)
            finally:
                validate.ROOT = original_root
            return errors

    def test_a_reference_may_not_cite_go_source_either(self):
        """The check used to run on front doors only, and on `cli/` paths only.

        Both citations it missed were in a reference, and neither named
        the repository root -- one pointed inside app-service, the other
        at a package path.
        """
        for citation in (
            "the loader (`controllers/load.go`, `LoadStatefulApp`) passes it",
            "do not read it off `pkg/appstate/state_transition.go`",
            "see cli/cmd/ctl/files/path.go:42",
        ):
            with self.subTest(citation=citation):
                errors = self.citation_errors_for(citation)
                self.assertEqual(len(errors), 1, errors)
                self.assertIn("belongs in verification", errors[0])

    def test_prose_that_merely_mentions_a_state_handler_is_left_alone(self):
        self.assertEqual(
            self.citation_errors_for("app-service's reconciler is what actually runs."), []
        )

    def citation_errors_for(self, body):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            path = root / "olares-test" / "references" / "model.md"
            path.parent.mkdir(parents=True)
            path.write_text(body + "\n", encoding="utf-8")
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_no_source_citations(path, errors)
            finally:
                validate.ROOT = original_root
            return errors

    def test_the_three_shared_sections_are_required_under_one_name(self):
        """Five names for the closing section, three for the verb index.

        An agent crossing from one skill to another had to scan for both
        rather than jump to a heading.
        """
        complete = (
            "> **Shared front door:** load the shared skill.\n\n"
            "## Verb index\n\n| Verb | Purpose | Read when triggered |\n|---|---|---|\n"
            "| `list` | list apps | [list](references/list.md) |\n\n"
            "## Safety and escalation\n\n- Stop when asked to.\n"
        )
        self.assertEqual(self.shared_section_errors_for("olares-test", complete), [])

        for missing, expected in [
            ("## Safety and escalation", "'## Safety and escalation'"),
            ("## Verb index", "'## Verb index'"),
            ("> **Shared front door:**", "Shared front door"),
        ]:
            with self.subTest(missing=missing):
                errors = self.shared_section_errors_for(
                    "olares-test", complete.replace(missing, "## Something else")
                )
                self.assertTrue(errors, f"{missing} went unnoticed")
                self.assertIn(expected, errors[0])

    def test_a_verb_index_that_does_not_say_what_to_read_is_refused(self):
        errors = self.shared_section_errors_for(
            "olares-test",
            "> **Shared front door:** load the shared skill.\n\n"
            "## Verb index\n\n| Verb | Purpose | Reference |\n|---|---|---|\n"
            "| `list` | list apps | [list](references/list.md) |\n\n"
            "## Safety and escalation\n\n- Stop when asked to.\n",
        )
        self.assertEqual(len(errors), 1, errors)
        self.assertIn("Read when triggered", errors[0])

    def test_the_skills_with_no_verb_tree_still_need_the_other_two(self):
        body = (
            "> **Shared front door:** load the shared skill.\n\n"
            "## Safety and escalation\n\n- Stop when asked to.\n"
        )
        for name in sorted(validate.NO_VERB_INDEX - {validate.SHARED_SKILL}):
            with self.subTest(skill=name):
                self.assertEqual(self.shared_section_errors_for(name, body), [])
        self.assertEqual(
            self.shared_section_errors_for(
                validate.SHARED_SKILL, "## Safety and escalation\n\n- Stop when asked to.\n"
            ),
            [],
        )

    def shared_section_errors_for(self, skill_name, body):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory).resolve()
            skill_dir = root / skill_name
            skill_dir.mkdir(parents=True)
            (skill_dir / "SKILL.md").write_text(body, encoding="utf-8")
            original_root = validate.ROOT
            validate.ROOT = root
            try:
                errors = []
                validate.validate_shared_sections(skill_dir, errors)
            finally:
                validate.ROOT = original_root
            return errors

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
                "[market charts](../olares-market/references/olares-market-chart-publish.md#upload)\n",
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


class StampTests(unittest.TestCase):
    """The stamp runs once per release, in a job nobody watches closely.

    Its failure modes are quiet ones -- a version left as the placeholder, a
    body example rewritten, a file it never opened -- so they are stated here
    rather than discovered in a published binary.
    """

    def write_suite(self, root: Path, names: list[str]) -> None:
        for name in names:
            skill_dir = root / name
            skill_dir.mkdir()
            (skill_dir / "SKILL.md").write_text(
                VALID_FRONTMATTER.replace("name: olares-test", f"name: {name}")
                + "\n## Example\n\nversion: 9.9.9-not-ours.1\n",
                encoding="utf-8",
            )

    def test_every_skill_is_stamped(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.write_suite(root, ["olares-a", "olares-b", "olares-c"])
            stamped = stamp.stamp(root, "1.13.0-cli.1")
            self.assertEqual(len(stamped), 3)
            for skill in stamped:
                self.assertEqual(
                    validate.read_frontmatter_version(skill), "1.13.0-cli.1"
                )

    def test_the_body_is_left_alone(self):
        """A reference that shows an agent a YAML example is prose, not the
        frontmatter, and rewriting it would corrupt what the agent copies."""
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.write_suite(root, ["olares-a"])
            stamp.stamp(root, "1.13.0-cli.1")
            text = (root / "olares-a" / "SKILL.md").read_text(encoding="utf-8")
            self.assertIn("version: 9.9.9-not-ours.1", text)
            self.assertEqual(text.count("1.13.0-cli.1"), 1)

    def test_a_release_that_is_not_one_is_refused(self):
        # Both halves matter: a bare x.y.z is the OS line's spelling, and an
        # empty stamp is what a workflow passes when its own version lookup
        # came back empty.
        for version in ["1.13.0", "", "latest", "v1.13.0-cli.1", "1.13.0-cli"]:
            with tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                self.write_suite(root, ["olares-a"])
                with self.assertRaises(stamp.StampError, msg=version):
                    stamp.stamp(root, version)
                self.assertEqual(
                    validate.read_frontmatter_version(root / "olares-a" / "SKILL.md"),
                    "1.12.7-cli.4",
                    f"{version!r} was refused but the file was written anyway",
                )

    def test_an_empty_tree_is_an_error_not_a_success(self):
        """Stamping nothing and reporting success is how a moved directory
        ships a suite still naming the placeholder."""
        with tempfile.TemporaryDirectory() as directory:
            with self.assertRaises(stamp.StampError):
                stamp.stamp(Path(directory), "1.13.0-cli.1")

    def test_a_skill_with_no_version_line_is_an_error(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            skill_dir = root / "olares-a"
            skill_dir.mkdir()
            (skill_dir / "SKILL.md").write_text(
                "---\nname: olares-a\n---\n\nbody\n", encoding="utf-8"
            )
            with self.assertRaises(stamp.StampError):
                stamp.stamp(root, "1.13.0-cli.1")

    def test_the_committed_suite_declares_the_placeholder(self):
        """What is in git is what a local build embeds, and it is not a release.

        A plausible-looking number here is the failure this replaced: a reader
        cannot tell whether it is authoritative, and the release that would
        have corrected it is the one that stamps over it anyway.
        """
        for skill in sorted(MODULE_PATH.parent.glob("olares-*/SKILL.md")):
            self.assertEqual(
                validate.read_frontmatter_version(skill),
                stamp.PLACEHOLDER_VERSION,
                f"{skill.parent.name} carries a hand-written version",
            )


if __name__ == "__main__":
    unittest.main()
