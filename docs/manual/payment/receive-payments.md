---
description: A video and screenshot walkthrough of getting paid with Olares Payment — sign in with LarePass, receive addresses ready out of the box, share your payment link, and watch funds arrive.
head:
  - - meta
    - name: keywords
      content: Olares Payment walkthrough, LarePass sign in, receive address, payment link, merchant dashboard
---

# Hands-on walkthrough

From sign-up to your first payment, no code required. Before you start, make sure you have an [Olares ID](/manual/get-started/create-olares-id) and the LarePass app installed.

<video controls preload="none" poster="/images/payment/nocode-demo-poster.png" style="width:100%;border-radius:8px"><source src="/videos/payment/nocode-demo.mp4" type="video/mp4" /></video>

*The full flow in 2 minutes: scan to sign in → receive address ready → invoice link → buyer pays with MetaMask → payment lands in your dashboard.*

## Step 1. Scan to sign in

Open [https://www.olares.com/payment/dashboard](https://www.olares.com/payment/dashboard/) in your browser and scan the QR code with LarePass.

![Sign in with LarePass](/images/payment/dashboard-login.png#bordered)

**Your first sign-in completes onboarding**: the system creates your merchant account and configures your LarePass wallet's EVM address as the receive address on every supported chain — no forms, no copying addresses by hand.

## Step 2. Confirm your receive address

The home page greets you with a **Ready to receive** card: this is your receive address, shared for USDC / USDT, and the same address works on all seven chains (Optimism, Base, BNB Smart Chain, Ethereum, Arbitrum One, Polygon, Avalanche).

![Home: ready to receive](/images/payment/dashboard-home.png#bordered)

This address is your LarePass wallet's EVM address — once a transfer confirms on-chain, funds go straight into your own wallet. The platform never holds them.

::: tip Want a different receive address?
The default isn't fixed: go to the **Checkouts** page, remove the default configuration, and replace it with any other EVM address you own.
:::

## Step 3. Manage your checkout

Open the **Checkouts** page to see the receive address and accepted coins on each of the seven chains. You can change an address or disable a chain whenever you need.

![Checkouts: receive addresses on seven chains](/images/payment/dashboard-checkouts.png#bordered)

The **Share & collect** section at the bottom lets you create a **fixed-amount invoice** — a one-off payment link for a specific amount. Send it to your buyer to get paid.

## Step 4. Watch funds arrive

After the buyer pays and the transfer confirms on-chain, funds land directly in your receive address. The **Transactions** page and the home page's Recent transactions list every payment — buyer, amount, coin, and chain.

::: tip Want to accept payments inside your own site or app?
Create API keys, call the API, and receive webhooks — see the [developer quickstart](/developer/payment/quickstart).
:::
