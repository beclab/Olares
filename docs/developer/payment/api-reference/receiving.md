---
description: Receiving configuration of the Olares Payment API — channels, receive wallets, on-chain transactions, payment method configs, and supported chains.
head:
  - - meta
    - name: keywords
      content: listChannels, listReceiveWalletTransactions, upsertOnchainPmc, listSupportedChains, receive wallet
---

# Receiving

Your receive addresses are pre-configured at first sign-in — these methods read them back, list incoming on-chain transactions, or change the configuration. Authentication and errors are covered in the [API overview](./index).

## listChannels

`POST https://www.olares.com/payment/api/listChannels`

```ts
const { channels } = await client.listChannels();
```

Receiving channels for your account. On-chain channels carry their receive wallets.

---

## listReceiveWalletTransactions

`POST https://www.olares.com/payment/api/listReceiveWalletTransactions`

```ts
const { items, hasMore, nextCursor } = await client.listReceiveWalletTransactions({
  address: '0x…', // optional; must belong to your account's receive wallets
  limit: 50,
});
```

On-chain inflows to your receive wallets, ordered by block number descending. Payment-orders history lives in [listPayments](./payments#listpayments) instead — this method is the wallet-level ledger, including direct transfers that never went through a checkout.

---

## listPaymentMethodConfigs

`POST https://www.olares.com/payment/api/listPaymentMethodConfigs`

Your account's current payment method configurations (`onchain` channel with its per-chain wallets and token allowlists).

---

## upsertOnchainPmc

`POST https://www.olares.com/payment/api/upsertOnchainPmc`

Overwrite the on-chain receiving configuration: per chain, the receive wallet and the token allowlist. An empty `tokens` list disables receiving on that chain.

```ts
await client.upsertOnchainPmc({
  chains: [
    { chainId: '10', receiveWallet: '0x…', tokens: ['USDC', 'USDT'] },
  ],
});
```

Receive addresses are pre-configured at first sign-in — use this only to change them.

---

## listSupportedChains

`POST https://www.olares.com/payment/api/listSupportedChains`

The official chain and token catalog: Optimism, Base, BNB Smart Chain, Ethereum, Arbitrum One, Polygon, Avalanche — with USDC/USDT per chain. Read-only.
