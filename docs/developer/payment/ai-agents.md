---
outline: [2, 3]
description: Use one prompt to have an AI agent build a store that accepts Olares Payment, install it on Olares OS, and get a public URL you can charge against.
head:
  - - meta
    - name: keywords
      content: Olares Payment, AI agent, olares-cli, Agent Skills, demo store, Olares app, chart
---

# Build a payment store with an AI agent

This article walks you through using an AI agent to build an online store that accepts stablecoins, install it on Olares OS, and get a fixed domain you can reach from anywhere.

::: warning Version requirement
This tutorial requires Olares OS 1.12.6 or later.
:::

Only two values need your attention. Note them down now and copy them as you go:

| Value | In this tutorial | Description |
|---|---|---|
| Olares ID domain prefix | `olarespayment` | The part of your Olares ID before the @ |
| Store entrance name | `demo` | Also the store's subdomain prefix. Pick your own, but keep it identical throughout. |

Following the example, the final address of the store is `https://demo.olarespayment.olares.com`.

The whole job takes three steps. Writing the code, packaging, and installing all go to the agent:

| Step | What you do | What the agent does | State when done |
|---|---|---|---|
| 1. Prepare | Create an API key in the Olares Payment merchant dashboard, register the webhook, and write the three secrets into `.env` | — | All secrets ready |
| 2. Develop and deploy | Send one prompt | Write the code, package and install the Olares app | The store is live with a fixed domain |
| 3. Verify | Open the store and place an order | — | The order turns "paid" |

Why can you register the webhook in step 1? An Olares app's domain is fixed. It is built from the entrance name and your Olares ID, and you specify the entrance name in the prompt. So `https://demo.olarespayment.olares.com/webhook` is known before you install anything.

## Before you start

- **Prepare an AI agent**: any AI coding agent, such as Codex, Claude Code, Cursor, or DeepSeek Harness.
- **Install olares-cli and Agent Skills**: see [Install olares-cli](../cli-install) and [Install and use Agent Skills](../cli-agent-skills).
- **Sign in to Olares OS with olares-cli**: see [Log in to Olares](../cli-log-in). Make sure you are signed in with the example Olares ID.

  ```bash
  olares-cli profile list
  ```

  Example output:

  ```text
      NAME                      OLARES-ID                 STATUS
  *   olarespayment@olares.com  olarespayment@olares.com  logged-in
  ```

  The leading `*` marks the current profile.

- **Docker**: this machine is logged in to Docker Hub or another public registry with `docker login`. Step 2 pushes images to it.

## Step 1: Prepare

1. Scan the QR code with LarePass to sign in to the Olares Payment merchant dashboard. Your merchant account is created automatically on first sign-in.

2. Open the **Checkouts** page, open **Advanced settings** of the default store, and create a key under **API keys**. You get `pk_live_…` and `sk_live_…`. (For the illustrated steps, see the first two steps of the [Quickstart](./quickstart).)

3. On the same panel, under **Webhooks**, set the webhook URL to `https://demo.olarespayment.olares.com/webhook` and save. You get the signing secret `whsec_…`.

   ![API key and webhook in the Olares Payment merchant dashboard](/images/payment/dashboard-api-webhook.png#bordered)

4. Create a repository and write the three secrets into `.env`:

   ```bash
   mkdir demo && cd demo
   git init
   echo ".env" >> .gitignore
   ```

   ```plain
   # .env
   PAYMENT_API_KEY=pk_live_…
   PAYMENT_API_SECRET=sk_live_…
   PAYMENT_WEBHOOK_SECRET=whsec_…
   ```

## Step 2: Develop and deploy

Start the agent in the repository directory (`codex`, `claude`, or open the folder in Cursor), then send the prompt below. Only `demo` needs changing: replace it with your entrance name, the same one you used in the webhook URL in step 1.

```plain
Build a minimal online store (Node.js): a product page, a checkout endpoint, and an order result page.
Take payments with Olares Payment:
- On checkout, call createPayment, return checkoutUrl to the frontend, and redirect to the hosted checkout page
- Expose a /webhook endpoint that receives payment.succeeded and marks the order paid once the signature is verified
- Read secrets from .env (PAYMENT_API_KEY / PAYMENT_API_SECRET / PAYMENT_WEBHOOK_SECRET);
  never hardcode them

Then package it as an Olares app and install it on my Olares OS:
- Build the image and push it to a public registry
- Generate the chart: fix the external entrance name to demo, make the entrance publicly accessible,
  and expose the payment secrets as configurable environment variables
- Upload and install it, then tell me the store's address

Payment API docs: https://www.olares.com/docs/developer/payment/llms-full.txt
```

Once the agent has read the docs, it does the rest in one pass: write the code, run it locally, then package and install it. Packaging details live in the `olares-chart` and `olares-market` skills, so you do not need to guide it. The entrance name is fixed to `demo`, so the address it produces is the one you registered the webhook with in step 1.

::: info When the address is not what you expected
If the address the agent reports after installation is not what you expected (Olares assigns another domain when the entrance name is taken), go back to the merchant dashboard and change the webhook URL to the actual domain.
:::

## Step 3: Verify

Open `https://demo.olarespayment.olares.com` and place a real order: redirect to the hosted checkout → complete the payment → return to the store, where the order shows "paid".

The order status is flipped by the webhook notification, so seeing "paid" means the whole chain works: payment, notification, and fulfillment.

When it finishes, the store page looks like this:

![Store page](/images/payment/demo-store.png#bordered)

::: info Example source
This example was generated with DeepSeek-V4.1-Flash.
:::

## Next steps

The minimal store can already take payments. From here you can keep building, for example by having the agent add an order-management backend. Once it is in good shape, [publish it to Olares Market](../develop/submit-apps) so more people can use it.
