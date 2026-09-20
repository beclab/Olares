---
outline: [2, 3]
description: Use one prompt to have an AI agent build a store that accepts Olares Payment, install it on Olares OS, and get a public URL you can charge against.
head:
  - - meta
    - name: keywords
      content: Olares Payment, AI agent, olares-cli, Agent Skills, demo store, Olares app, chart
---

# Build a payment store with an AI agent

This guide shows how to use an AI agent to build a minimal store that accepts Olares Payment and install it on Olares OS. After installation, the store gets a public domain and can take orders from anywhere.

The walkthrough below uses `olarespayment@olares.com`. Replace it with your own Olares ID.

:::warning
This tutorial requires Olares OS 1.12.5 or later.
:::

## Prepare

- **Prepare an AI agent**: any AI coding agent, such as Codex, Claude Code, Cursor, or DeepSeek Harness.
- **Install olares-cli and Agent Skills**: see [Install olares-cli](../cli-install) and [Install and use Agent Skills](../cli-agent-skills).
- **Sign in to Olares OS with olares-cli**: see [Log in to Olares](../cli-log-in). Make sure you are signed in with the example Olares ID, as shown here.

  ![olares-cli profile list signed-in status](/images/payment/cli-profile-list.png#bordered)

- **Push images to a remote registry**: this machine is signed in to Docker Hub or another registry, so you can push images for Olares nodes to pull.
- **Create an API key in the Olares Payment dashboard**: copy `pk_live_…` (API key) and `sk_live_…` (API secret). See [Quickstart](./quickstart). Set the webhook URL to `https://demo.olarespayment.olares.com/webhook`. After you save, you get `whsec_…`.

  ![API key and webhook in the Olares Payment dashboard](/images/payment/dashboard-api-webhook.png#bordered)

  In that URL, `olarespayment.olares.com` comes from the Olares ID `olarespayment@olares.com`, with `@` replaced by `.`. Replace it with your own Olares ID.

- **Prepare a working directory**: create an empty directory. Write the API key, API secret, and webhook secret into `.env`:

```plain
# .env
PAYMENT_API_KEY=pk_live_…
PAYMENT_API_SECRET=sk_live_…
PAYMENT_WEBHOOK_SECRET=whsec_…
```

## Build and install on Olares OS

Send the payment API document and the prompt below to the agent. When it finishes, the store is installed on your Olares OS. Open it to place an order and get paid.

<PaymentLlmsLink />

```plain
Build a minimal store (Node.js) that can take payments. Price the
product at $0.01. Read credentials from .env. Integrate using the
Olares Payment API docs I provide.

Package it and install it on the Olares OS I have already signed
in to with olares-cli.
Use https://demo.olarespayment.olares.com as the store URL.
Opening it should let me place an order and complete a payment.
```

:::tip
The store URL in the prompt is an example. Replace `olarespayment.olares.com` with the address derived from your Olares ID.
:::

After it finishes, the store looks like this:

![Demo store](/images/payment/ds-v4-flash-store.png#bordered)

:::tip
This demo was generated with DeepSeek-V4.1-Flash.
:::

## Next steps

The minimal store can already take payments. From here you can keep building, for example by having the agent add an order-management backend. Once it is in good shape, [publish it to Olares Market](../develop/submit-apps) so more people can use it.
