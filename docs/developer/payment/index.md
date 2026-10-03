---
description: Accept stablecoin payments with Olares Payment — sign in with LarePass, get your API keys, and integrate one API. Funds settle on-chain, directly to your own wallet.
head:
  - - meta
    - name: keywords
      content: Olares Payment, stablecoin payments, USDC, USDT, payment API, merchant dashboard, LarePass, webhooks
---

# Olares Payment

Olares Payment lets you integrate stablecoin payments into your own service — accept USDC and USDT from any EVM wallet, with funds settling **on-chain, directly to your own wallet**. The gateway never holds your money.

::: tip Want to run before you read?
Clone the [demo store example](https://github.com/beclab/olares-payment-developer-example) — a complete, minimal integration you can run in minutes. Come back here to understand what each step does.
:::

## One story

![From sign-in to your first payment: LarePass sign-in, receive addresses ready, API keys, integrate the API, buyer pays with any wallet](/images/payment/payment-flow-en.svg#bordered)

- **Merchant side = LarePass.** Scan a QR code once. Your merchant account is created automatically, and your LarePass wallet's EVM address is pre-configured as the receive address on every supported chain.
- **Buyer side = any wallet.** Buyers pay at a hosted checkout with MetaMask, OKX, or any other EIP-1193 browser wallet. No Olares account needed.

## Why merchants choose it

| Property | What it means |
|---|---|
| **Self-custody settlement** | Buyers transfer straight to your receive address on-chain. The gateway verifies the transfer; it never takes custody of funds. |
| **Zero-setup onboarding** | First LarePass sign-in creates your account and configures all 7 chains for you — no forms, no manual address entry. |
| **One API** | Create a payment, get a hosted checkout page, confirm via webhook or polling. That's the whole integration. |
| **Global stablecoins** | USDC and USDT on Optimism, Base, BNB Smart Chain, Ethereum, Arbitrum One, Polygon, and Avalanche. |

## Core concepts

| Concept | Description |
|---|---|
| **Payment** | A single payable order, identified by a `paymentId` (wire: `intent_id`). Created via API, paid at the hosted checkout. |
| **Checkout** | The hosted payment page returned at creation (the `checkoutUrl` field). Send the link to your buyer — it handles coin, network, and wallet connection. |
| **Status** | A payment moves `REQUIRES_PAYMENT_METHOD → PROCESSING → SUCCEEDED`, or ends `CANCELED` when it expires (default TTL 30 minutes). A failed on-chain attempt emits `payment.failed` and the payment falls back to awaiting a new attempt. |
| **Refund** | Money sent back along the route the payment arrived on — out of your receive wallet, into the address that paid. Partial or full; the gateway verifies the transfer on-chain, and settlement stays self-custodial as it does for payments. |
| **Refund execution link** | What `createRefund` hands back instead of calldata: a short-lived, single-refund link an operator opens to sign the transfer with the original receive wallet. |
| **API keys** | A `pk_live_…` / `sk_live_…` pair created in the dashboard. The secret signs requests (HMAC) and must stay on your server. |
| **Webhook** | An HTTPS endpoint you register in the dashboard. The gateway POSTs signed events such as `payment.succeeded` to it. |
| **Webhook secret** | A `whsec_…` used to verify webhook signatures. Independent from your API keys. |

## Where to go next

<div class="cta">
  <a href="./quickstart">
    <h3>Quickstart →</h3>
    <p>From LarePass sign-in to your first payment in five steps.</p>
  </a>
</div>

- [API reference](./api-reference/) — every method, parameter, and error code
- [Refunds](./api-reference/refunds) — refund a payment back along its original route
- [Webhooks](./webhooks) — signed events, retries, and verification
- [Build a payment store with an AI agent](./ai-agents) — one prompt that builds a store and installs it on Olares OS
- [Demo store example](https://github.com/beclab/olares-payment-developer-example) — a minimal reference integration you can clone
- [User manual](/manual/payment/) — the product introduction, no engineering background needed
