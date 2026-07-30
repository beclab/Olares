# Market-ready requirements: what public distribution demands

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first, and confirm the app already installs and reaches `running` on your Olares via the [`../../olares-chart/SKILL.md`](../../olares-chart/SKILL.md) deploy flow.
> This file sorts the requirements for a **public Market** listing by what happens when you get them wrong. It assumes the chart is already functionally complete (storage / middleware / entrances / a pullable image) from `olares-chart`.

## Sort by consequence, not by checkbox

Olares Market is **permissionless** — there is no review board deciding whether your app is good enough. Automated checks judge form; nothing judges substance. So the requirements are not one flat list, and treating them as one is why submissions feel heavier than they are:

| Tier | If you get it wrong | Can you fix it later? |
|---|---|---|
| **Hard gates** | Rejected — `lint` fails or GitBot blocks the PR | Must fix before merge |
| **Functional reach** | Merges fine, then fails to install for some users | Yes, `UPDATE` PR — but users hit it meanwhile |
| **Listing quality** | Merges and installs fine, but few people try it | Yes, `UPDATE` PR, any time |

Only the first tier blocks you. **A minimal but honest listing can go live and improve afterwards** — a path that reviewed stores like Google Play do not offer, and the main reason publishing here is lighter than it looks.

## Tier 1: hard gates

### Enforced by `olares-cli chart lint`

| Field | Rule |
|---|---|
| `metadata.name` | 1–30 chars; matches the folder name and `Chart.yaml` `name` |
| `metadata.title` | Required, 1–30 chars |
| `metadata.description` | Required |
| `metadata.icon` | Required, and must be a valid `http(s)` URL — producing one: the icon reference from the parent SKILL |
| `metadata.version` | Required, valid semver; equals `Chart.yaml` `version` |

`lint` validates the icon **URL's form, not that it resolves**. A typo'd host passes `lint` and shows a broken image in the Market.

### Enforced by GitBot at PR time

Folder name (lowercase letters and digits only, **no hyphens**, <=30 chars), PR title format, the version segment of the PR title matching `Chart.yaml` `version`, file scope, no duplicate open PR, no stray `.suspend` / `.remove`, and **valid `metadata.categories`** — categories are the one metadata field GitBot enum-checks and local `lint` does not, so a chart that lints clean still gets blocked here. Details in the submit flow from the parent SKILL.

## Tier 2: functional reach

Nothing checks these. They do not block the merge — they decide whether the app actually installs for the user who clicks Install.

| Concern | Local deploy needed | Public Market needs |
|---|---|---|
| **Image arch** | Single-arch matching this node (`olares-cli cluster node list`) | Multi-arch build (`--platform linux/amd64,linux/arm64`) |
| **`spec.supportArch`** | Optional | Must list every arch your image actually supports |
| **`spec.accelerator`** | Only if the app needs GPU on **this** node | Fully declared with quantities when the app uses GPU/NPU; the mode -> arch cross-check applies at `lint` |

`supportArch` and the image platforms must agree in **both** directions. Declaring an arch you did not build fails the install; omitting one you did build hides the app from those users.

## Tier 3: listing quality

Nothing enforces any of this. It is what makes the difference between a listing people install and one they scroll past.

| Field | Why it matters |
|---|---|
| `spec.fullDescription` | The Market body text — the only place to explain what the app is for |
| `spec.developer` / `website` / `sourceCode` / `submitter` | Credibility; a listing with no visible author reads as abandonware |
| `spec.featuredImage` / `promoteImage[]` | JPEG/PNG/WEBP 1440x900, <=8 MB each; one featured, up to 8 promote (>=2 recommended). Producing them: the listing-images reference from the parent SKILL |
| `spec.locale` | `[en]` at minimum; add more only if actually translated |
| `entrances[].icon` | Optional per-entrance icon; reuse `metadata.icon` unless entrances genuinely differ |

Since none of this is enforced, it is also the safest thing to defer. Ship, then improve.

## Pre-PR gate

- [ ] **Tier 1 clean:** `olares-cli chart lint ./<app>` passes; folder name valid; `categories` uses accepted values for **both** OS 1.11 and 1.12
- [ ] **Tier 2 honest:** images built for every arch in `spec.supportArch`, and no arch omitted
- [ ] **Tier 3 as far as you care to go** — none of it blocks you
- [ ] **`owners` file** in the chart root with the submitter's GitHub username
- [ ] **Versions:** `metadata.version` = `Chart.yaml` `version`; bump both together for updates
- [ ] **Re-lint** after every edit, and confirm the app still installs to `running` on a real Olares

Then proceed to the submit flow from the parent SKILL.

## From "runs locally" to "in the public Market"

1. Confirm the local install already reached `running` (via the [`../../olares-chart/SKILL.md`](../../olares-chart/SKILL.md) deploy flow).
2. Fix Tier 1 — this is the short list, and the only one that can reject you.
3. Rebuild multi-arch if currently single-arch, then set `spec.supportArch` to match.
4. Add as much of Tier 3 as you have material for.
5. Re-run `lint` -> `package`.
6. Open the PR. Do **not** skip re-validation: a chart that ran locally can still carry stub categories that GitBot rejects.

Functional refine (storage / middleware / entrances) should already be done from the local phase — usually no changes needed there.

> **Paid (pay-to-download)** is a public-Market app plus `price.yaml` + a `VERIFIABLE_CREDENTIAL` license check; enter it from the parent SKILL's paid-app route.
