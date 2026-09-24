---
description: Account and connectivity methods of the Olares Payment API — getAccount, ping, and verifyTransaction.
head:
  - - meta
    - name: keywords
      content: getAccount, ping, verifyTransaction, Olares Payment API
---

# Account & connectivity

Key identity, account lookup, health check, and independent on-chain verification. Authentication and errors are covered in the [API overview](./index).

## getAccount

`POST https://www.olares.com/payment/api/getAccount`

```ts
const account = await client.getAccount();
// { accountId, did, olaresId, status }
```

The merchant account bound to your API key.

---

## ping

`POST https://www.olares.com/payment/api/ping`

```ts
const { serverTime } = await client.ping();
```

Public health check returning gateway time. Note the SDK still sends the signature headers, so construct the client with keys.

<TryIt endpoint="ping" />

---

## verifyTransaction

`POST https://www.olares.com/payment/api/verifyTx`

For day-to-day fulfillment use `getPayment` and webhooks. `verifyTransaction` independently reads an EVM receipt by hash:

```ts
const v = await client.verifyTransaction(txHash, 'optimism', 'mainnet');
if (v.confirmed) {
  // v.blockNumber, v.status === 'success'
}
```

Without `rpcUrl`, the call is proxied through the gateway (`POST /api/verifyTx`). With `rpcUrl`, the SDK queries that EVM RPC directly — useful when you don't want to take the gateway's word for it. EVM chains only; `chain` / `network` are needed only on the gateway path.
