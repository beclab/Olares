# Market-ready requirements: what public distribution demands

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first, and confirm the app already installs and reaches `running` on your Olares via [`../../olares-chart/references/olares-chart-deploy.md`](../../olares-chart/references/olares-chart-deploy.md).
> This file is the requirements matrix + checklist for a **public Market** listing. It assumes the chart is already functionally complete (storage / middleware / entrances / a pullable image) from `olares-chart`.

## What local deploy gives you vs. what the Market demands

Local deploy (in `olares-chart`) proves the app **runs**: `chart lint` OK, a pullable image for **this node's** arch, and an install that reaches `running`. The public Market demands more — none of it is enforced by `chart lint`:

| Concern | Already done by local deploy | Public Market demands |
|---|---|---|
| **Functional refine** (storage / middleware / entrances) | Required — done | Same, no change |
| **`chart lint` passes** | Required — done | Same, re-run after market polish |
| **Image arch** | Single-arch matching the target node (`olares-cli cluster node list`) | Multi-arch build (`--platform linux/amd64,linux/arm64`); declare matching `spec.supportArch` |
| **`metadata.title`** | Stub (`title=name`) OK | Human-readable, <=30 chars |
| **`metadata.description`** | One-line stub OK | Accurate summary for the Market listing |
| **`metadata.icon`** | Default CDN icon OK | Custom PNG/WEBP 256x256, <=512 KB |
| **`metadata.categories`** | `Utilities` stub OK (`lint` does not enum-check) | Valid categories for **both** OS 1.11 and 1.12 (GitBot `CheckWithTitle` enforces) |
| **`spec.developer` / `website` / `sourceCode` / `submitter`** | Optional | Required for a credible listing |
| **`spec.fullDescription`** | Optional | Required — longer Market body text |
| **`spec.featuredImage` / `promoteImage`** | Skip | Strongly recommended — see [promote-apps](https://docs.olares.com/developer/develop/promote-apps.html) |
| **`spec.locale`** | Skip | Recommended (`en` at minimum) |
| **`spec.supportArch`** | Optional (omit unless using accelerator modes) | Required — must match image platforms (`amd64`, `arm64`, or both) |
| **`spec.accelerator` / GPU resources** | Only if the app needs GPU on **this** node | Fully declared when the app uses GPU/NPU; mode -> arch cross-check applies at `lint` for schema >= 0.12.0 |
| **`owners` file** | Not needed | Required in the OAC root for the `beclab/apps` PR |
| **Validate** | `lint` + upload + install | Same, then the PR — [olares-publish-submit.md](olares-publish-submit.md) |

## What `lint` does NOT check (market-only)

Local `chart lint` runs `CheckConsistency` — structural and cross-field validation. It does **not**:

- Validate `metadata.categories` against the Market enum (GitBot `CheckWithTitle` does this at PR time)
- Require `featuredImage`, `promoteImage`, `fullDescription`, or `developer`
- Require multi-arch images or non-empty `spec.supportArch`

So a chart can pass `lint` (and run fine locally) with stub metadata and still fail Market submission.

## Market-ready checklist

Complete **after** the app runs locally. Use this as a pre-PR gate.

- [ ] **Metadata:** `title`, `description`, custom `icon`, dual-version `categories`
- [ ] **Spec marketing:** `developer`, `website`, `sourceCode`, `submitter`, `fullDescription`
- [ ] **Listing assets:** `featuredImage`, `promoteImage[]` (URLs reachable by the Market CDN)
- [ ] **Locale:** `spec.locale: [en]` (add more if translated)
- [ ] **Architecture:** multi-arch images pushed; `spec.supportArch` lists every supported arch
- [ ] **Resources:** if GPU/NPU needed, `spec.accelerator[]` complete with quantities; consider `--new-schema`
- [ ] **Versions:** `metadata.version` = `Chart.yaml` `version`; bump together for updates
- [ ] **`owners` file** in the chart root with the submitter's GitHub username
- [ ] **Folder name** valid for `beclab/apps` (lowercase alphanumeric, no hyphens, <=30 chars)
- [ ] **Re-lint:** `olares-cli chart lint ./<app>` after all edits
- [ ] **Runs locally** on a real Olares (upload + install -> `running`)

Then proceed to [olares-publish-submit.md](olares-publish-submit.md).

## From "runs locally" to "in the public Market"

Common path: get the app running locally first (in `olares-chart`), then polish for public listing.

1. Confirm local install already reached `running` ([../../olares-chart/references/olares-chart-deploy.md](../../olares-chart/references/olares-chart-deploy.md)).
2. Work through the market-ready checklist above.
3. Rebuild images multi-arch if currently single-arch.
4. Add `spec.supportArch` matching the image platforms.
5. Re-run `lint` -> `package`.
6. Open the PR to `beclab/apps` — do **not** skip re-validation; Market ingest may reject charts that ran locally but lack valid categories.

Functional refine (storage / middleware / entrances) should already be done from the local phase — usually no changes needed there.

> **Paid (pay-to-download)** is a public-Market app plus `price.yaml` + a `VERIFIABLE_CREDENTIAL` license check — see [olares-publish-paid-apps.md](olares-publish-paid-apps.md).
