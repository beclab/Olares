---
description: Olares Payment webhooks — register an endpoint, verify signatures, handle payment.succeeded, and test deliveries from the dashboard.
head:
  - - meta
    - name: keywords
      content: Olares Payment webhooks, payment.succeeded, webhook signature, constructEvent, retry, whsec
---

# Webhooks

Webhooks push payment lifecycle events to your server, so you can fulfill orders without polling. Deliveries are signed; verify every one before trusting it.

## How it works

1. You register an HTTPS endpoint in the dashboard (**Checkouts → Advanced settings → Webhooks**).
2. When a payment changes state, the gateway POSTs a signed JSON event to your endpoint.
3. Your server verifies the signature and answers `2xx`. Anything else triggers retries.

Endpoints created in the dashboard subscribe to `payment.succeeded`, `refund.succeeded`, and `refund.failed` by default.

![Registering an endpoint](/images/payment/dashboard-advanced-settings.png#bordered)

## Events

| Event | When | Extra fields |
|---|---|---|
| `payment.succeeded` | Payment confirmed on-chain | `credential` (tx hash, amount, currency, chain), `paid_at` |
| `payment.failed` | The buyer's submitted transaction failed on-chain; the payment stays open for another attempt | `tx_hash`, `fail_reason` |
| `payment.canceled` | The payment expired unpaid | `cancellation_reason` (`expired`) |
| `refund.succeeded` | A refund transfer passed on-chain verification | `refund_id`, `amount`, `token_symbol`, `chain`, `network_id`, `tx_hash`, `succeeded_at` |
| `refund.failed` | A refund transaction reverted or did not match the frozen route — retryable, not final | the same, plus `fail_reason` and `failed_at` |
| `endpoint.test` | Sent by the dashboard **Test** button so you can validate your receiver — answer 200 and ignore it | Sample payload only |

Payment events carry `event_type`, `intent_id`, `merchant_account_id`, `buyer`, and your `metadata`. `buyer` is a snapshot of the buyer identity provided at creation — it is `null` when none was given (anonymous order). Refund events have [their own payload shape](#refund-events).

```json
{
  "event_type": "payment.succeeded",
  "intent_id": "pi_1a07b0148515b9a95bc",
  "merchant_account_id": "acct_x8y9…",
  "buyer": {
    "kind": "olares",
    "olares_id": "buyer.olares.cn",
    "did": "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  },
  "metadata": { "order_id": "ord_001" },
  "credential": {
    "tx_hash": "0x9f3c…",
    "pay_amount": "4990000",
    "pay_currency": "USDC",
    "chain": "optimism",
    "chain_type": "CHAIN_TYPE_EVM",
    "network_id": "10"
  },
  "paid_at": "2026-09-08T15:20:00Z"
}
```

Fulfill on `credential.tx_hash` presence — the same rule as `getPayment`'s `paid` verdict.

## Refund events

`refund.succeeded` and `refund.failed` do **not** use the payload above. They carry a refund-shaped body, so read `event_type` first and only then decide what to parse:

```json
{
  "event_type": "refund.succeeded",
  "refund_id": "re_7f2c9a1b34",
  "intent_id": "pi_1a07b0148515b9a95bc",
  "merchant_account_id": "acct_x8y9…",
  "amount": "600000",
  "token_symbol": "USDC",
  "chain": "optimism",
  "network_id": 10,
  "tx_hash": "0x4d1e…",
  "reason": "customer downgraded the plan",
  "buyer": {
    "kind": "olares",
    "olares_id": "buyer.olares.cn",
    "did": "did:key:z6MkpTHR8VNsBxYAAWHut2Geadd9jSwuBV8xRoAnwWsdvktH"
  },
  "succeeded_at": "2026-09-08T15:20:00Z"
}
```

Three things to get right:

| Rule | Why |
|---|---|
| **Branch on `event_type` before parsing** | Decoding a refund event as a payment callback silently drops `refund_id`, `amount`, and `token_symbol` — the fields you need to post the credit. |
| **Refund events carry no `metadata`** | The refund callback protocol has no such field. Correlate to your own order through `intent_id` (TS: `paymentId`). |
| **`refund.failed` is not terminal** | The amount stays reserved and the operator can re-sign a corrected transaction from the same execution link. Do not release the reservation or reopen the order on it. |

`amount` is in the original token's minor units, not cents — the same unit `createRefund` takes. `reason` is your own internal note echoed back; it was never shown to the buyer.

See [Refunds](./api-reference/refunds) for the full lifecycle.

## Verify signatures

Each delivery has two headers:

- `x-olares-payment-webhook-timestamp` — unix time in milliseconds
- `x-olares-payment-webhook-signature` — `hex(HMAC-SHA256(whsec_…, "{timestamp}\n{rawBody}"))`

The signing secret (`whsec_…`) is issued at registration and can be revealed in the dashboard at any time. It is independent of your API keys.

Verification rules:

- The signed payload is `{timestamp}\n{rawBody}` — **the exact raw request bytes**. Parse JSON only after verification.
- Reject timestamps more than 5 minutes old (replay protection).
- In Express, use `express.raw` for the webhook route — a prior `express.json()` would destroy the raw body.

```ts
import { webhooks } from '@olares/payment-sdk';

app.post('/webhook', express.raw({ type: 'application/json' }), (req, res) => {
  try {
    const event = webhooks.constructEvent(req, process.env.PAYMENT_WEBHOOK_SECRET!);
    switch (event.type) {
      case 'payment.succeeded':
        // event.paymentId, event.credential.txHash, event.metadata
        break;
      case 'payment.failed':
        break;
      case 'payment.canceled':
        break;
      case 'refund.succeeded':
        // event.refundId, event.paymentId, event.amount, event.txHash — no metadata here
        break;
      case 'refund.failed':
        // event.failReason; the refund can still be retried, so do not undo anything yet
        break;
    }
    res.status(200).send('ok');
  } catch {
    res.status(400).send('bad signature');
  }
});
```

You can also pass `{ body, headers }` when your framework already buffered the body. A failed verification throws `PaymentError` — never treat the delivery as valid.

## Retries

Your endpoint must answer `2xx`. Anything else — non-2xx, timeout (60s), unreachable — counts as a failed attempt and is retried on this schedule:

| Attempt | 1 | +1 | +2 | +3 | +4 | +5 | +6 | +7 |
|---|---|---|---|---|---|---|---|---|
| Delay after failure | — | 1 min | 5 min | 30 min | 2 h | 12 h | 24 h | 24 h |

After 8 total attempts the delivery is dead-lettered (`failed`). Handle events **idempotently** — retries mean the same event can arrive more than once.

## Test and debug from the dashboard

- **Test** — sends a real signed `endpoint.test` delivery to your endpoint. The fastest way to validate your receiver. Your handler should answer 200 to it (unknown event types are safe to ignore).
- **Delivery log** — every delivery with its status, attempts, and last error.
- **Replay** — resend a failed delivery once your endpoint is fixed.
