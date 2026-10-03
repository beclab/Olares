---
description: Payment methods of the Olares Payment API — createPayment, getPayment, listPayments, BuyerRef tiers, the Payment object, and idempotency.
head:
  - - meta
    - name: keywords
      content: createPayment, getPayment, listPayments, BuyerRef, idempotency, Olares Payment
---

# Payments

Create, query, and list payments. Authentication and errors are covered in the [API overview](./index).

## createPayment

`POST https://www.olares.com/payment/api/createPayment`

```ts
createPayment(
  params: CreatePaymentRequest,
  opts?: { idempotencyKey?: string },
): Promise<CreatePaymentResult>
```

The receiving account is inferred from the key. One call returns the hosted checkout.

### CreatePaymentRequest

| Field | Type | Required | Description |
|---|---|---|---|
| `amountCents` | `number` | Yes | Price in cents; `1000` = $10.00. |
| `currency` | `'usd'` | No | Pricing currency; defaults to `usd`. |
| `returnUrl` | `string` | No | Absolute http(s) URL to redirect to after payment. Omit to stay on the success page. |
| `metadata` | `Record<string, unknown>` | No | Pass-through object echoed in queries, lists, and webhooks — use it for reconciliation (e.g. your `order_id`). |
| `buyer` | [`BuyerRef`](#buyerref) | No | Buyer disclosure. Omit for an anonymous payment. |

### BuyerRef

Three mutually exclusive tiers. Mixing fields or half-disclosing is rejected with `1100`.

| Tier | Fields | Gateway behavior |
|---|---|---|
| omitted | — | Anonymous. No identity, no account created. |
| `{ kind: 'external' }` | `ref` (1–128 chars); optional `display.name` / `display.avatarUrl` | Your own label for the buyer. Not verified; snapshotted onto the payment for dashboard reconciliation. |
| `{ kind: 'olares' }` | `olaresId` **and** `did`, both required | Resolved through the DID gate; closes the buyer's customer account. |

`ref` is **your** user identifier, not an Olares account. `display` is presentational; only its format is validated.

```ts
await client.createPayment({ amountCents: 500 });

await client.createPayment({
  amountCents: 500,
  buyer: { kind: 'external', ref: 'user_8817', display: { name: 'Ada' } },
});

await client.createPayment({
  amountCents: 500,
  buyer: { kind: 'olares', olaresId: 'alice.olares.com', did: 'did:olares:0x…' },
});
```

### CreatePaymentResult

| Field | Description |
|---|---|
| `paymentId` | Payment ID for later queries and reconciliation. |
| `checkoutUrl` | The hosted checkout for this payment (a link carrying `intent_id` and `client_secret`). Send it to the buyer. |

Payments expire (default 30 minutes, `INTENT_TTL_SECS` on the gateway). An expired checkout can no longer be paid.

### Idempotency

```ts
await client.createPayment(params, { idempotencyKey: 'order:ord_123' });
```

The gateway deduplicates on `(scope, account, key)`. Replaying the same key with the same body returns the first result; the same key with a different body fails with `400`. Derive the key from your business identity and reuse it across retries — the SDK never generates one.

---

## getPayment

`POST https://www.olares.com/payment/api/getPayment`

```ts
getPayment(paymentId: string): Promise<PaymentResult>
```

An unpaid payment is **not** an error. `paid: true` if and only if the status is `PAYMENT_STATUS_SUCCEEDED` **and** the latest attempt carries a `txHash`. A success status without a `txHash` returns `paid: false` — do not fulfill.

```ts
const result = await client.getPayment(paymentId);
if (result.paid) {
  const { txHash, payAmount, payCurrency, chain } = result.credential;
} else {
  // result.payment.status
}
```

This endpoint is dual-authenticated: HMAC for the merchant, or `client_secret` for the buyer's checkout page. A keyless SDK client can query with `getPayment(paymentId, { clientSecret })`.

On a succeeded payment the result also carries `refundSummary` — the order's refund ledger (received, refunded, still refundable). Because the same object is served to the buyer's checkout session, its `refunds` list holds **succeeded refunds only**. See [Refunds](./refunds#reading-refunds-on-a-payment).

<TryIt endpoint="getPayment" />

---

## listPayments

`POST https://www.olares.com/payment/api/listPayments`

```ts
listPayments(params?: ListPaymentsRequest): Promise<ListPaymentsResponse>
```

| Param | Description |
|---|---|
| `status` | Exact status filter, e.g. `PAYMENT_STATUS_SUCCEEDED`. |
| `metadata` | JSONB containment match (`metadata @> filter`). |
| `cursor` | `nextCursor` from the previous page. |
| `limit` | Default 20, max 200. |

List items never include `clientSecret`.

```ts
const { items, hasMore, nextCursor } = await client.listPayments({
  status: 'PAYMENT_STATUS_SUCCEEDED',
  metadata: { order_id: 'ord_123' },
  limit: 20,
});
```

---

## The Payment object

Returned by `getPayment` / `listPayments` (TS names; wire uses `snake_case`):

| Field | Description |
|---|---|
| `paymentId` | Payment ID |
| `merchantAccountId` | Receiving account |
| `buyer` | Buyer snapshot from creation; `null` when anonymous |
| `amountCents` / `currency` | Pricing |
| `status` | `PAYMENT_STATUS_*` |
| `settlementCurrency` / `settlementAmount` | On-chain settlement currency and amount |
| `metadata` | Pass-through object from creation |
| `clientSecret` | Only present when the object is carried inside a creation response; query and list results never include it |
| `latestAttempt` | Latest attempt, including `txHash` |
| `refundSummary` | [Refund ledger](./refunds#reading-refunds-on-a-payment). Non-null on `getPayment` for a succeeded payment only; `listPayments` items always carry `null` |
| `expiresAt` / `canceledAt` / `paidAt` / `createdAt` / `updatedAt` | RFC 3339 timestamps |

`PaymentResult` additionally carries `credential` when `paid: true` — `txHash`, `payAmount`, `payCurrency`, `chain`, `chainType`, `networkId` — the same shape as the `payment.succeeded` webhook credential.
