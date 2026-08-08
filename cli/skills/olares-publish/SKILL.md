---
name: olares-publish
version: 4.7.0
description: "Publish an Olares app that already runs locally to a public Olares Market listing — release targets, market-ready metadata/architectures, producing the listing assets (256x256 icon, 1440x900 featured/promote images), the beclab/apps PR and GitBot lifecycle, and paid listings. Use for submitting, distributing, 上架, opening a beclab/apps PR, generating an app icon or Market screenshots, or selling an app; not for browsing `market list` or managing installed apps."
compatibility: Requires olares-cli on PATH; PR submission needs a GitHub account
metadata:
  openclaw:
    requires:
      bins:
        - olares-cli
---

# Publish an Olares app to the public Market

> **Shared front door:** load [`../olares-shared/SKILL.md`](../olares-shared/SKILL.md) for suite routing and platform entry points. Publishing itself does not require an Olares login; apply the shared auth gate only when returning to local upload/install validation.

> **This skill is the public-distribution step, not the authoring step.** Turning a repo / compose / Helm chart into an Olares app, refining it, and proving it installs and reaches `running` on your own Olares all live in [`../olares-chart/SKILL.md`](../olares-chart/SKILL.md). Come here **after** the app already runs locally and you want it in the public Olares Market.

Use `olares-cli chart lint --help` and [`olares-chart`](../olares-chart/SKILL.md) for chart syntax and authoring.

## When to use

- Publish / publicly list / submit / distribute / 上架 an app to the **public** Olares Market
- Open or fix a PR to [`beclab/apps`](https://github.com/beclab/apps)
- Sell an app (pay-to-download / paid listing)
- Produce the listing assets themselves — app icon, Market screenshots / 宣传图
- Keywords: publish to Market, submit to beclab/apps, app store listing, app icon, `metadata.icon`, `featuredImage` / `promoteImage`, `spec.supportArch`, multi-arch, GitBot, `owners` file, `price.yaml`, paid app

> **Prerequisite — the app must already run on your Olares.** Public submission without a working local install wastes GitBot cycles and reviewer time. Do the deploy loop in the [`olares-chart`](../olares-chart/SKILL.md) skill (its Deploy step) first.

## Mental model

Publishing is a one-way contribution to `beclab/apps`, not an `olares-cli publish` lifecycle. Start from a locally proven chart, add public-listing metadata/assets and matching multi-arch support, re-lint, then submit through GitHub and GitBot.

GitBot checks mechanical policy; it does not validate product quality, screenshot truthfulness, or security claims. The submitter owns those claims.

## Workflow index

| Current intent / state | Read |
|---|---|
| Decide what blocks submission, what can be deferred, metadata depth, architectures | [release targets](references/olares-publish-targets.md) |
| Produce the required 256×256 icon from upstream material | [icon workflow](references/olares-publish-icon.md) |
| Capture/source 1440×900 Market images | [listing images](references/olares-publish-listing-images.md) |
| Compose listing copy and image layout | [listing layout](references/olares-publish-listing-layout.md) |
| Prepare folder, owners, PR title, and respond to GitBot state | [submission workflow](references/olares-publish-submit.md) |
| Add pay-to-download | [paid apps](references/olares-publish-paid-apps.md) |

## Release decisions

- The icon is on the critical path because `metadata.icon` is lint-required. Hero/screenshots are optional and may follow in an `UPDATE`.
- Build images for the architectures declared by `spec.supportArch`; a local single-node success does not prove public multi-arch availability.
- Assets must come from the upstream project, its site, or the user's running instance. Do not invent logos, screenshots, testimonials, or capabilities.
- Public chart credentials must come from install-time inputs, middleware, or chart generation; never commit live secrets.
- Paid distribution adds `price.yaml` and application-side `VERIFIABLE_CREDENTIAL` enforcement; it does not replace the normal public-listing requirements.

## Validation and asynchronous semantics

- Re-run chart lint after every manifest or asset-path change.
- Local install and `running` validation belong to Chart/Market and can be asynchronous; do not infer that a successful upload ACK proves installability.
- GitBot processing is asynchronous. Interpret its current label/check state using the submission reference; do not repeatedly open duplicate PRs.

## Agent boundaries

- **Do NOT** fork, push, or open PRs on the developer's behalf without explicit consent — these write to their GitHub account.
- **The fork here is legitimate and required.** Submitting to the public Market is the standard open-source contribution flow: fork [`beclab/apps`](https://github.com/beclab/apps), push a branch, open a PR. This is the *public app catalog*, a different repo from the beclab dev repos the workspace "no-fork, push to `origin`" rule covers — that rule does not apply to `beclab/apps` submissions.
- **Do NOT** run `did-cli rsa set` (an on-chain, gas-costing write), touch the wallet mnemonic, or handle `rsa-private.pem` for paid apps — guide the user to run those themselves.
- **Do** verify the chart against the market-ready checklist, author `price.yaml`, wire the manifest, and interpret GitBot labels.
- Stop on unclear upstream licensing, asset provenance, architecture support, owner identity, pricing, wallet actions, or permission to use GitHub.
- Upload/install verbs for local validation belong to [`olares-market`](../olares-market/SKILL.md).
