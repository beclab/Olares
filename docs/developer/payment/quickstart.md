---
description: From LarePass sign-in to your first stablecoin payment in five steps — create API keys, create a payment, send the checkout link, and confirm fulfillment.
head:
  - - meta
    - name: keywords
      content: Olares Payment quickstart, create payment, checkout URL, API keys, HMAC, LarePass sign in
---

# Quickstart

Integrate stablecoin payments into your own service in five steps. You need an [Olares ID](/manual/get-started/create-olares-id) and the LarePass app on your phone, plus a server runtime (Node.js 18+ or Go).

::: tip Prefer code over reading?
The [demo store example](https://github.com/beclab/olares-payment-developer-example) is this quickstart as a runnable repository — clone it, add your keys, and watch a payment arrive.
:::

## Step 1. Sign in with LarePass

Open [https://www.olares.com/payment/dashboard](https://www.olares.com/payment/dashboard/) in your browser — that's the merchant dashboard — and scan the QR code with LarePass.

![Sign in to Olares Payment](/images/payment/dashboard-login.png#bordered)

On first sign-in, Olares Payment creates your merchant account and **pre-configures your receive addresses**: your LarePass wallet's EVM address is set up on every supported chain. You can start receiving immediately — no forms, no manual address entry.

![Ready to receive](/images/payment/dashboard-home.png#bordered)

## Step 2. Create an API key

In the dashboard, go to **Checkouts → Advanced settings → API keys**, name the key (for example, `Store backend`), and click **New key**.

![API keys and webhooks](/images/payment/dashboard-advanced-settings.png#bordered)

You get a `pk_live_…` / `sk_live_…` pair:

- `pk_live_…` identifies your integration.
- `sk_live_…` signs requests (HMAC). **Keep it on your server — never ship it to a browser, mobile app, or public repository.**

## Step 3. Install the SDK

<Tabs>
<template #TypeScript>

```bash
npm install @olares/payment-sdk
```

```ts
import { MerchantClient } from '@olares/payment-sdk';

const client = new MerchantClient({
  apiKey: process.env.PAYMENT_API_KEY!,     // pk_live_…
  apiSecret: process.env.PAYMENT_API_SECRET!, // sk_live_…
});
// Defaults to the production gateway https://www.olares.com/payment — no baseUrl needed.
```

</template>
<template #Go>

```bash
go get github.com/Above-Os/olares-payment/packages/payment-sdk-go
```

```go
import paymentsdk "github.com/Above-Os/olares-payment/packages/payment-sdk-go"

merchant := paymentsdk.NewMerchantClient(paymentsdk.Config{
    APIKey:    os.Getenv("PAYMENT_API_KEY"),    // pk_live_…
    APISecret: os.Getenv("PAYMENT_API_SECRET"), // sk_live_…
})
```

</template>
<template #cURL>

Every endpoint is a plain `POST /api/<method>` with HMAC headers — usable from any language:

```bash
# See "API reference → Authentication" for how to compute the signature.
curl -X POST https://www.olares.com/payment/api/ping \
  -H 'content-type: application/json' \
  -d '{}'
# {"code":0,"payload":{"server_time":"2026-09-07T08:30:00Z"}}
```

</template>
</Tabs>

## Step 4. Create a payment

`createPayment` returns a hosted checkout. Send the link to your buyer — the checkout handles coin, network, and wallet connection.

<Tabs>
<template #TypeScript>

```ts
const { paymentId, checkoutUrl } = await client.createPayment({
  amountCents: 499,              // $4.99, priced in USD
  metadata: { order_id: 'ord_001' },
  returnUrl: 'https://your-shop.example/orders/ord_001', // optional redirect after payment
});
// → checkoutUrl: https://www.olares.com/payment/?intent_id=pi_…&client_secret=…
```

</template>
<template #Go>

```go
resp, err := merchant.CreatePayment(ctx, &paymentv1.CreatePaymentReq{
    AmountCents: 499,
    Metadata:    &structpb.Struct{ /* order_id: ord_001 */ },
})
// resp.IntentId, resp.CheckoutUrl
```

</template>
<template #cURL>

```bash
curl -X POST https://www.olares.com/payment/api/createPayment \
  -H 'content-type: application/json' \
  -H "x-olares-payment-key: $PAYMENT_API_KEY" \
  -H "x-olares-payment-timestamp: $(date +%s%3N)" \
  -H "x-olares-payment-nonce: $(uuidgen)" \
  -H "x-olares-payment-signature: $SIGNATURE" \
  -d '{"amount_cents":499,"metadata":{"order_id":"ord_001"}}'
```

</template>
</Tabs>

::: details Safe retries with an idempotency key
Pass a stable key (`{ idempotencyKey: 'order:ord_001' }` / `WithIdempotencyKey(...)` / `Idempotency-Key` header). A replayed request returns the first result instead of creating a duplicate payment.
:::

Your buyer sees the hosted checkout and pays with any EVM wallet — no Olares account needed:

![Hosted checkout](/images/payment/checkout-page.png#bordered)

## Step 5. Confirm and fulfill

A payment is fulfilled when `getPayment` returns `paid: true` — status `SUCCEEDED` **and** an on-chain transaction hash. Deliver only on that signal.

<Tabs>
<template #TypeScript>

```ts
const result = await client.getPayment(paymentId);
if (result.paid) {
  const { txHash, payAmount, payCurrency, chain } = result.credential;
  // fulfill the order
}
```

</template>
<template #Go>

```go
res, err := merchant.GetPayment(ctx, intentId)
if res.Paid {
    // res.Credential.TxHash — fulfill the order
}
```

</template>
<template #cURL>

```bash
curl -X POST https://www.olares.com/payment/api/getPayment \
  -H 'content-type: application/json' \
  -H "x-olares-payment-key: $PAYMENT_API_KEY" \
  -H "x-olares-payment-timestamp: $(date +%s%3N)" \
  -H "x-olares-payment-nonce: $(uuidgen)" \
  -H "x-olares-payment-signature: $SIGNATURE" \
  -d '{"intent_id":"pi_…"}'
```

</template>
</Tabs>

Instead of polling, register a [webhook](./webhooks) and fulfill on `payment.succeeded`. Payments left unpaid expire (default 30 minutes) and converge to `CANCELED` — there is no cancel API.

## Next steps

- [API reference](./api-reference/) — all methods, parameters, and error codes
- [Refunds](./api-reference/refunds) — send a payment back along its original route
- [Webhooks](./webhooks) — signatures, retries, and local testing with a free public URL
- [Build a payment store with an AI agent](./ai-agents) — hand this documentation to an agent and get a store installed on Olares OS
- [Demo store example](https://github.com/beclab/olares-payment-developer-example) — a minimal reference integration you can clone
