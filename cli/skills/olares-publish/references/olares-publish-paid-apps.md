# Paid apps: price.yaml, on-chain RSA, Merchant

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md), [olares-publish-targets.md](olares-publish-targets.md), and [olares-publish-submit.md](olares-publish-submit.md) first.
> A paid app is a **public-Market** app plus pricing + license enforcement. Everything in the public-distribution path still applies; this file adds only the paid extras. Official docs: [Publish paid applications](https://docs.olares.com/developer/develop/paid-apps.html).

## What "paid" adds to a normal market app

A paid (pay-to-download) app is **almost identical** to a free market app. Only two chart-level additions, plus a one-time on-chain identity setup and the Merchant management app:

1. A `price.yaml` at the OAC root.
2. The `VERIFIABLE_CREDENTIAL` env exposed to the workload so it can read the buyer's purchase credential.

Only pay-to-download is supported. In-app purchases and subscriptions are not implemented (`products: []`).

## Prerequisites

- **Closed Beta** — paid distribution requires approved developer status; apply by email (see the official doc). Do not assume access.
- Olares ID (DID), e.g. `alice@olares.com`.
- Olares OS host running **>= 1.12.3** (paid distribution floor; separate from the chart skill's general 1.12.6 porting baseline).
- Local machine with Node.js (for `did-cli`).

## Step 1 — Register RSA keys on-chain (one-time)

The buyer's license is signed against an RSA public key bound to your Olares ID on-chain. Registration is an on-chain write and costs a small gas fee on Optimism.

1. Fund a wallet: import your LarePass mnemonic (LarePass -> Settings -> Safety -> Mnemonic phrase) into MetaMask, switch to **Optimism (OP Mainnet)**, send a small amount of ETH (>= ~0.0005 ETH recommended) to the wallet address.
2. Generate and register the key with `did-cli`:

```bash
npm install -g @beclab/olaresid

did-cli rsa generate                                   # writes ./rsa-public.pem + ./rsa-private.pem
did-cli rsa get   alice@olares.com --network mainnet   # read; expect "No RSA public key set" before first registration
did-cli owner     alice@olares.com --network mainnet   # read; confirm owner wallet == the funded wallet

export PRIVATE_KEY_OR_MNEMONIC="<your mnemonic>"       # wrap in double quotes; only the on-chain write below needs it
did-cli rsa set   alice@olares.com ./rsa-public.pem --network mainnet   # on-chain write, costs gas
unset PRIVATE_KEY_OR_MNEMONIC                          # clear the secret immediately

did-cli rsa get   alice@olares.com --network mainnet   # verify it is now set
```

> **Security — never expose secrets.** The mnemonic and `rsa-private.pem` are high-privilege credentials. Never commit them, never paste them into shared logs, and `unset` the mnemonic env var when done. **The agent must not handle the mnemonic/private key or run `did-cli ... set` (an on-chain, gas-costing write) without the user's explicit consent** — guide the user to run these themselves.

## Step 2 — Configure the OAC

### `price.yaml` at the chart root

```yaml
developer: alice123.olares.com          # developer's Olares instance address
paid:
  product_id: apps-appname-paid         # format: repoName-appName-productId
  price:
    - chain: optimism                   # optimism | eth
      token_symbol: USDC                # USDC | USDT
      receive_wallet: "0xcbbcd55960eC62F1dCFBac17C2a2341E4f0e81c8"
      product_price: 100000             # example value from the official doc; see paid-apps.html for the unit
  description:
    - lang: en
      title: Purchase item title
      description: Purchase item description
      icon: https://item.icon.url
products: []                            # in-app purchase / subscription items — not implemented yet
```

### Expose `VERIFIABLE_CREDENTIAL` for license checks

The app reads the buyer's purchase credential from an injected env var. Add it to the pod template where you enforce purchase:

```yaml
- name: VERIFIABLE_CREDENTIAL
  value: "{{ .Values.olaresEnv.VERIFIABLE_CREDENTIAL }}"
```

Declare it in `OlaresManifest.yaml` (env rules — [`../../olares-chart/references/olares-chart-env.md`](../../olares-chart/references/olares-chart-env.md)):

```yaml
envs:
  - envName: VERIFIABLE_CREDENTIAL
    required: true
    type: string
    editable: false
    applyOnChange: true
```

### Submit

Submit through the standard flow — see [olares-publish-submit.md](olares-publish-submit.md). No extra PR type for paid; it is a normal `NEW`/`UPDATE` PR whose OAC happens to contain `price.yaml`.

## Step 3 — Merchant app (checkout + license management)

Merchant is the developer-side panel that issues licenses and verifies them when buyers' apps start.

1. Install **Merchant** from Olares Market.
2. Open it, enter your developer Olares ID, click **Import Developer** — it loads your on-chain product info and RSA public key.
3. `cat ./rsa-private.pem`, paste the full content into the dialog, click **Import private key**.

> **Keep the host online 24/7.** Purchase verification needs real-time callbacks to the host running Merchant. An offline host or flaky network can block buyers from completing a purchase.

### Updating developer info

To change RSA keys, product info, pricing, or receiving wallet:

1. Regenerate keys / edit the OAC, submit an `UPDATE` PR ([olares-publish-submit.md](olares-publish-submit.md)), wait for merge.
2. Reopen Merchant; when prompted, click **Update / Re-install** to apply the latest config.

## Agent boundaries

- **Do NOT** run `did-cli rsa set` (the on-chain, gas-costing write), touch the mnemonic, or handle `rsa-private.pem` for the user — these are secret-bearing / chain-mutating actions. Guide the user to run them. (`did-cli rsa get` / `owner` are read-only and safe.)
- **Do** author `price.yaml`, wire `VERIFIABLE_CREDENTIAL` into the template + manifest, and validate the chart like any market app.
