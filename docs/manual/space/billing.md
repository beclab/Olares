---
outline: [2, 3]
description: Understand Olares Space billing, including service charges, usage-based pricing, payment workflows, and promotional coupons.
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, billing, Stripe, payment, coupons
---

# Billing and payments

Olares partners with Stripe for secure payment processing. Your invoices and receipts are sent by email so you can keep records. If you have any questions during the payment process, please contact the Olares team for help.

## Billing overview

### What you are billed for

The following services or products are billed:

| Item | Billing type | Description |
|---|---|---|
| **Subscription** | Monthly subscription | Monthly plan fee for the paid plan **Basic**. The **Free** plan has no subscription fee. |
| **Backup storage** | Postpaid (usage-based) | Host data backed up to public storage. The first 50 GB per month is free. Additional usage is billed at the end of the billing cycle. |
| **Cloudflare network traffic** | Postpaid (usage-based) | Traffic between the public internet and Olares Space, routed through Cloudflare's network (CDN/WAF/DDoS) when accessing self-hosted Olares via public links. The first 50 GB per month is free. |
| **Olares Tunnel traffic (FRP)** | Postpaid (usage-based) | Traffic between Olares Space and your self-hosted Olares, forwarded through Olares Tunnel (reverse proxy). The first 50 GB per month is free. |

:::tip Avoid traffic charges
To avoid Cloudflare network and Olares Tunnel traffic charges, access your Olares via [LarePass VPN](../larepass/private-network.md) instead of public links.
:::

### When bills are generated

The following actions generate bills:

| Action | Billing result |
|---|---|
| **Upgrade your plan** | A charge is generated at checkout when you upgrade from Free to Basic. |
| **Back up data to Olares Space** | Storage usage is billed at the end of the billing cycle. |
| **Access Olares via a public link** | Cloudflare network traffic and Olares Tunnel (FRP) traffic are billed based on usage beyond the free monthly quota. |

### How fees are deducted monthly

Olares operates on a monthly billing cycle. Each month, a single bill is generated that includes the base subscription fee plus any additional charges incurred during the month, such as extra traffic and storage fees.

- **Automatic deduction**: Your first payment authorizes automatic deductions, and the first payment method is set as default. To stop automatic deductions, go to **Usage & billing** > **Payment methods**, and [delete your saved payment method](#manage-payment-methods).
- **Payment validity period**: Bills must be paid within 3 hours. Unpaid bills are canceled automatically. Make sure you settle your bills on time to avoid service interruption.

## View your plan

Check your current subscription plan and billing cycle.

1. Click **Usage & billing** in the left navigation pane.
2. Select the **Plan** tab.

   ![Plan tab](/images/how-to/space/billing_plan.png#bordered)

3. Review the following information:
   - **Plan**: Your current subscription plan, such as **Basic** or **Free**. To upgrade from Free to Basic, see [Change your plan](#change-your-plan).
   - **Billing cycle**: The payment method, billing period, and total amount.
   - **Details**: The cost breakdown, including the plan fee.

## Change your plan

Upgrade from Free to Basic to get more traffic, bandwidth, and storage.

1. Click **Usage & billing** in the left navigation pane.
2. Select the **Plan** tab.
3. Click **Change** next to your current plan.

   ![Change plan](/images/how-to/space/billing_plan_change.png#bordered)

4. Compare the available plans. For example, **Basic** includes more monthly traffic, higher bandwidth, regional FRP nodes, Olares Space storage, and priority email support.

   ![Compare plan](/images/how-to/space/billing_plan_compare.png#bordered)

   For the latest plan features and pricing, see [Olares Space plans](https://www.olares.com/space/plans).

5. Hover over **Basic**, and then click **Select**.
6. In the **Checkout** panel, review your account details, payment method, and amount due.
7. Click **Confirm** to complete the upgrade.

:::info
- Your account balance is applied first. Only the remaining amount is charged to your saved card.
- The plan change starts a new billing cycle right away.
- Any unused quota from your current billing period will not be refunded.
:::

## View and download invoices

View and download your past invoices.

1. Click **Usage & billing** in the left navigation pane.
2. Select the **Invoices** tab.

   ![Invoices tab](/images/how-to/space/billing_invoices.png#bordered)

3. Review the invoice list.
4. To download an invoice, click <i class="material-symbols-outlined">download</i> in the **Action** column.

   The downloaded PDF includes the invoice number, issue date, due date, billing details, and a breakdown of charges.

## Redeem a coupon

You can redeem most coupon codes on the **Coupon** tab. If a promotion provides a different redemption method, such as binding your account email, follow the instructions in the campaign material.

1. Click **Usage & billing** in the left navigation pane.
2. Select the **Coupon** tab.

   ![Coupon tab](/images/how-to/space/billing_coupon.png#bordered)

3. Enter your coupon code in the input field.
4. Click **Redeem now**.

:::info
Coupon values are non-refundable.
:::

## Manage payment methods

Add, update, or remove your credit card on the **Payment methods** tab.

1. Click **Usage & billing** in the left navigation pane.
2. Select the **Payment methods** tab.
3. To add a card, click **Add credit card**, enter the card details, and then click **Save**. 

    :::info
    Olares Space only supports one saved payment method at a time.
    :::

4. To replace the saved card, click **Edit**, enter the new card details, and then click **Save**.
5. To remove the saved card, click **Delete**.

   ![Payment methods panel](/images/how-to/space/payment_methods.png#bordered)

## FAQs

### Will I be charged if someone accesses the WordPress site deployed on Olares?

Yes. When someone accesses your Olares services through a public link, two types of traffic charges may apply:

- **Cloudflare network traffic**: Traffic routed through Cloudflare's network for CDN, WAF, and DDoS protection.
- **Olares Tunnel traffic (FRP)**: Traffic forwarded through Olares Tunnel to your self-hosted Olares.

Both include a free monthly quota. Usage beyond the quota is billed. To avoid these charges, access your Olares via [LarePass VPN](../larepass/private-network.md) instead of public links.

### What happens if my bill is less than $1?

If your total bill is under $1, it won't trigger a card charge. Instead, this amount will be added to your balance and rolled into the next bill.

### How do I resolve a negative balance?

A negative balance doesn't always mean you owe money. To check the details, click your profile avatar in the upper-right corner and click **Balance**. The **Balance details** page opens, showing the breakdown.

If you owe any money, pay it as soon as possible to avoid service interruptions.
