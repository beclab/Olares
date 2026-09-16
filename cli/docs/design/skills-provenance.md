# Skill provenance: why the staleness notice compares content, not versions

The `olares-*` agent skills are compiled into `olares-cli` ([skills/embed.go](../../skills/embed.go)), so reading a skill cannot fetch a version that disagrees with the verbs the binary has. That closes the gap embedding was meant to close, and leaves one open: the skills are also written to disk, and upgrading `olares-cli` does not rewrite them.

This document holds the reasoning behind how that remaining gap is detected and reported. It used to sit in [cli/README.md](../../README.md), where it was about a fifth of the file and stood between a reader and the two commands they actually needed (`skills install`, `skills export`). The README now states the conclusions; the arguments are here.

## The notice, and why a version comparison would not do

Run `skills install` again after every CLI upgrade. Until you do, each command prints a line on stderr saying the installed copy came from a different build. `OLARES_CLI_NO_SKILL_NOTICE=1` silences it.

That line is decided by content, not by the version the skills declare. `skills install` and `skills export` leave a `.olares-cli-suite` file recording a hash of what they wrote, and the notice compares it against what the running binary carries ([cmd/ctl/skills/notice.go](../../cmd/ctl/skills/notice.go)).

Comparing the declared versions instead would be silent in the case that matters most. On the daily channel the version moves every few weeks while the skills move every day, so two copies a month apart carry one label. A version comparison sees two equal strings and says nothing, precisely when an agent is most likely to be reading instructions that have moved. When that happens the notice says so in as many words — same version, different copy — rather than printing one version twice.

Content also settles the reverse case, where two channels ship the same skills under different labels and a label comparison would have them accuse each other forever.

## The hash is provenance, not integrity

It says which build wrote this copy. It is not a checksum of the tree, so editing an installed skill for your own machine is not reported as drift — that is a deliberate allowance, not an oversight: the local-editing loop in the README's developer section depends on it.

Agent directories hold no marker of their own. They are links (or copies) of the store, and the store is what the notice reads.

## One store, two binaries

An Olares host that also has the npm copy installed has two `olares-cli` builds sharing `~/.agents/skills`, and they are never the same release. Whichever ran `skills install` last owns the store, and the other one says so on every command — correctly, since the skills there do document the other build.

If that is the setup you want, silence the one you drive less: `export OLARES_CLI_NO_SKILL_NOTICE=1` in the shell where it is noise, or re-run `skills install` from whichever binary your agents should be reading.

## Where a copy can come from, and what each costs

| Route | Writes | Version you get | Notice sees it |
| --- | --- | --- | --- |
| `olares-cli skills install` | `~/.agents/skills` + existing agent directories | exactly the binary's | yes |
| `olares-cli skills export <dir>` | wherever you say | exactly the binary's | no |
| `olares-cli skills read <skill>` | nothing | exactly the binary's | n/a |
| `npx skills add beclab/Olares` | the `skills` CLI decides | this repository's `main`, whose skills declare `0.0.0-cli.0` | no |

The first three are the same bytes, so they cannot disagree with the verbs the binary has. The fourth can, and silently: a skill declares `requires.bins: [olares-cli]`, which any build satisfies, so an agent reading `main`'s instructions against a six-month-old binary gets told to run flags that do not exist.

"Notice sees it" means the notice reads the marker `skills install` leaves in the store. A copy written by anything else is a copy it has nothing to compare.

Fetching from `main` has a second cost: the version in git is a placeholder. A release stamps the version it is building into the frontmatter just before compiling ([skills/stamp.py](../../skills/stamp.py)), so `0.0.0-cli.0` is what a copy taken from the repository says about itself, forever.

## ClawHub, retired

The suite was published to [ClawHub](https://clawhub.ai) before it was compiled into the binary, and for a while `cli/skills/publish.sh` kept that path open. It was removed once it could not work and was not wanted.

It could not work because a skill's version names the release it ships in (`1.12.7-cli.4`), which is numerically below the per-skill numbering the registry already holds (`olares-chart` had reached `4.18.0`), so a push is refused by a registry that requires increasing versions.

It was not wanted because a registry copy sits in the bottom half of the table above: it can disagree with the binary in front of it, silently, which is the failure embedding was adopted to end.

The script's only remaining use was its `--dry-run`, a frontmatter check that [`skills/validate.py`](../../skills/validate.py) already made more strictly — with one exception, the 1024-character `description` cap, which moved into `validate.py` with the removal.

> The listings published before the retirement are still on the registry and there is no longer anything here that can update or withdraw them. Treat what is there as a copy from before the suite moved into the binary.

## No Claude Code plugin marketplace entry

Deliberate. A marketplace entry points at a git ref, which puts it in the bottom half of the table above: a version nobody can check against the binary in front of it.

## Related

- [cli/README.md](../../README.md) — the conclusions, and the commands
- [cli/skills/README.md](../../skills/README.md) — writing and maintaining the skills themselves
