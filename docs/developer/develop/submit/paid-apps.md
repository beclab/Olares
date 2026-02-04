---
outline: [2, 3] 
description: Learn how to publish paid apps in Olares Market and manage sales using the Merchant app.
---
# Publish paid applications

Starting from Olares v1.12.3, Olares Market supports paid application distribution. To sell apps, developers must register their Olares ID (DID) on the blockchain and install the Merchant app to manage licenses, orders, and payout records.

:::info Closed Beta 
This feature is currently in Closed Beta. To publish paid apps, please contact us at [hi@olares.com](mailto:hi@olares.com) to apply for access.

Currently, only paid applications (pay-to-download) are supported. In-app purchases and subscriptions are under development. 
:::

This guide explains how to publish paid applications and configure Merchant.

## Prerequisites

Before starting, ensure that you have:
- Olares ID: A valid Olares ID (e.g., `alice123@olares.com`).
- Olares OS: A working Olares host running v1.12.3 or later.
- Developer access: Approved developer status (apply via email).
- Environment: A local computer with Node.js installed.

## Set up a wallet and fund gas fees

To enable payments, you must register your Olares ID (DID) on the blockchain. This is an on-chain interaction and requires a small Gas Fee.

### Set up a crypto wallet

This guide uses MetaMask as an example. You may use other wallet apps.

1. Open LarePass on your phone.
2. Go to **Settings** > **Safety** > **Mnemonic phrase** to view your mnemonic phrase.
    :::warning Security alert
    Your mnemonic phrase is a high-privilege credential for your account. Never share it or upload it to a public code repository. Perform the following steps in a secure local environment.
    :::
3. Install the MetaMask extension in your browser.
4. In MetaMask, select **Add wallet** > **Import a wallet**, and enter your Olares mnemonic phrase.

### Fund your wallet

Registration consumes a small amount of ETH on the Optimism network as gas fees.

