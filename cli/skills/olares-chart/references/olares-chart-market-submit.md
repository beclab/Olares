# Market submit: PR to beclab/apps

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) and [olares-chart-publish-targets.md](olares-chart-publish-targets.md) first.
> This is the **Publish-market** capability — the final step for release target **market-distribute**. Official docs: [Submit applications](https://docs.olares.com/developer/develop/submit-apps.html), [Distribute index](https://docs.olares.com/developer/develop/distribute-index.html).

## Before you start

**Local validation must pass first.** Upload + install on the developer's Olares and confirm `running` — see [olares-chart-publish-verify.md](olares-chart-publish-verify.md). Market submission without a working install wastes GitBot cycles and reviewer time.

**Market-ready checklist must be complete** — see [olares-chart-publish-targets.md](olares-chart-publish-targets.md#market-distribute-market-ready-checklist). In particular:

- Dual-version `metadata.categories` (GitBot rejects invalid values; local `lint` does not enum-check)
- Multi-arch images + matching `spec.supportArch`
- Full metadata and listing images

## Agent boundaries

- **Do NOT** fork, push, or open PRs on the developer's behalf without explicit consent
- **Do** verify the chart against the market-ready checklist
- **Do** guide the developer through fork → add OAC → draft PR → ready for review
- **Do** help interpret GitBot labels and fix chart issues when PR is `waiting to submit`

## Step 1: Prepare the OAC

1. Ensure `olares-cli chart lint ./<app>` passes after market polish
2. Package (optional for PR — the folder is what gets committed, not the `.tgz`):
   ```bash
   olares-cli chart package ./<app>
   ```
3. Verify folder name constraints (used in PR title and GitBot validation):
   - Lowercase letters and digits only
   - **No hyphens** (`-`)
   - ≤ 30 characters
   - Must match `Chart.yaml` `name`, `metadata.name`, and PR title folder segment

4. Add an **`owners` file** (no extension) in the chart root:
   ```yaml
   owners:
   - <github-username>
   - <collaborator-username>   # optional
   ```
   Listed owners can independently submit changes for this app. See [distribute-index](https://docs.olares.com/developer/develop/distribute-index.html).

5. Confirm the OAC root contains **no control files** (`.suspend`, `.remove`) — those are only for SUSPEND/REMOVE lifecycle PRs.

## Step 2: Fork and add the chart

1. Fork [beclab/apps](https://github.com/beclab/apps)
2. Add the complete OAC folder under the fork root (same layout as local chart dir)
3. Push to a branch on the fork

For team collaboration: add maintainers to `owners` and grant push access to the fork so teammates can push to the PR branch.

## Step 3: Open a draft PR

Target: `beclab/apps:main`

### PR title format (strict — GitBot rejects otherwise)

```text
[PR type][Chart folder name][Version] Title content
```

| Field | Values |
|---|---|
| PR type | `NEW` (first submission), `UPDATE`, `SUSPEND`, `REMOVE` |
| Chart folder name | OAC directory name — must match naming convention |
| Version | Must equal `Chart.yaml` `version` **and** `metadata.version` in `OlaresManifest.yaml` |
| Title content | Brief summary |

Example: `[NEW][myapp][0.0.1] Add My App`

### GitBot rules for NEW

- **File scope:** PR only adds/modifies content under the chart folder declared in the title
- **No duplicate PR:** no other Open/Draft PR for the same chart folder
- **Clean structure:** folder name must not already exist in `beclab/apps:main`; no `.suspend` or `.remove` in the OAC root
- **Valid categories:** `metadata.categories` must use accepted Market values (both 1.11 and 1.12 where applicable)

Start as **Draft** — push fixes while GitBot re-checks. Click **Ready for review** when complete.

## Step 4: Track PR status

### Type labels

`NEW`, `UPDATE`, `REMOVE`, or `SUSPEND` — confirms PR type in title is recognized.

**Do not change PR type after labeled.** Close and open a new PR if the type was wrong.

### Status labels

| Label | Meaning | Action |
|---|---|---|
| `waiting to submit` | Issues found | Push fixes; GitBot re-checks |
| `waiting to merge` | All checks passed, queued for auto-merge | **Do not push new commits** |
| `merged` | In `beclab/apps:main` | App indexes into public Market shortly |
| `closed` | Invalid or unrecoverable | Fix issues, submit a **new** PR (don't reopen) |

After merge, the app appears in Olares Market (`market.olares` source) after a short indexing delay.

## Lifecycle after first publish

All post-publish actions are PRs to `beclab/apps:main`. See [Manage the app lifecycle](https://docs.olares.com/developer/develop/manage-apps.html).

| Action | PR type | Key rules |
|---|---|---|
| New version / config change | `UPDATE` | Version **must bump**; no `.suspend`/`.remove` in OAC root |
| Pause listing (existing installs keep working) | `SUSPEND` | Version bump; add empty `.suspend` file; no `.remove` |
| Permanent removal | `REMOVE` | Version **same** as current; `.remove` is the only file left in OAC root |

**No rollbacks** — fix forward with a new `UPDATE` version.

Before any lifecycle PR: sync fork and rebase onto latest `main` to reduce conflicts.

## Optimize listing (optional, post-merge)

Improve Market presentation with [promote-apps](https://docs.olares.com/developer/develop/promote-apps.html):

- `spec.promoteImage[]` — screenshot carousel
- `spec.featuredImage` — hero image on app detail page

Submit these via an `UPDATE` PR with a version bump.

## Troubleshooting GitBot rejection

| Symptom | Likely cause | Fix |
|---|---|---|
| Invalid categories | stub `Utilities` only, or wrong enum value | Add both 1.11 + 1.12 category values per [manifest docs](https://docs.olares.com/developer/develop/package/manifest.html#categories) |
| Name/version mismatch in title | PR title folder or version ≠ chart | Align PR title with folder name and `Chart.yaml` / `metadata.version` |
| File scope violation | Changed files outside chart folder | Restrict PR to OAC folder only |
| Duplicate PR | Another open PR for same app | Close duplicate or wait for the other to merge |
| Reserved folder name | name matches Olares reserved keyword | Rename app id (folder + Chart.yaml + metadata.name) |

For chart content issues surfaced during ingest (not GitBot title/scope), fix the OAC locally, re-run `lint`, push to the PR branch.

## Relationship to Publish-local

| Step | Publish-local | Publish-market |
|---|---|---|
| Prove app runs | ✅ upload + install on developer's Olares | ✅ prerequisite |
| Public catalog listing | ❌ (upload source is private to that Olares) | ✅ PR merged → `market.olares` |
| Metadata bar | stub OK | full market-ready checklist |

Many developers stop at Publish-local. Publish-market is the optional upgrade when they want the app in the public index.