1. In MetaMask, switch to **Optimism (OP Mainnet)**.
2. Copy your wallet address. It should start with `0x`.
    ![Copy MetaMask wallet address](/images/developer/develop/distribute/metamask-wallet-add.png#bordered){width=50%}
3. Send a small amount of ETH to this address. At least 0.0005 ETH is recommended for subsequent gas fees.

## Generate and register RSA keys

To ensure transaction security and license uniqueness, generate an RSA key pair and register the public key on-chain under your Olares ID.

### Generate an RSA key pair

1. Install the `did-cli` tool:
    ```bash
    npm install -g @beclab/olaresid
    ```
2. Generate a default 2048-bit RSA key pair:
    ```bash
    did-cli rsa generate
    ```
    When the command succeeds, you should see output similar to the following:
    ![RSA key pair generated](/images/developer/develop/distribute/metamask-rsa-key-pairs.png#bordered){width=60%}
3. View the generated keys:
    ```bash
    cat ./rsa-public.pem
    cat ./rsa-private.pem
    ```
    :::info Security alert
    Keep `rsa-private.pem` safe. You will need it later when configuring the Merchant app.
    :::
### Register the RSA public key

Bind the generated public key to your Olares ID. It requires signing with your mnemonic phrase and will consume gas.

1. Export your mnemonic as a temporary environment variable and wrap the value in double quotes:
    ```bash
    export PRIVATE_KEY_OR_MNEMONIC="xx xxx xx ..."
    echo $PRIVATE_KEY_OR_MNEMONIC
    ```
2. Check whether your Olares ID already has an RSA public key:
    ```bash
    did-cli rsa get alice123@olares.com --network mainnet
    ```
    If you see the following message, you can continue:
    
    ```text
    ❌ No RSA public key set for this domain
    ```
3. Verify the owner wallet address and make sure it matches the wallet you funded previously:
    
    ```bash
    did-cli owner alice123@olares.com --network mainnet
    ```

    Example output:

    ```text
    👤 Owner: 0x3.....
    ```
4. Register the RSA public key on-chain:
    
    ```bash
    did-cli rsa set alice123@olares.com ./rsa-public.pem --network mainnet
    ```
5. Verify registration:
    ```bash
    did-cli rsa get alice123@olares.com --network mainnet
    ```
6. Clear the mnemonic phrase from your environment:
    ```bash
    unset PRIVATE_KEY_OR_MNEMONIC
    echo $PRIVATE_KEY_OR_MNEMONIC
    ```
    If nothing is printed, it is cleared.

## Configure your app (OAC)

A paid app is almost the same as a free app. You only need two additional items:
- Add a `price.yaml` file at the chart root (OAC root).
- Expose `VERIFIABLE_CREDENTIAL` to the app so it can read the user's purchase credential.

### Add pricing configuration
Create `price.yaml` at the chart root:
```yaml
# Developer's Olares instance address
developer: alice123.olares.com

# Paid app
paid:
  # product_id format: repoName-appName-productId
  product_id: apps-appname-paid
  price:
    - chain: optimism | eth
      token_symbol: USDC | USDT
      receive_wallet: "0xcbbcd55960eC62F1dCFBac17C2a2341E4f0e81c8"
      product_price: 100000
  description:
    - lang: en
      title: Purchase item title
      description: Purchase item description
      icon: https://item.icon.url

# In-app purchase or subscription items (not implemented yet)
products: []
```
### Enable license checks
Your app can read the user's purchase credential via an injected environment variable.

Add the env var in the pod template where you need purchase enforcement:
```yaml
- name: VERIFIABLE_CREDENTIAL
  value: "{{ .Values.olaresEnv.VERIFIABLE_CREDENTIAL }}"
```

Declare the variable in `OlaresManifest.yaml`:
```yaml
envs:
  - envName: VERIFIABLE_CREDENTIAL
    required: true
    type: string
    editable: false
    applyOnChange: true
``` 
For detailed environment variable behavior, see [environment variables in `OlaresManifest.yaml`](/developer/develop/package/manifest.md#envs).

### Submit your app

Follow the [standard submission flow](/developer/develop/submit/submit-apps.md) to create a Pull Request.

## Install and set up Merchant

The Merchant app is your checkout and management panel. It allows you to:
- View developer info (product list, RSA keys, payout wallet, orders).
- Issue product licenses.
- Verify licenses when apps start.

:::info Keep Merchant online
Purchase requests require real-time callbacks to your Olares host for verification. Keep the host running Merchant online 24/7 with a stable, fast network connection. Host offline or unstable network may directly prevent users from completing purchases.
:::

### Install Merchant

1. Open Olares Market and search for "Merchant".
2. Click **Get**, then **Install**.

### Initial setup

After installation completes, open Merchant from Launchpad.
1. On the login screen, enter your developer Olares ID and click **Import Developer**.
2. Merchant automatically loads your on-chain product info and RSA public key.
3. Open your private key file:
    ```bash
    cat ./rsa-private.pem
    ```
    Copy the full content, paste it into the Merchant dialog, then click **Import private key**.

## Manage sales in Merchant

After setup, Merchant opens the Home dashboard, where you can monitor identity status, key configuration, wallet assets, and transaction history.

:::info 
The Store and Buy App pages in the left navigation are currently for debugging/testing only. You can ignore them in the current publishing workflow.
:::

### Sync status

The status indicator above the transaction list supports manual refresh:
- `Synced`: Up to date.
- `Syncing`: Pulling data from chain.
- `Not synced`: Not synchronized.
- `Error`: Sync failed. Check network connectivity.

### Transaction fields

| Field | Description|
|--|--|
| Wallet | Developer wallet address managed by Merchant|
| Type | Transaction type:<br>`Incoming`: revenue as receiver<br>`Outgoing`: payment sent as sender|
| Tx Hash | The hash of the transaction record, a unique identifier for the on-chain<br> transaction.|
| From/To | Counterparty address, shown based on the transaction type:<br>`Incoming`: sender address<br>`Outgoing`: recipient address |
| Contract | Token contract address, indicating the asset type and source involved in <br>the transaction.<br>**Native**: native ETH transfer with no contract interaction <br>**[Contract address]**: ERC-20 token smart contract address |
| Amount | Transaction amount, denominated in the token specified by the Contract.|
| Time | The timestamp when the transaction is confirmed and recorded on-chain.|
| Product | **Product ID**: a transaction related to an Olares app purchase<br>**-**: a regular transfer not associated with app sales |

### Update developer information

To change RSA keys, product info, pricing, or receiving wallet:
1. [Regenerate RSA keys](#generate-and-register-rsa-keys) or [update your app configurations](#configure-your-app-oac), then submit a PR with updated information.
2. Wait for the PR to be reviewed and merged.
3. After the PR is merged, open Merchant from Launchpad. When prompted, click **Update / Re-install** to apply the latest configuration. Changes take effect immediately.