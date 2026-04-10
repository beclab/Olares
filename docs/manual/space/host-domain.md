---
outline: [2, 3]
description: Set up custom domains in Olares Space with domain verification and DNS configuration. Create organizational Olares IDs and manage domain settings for your team.
---

# Set up a custom domain

Whether you're an organizational user wanting employees to use a company-specific domain, or you simply want to use a domain that you own, Olares Space allows you to set up a custom domain for your Olares system.

:::tip First time setting up a custom domain?
For a complete end-to-end walkthrough including DID creation, Olares installation, and member onboarding, see [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).
:::

## Prerequisites

Make sure you have:
- An account in the DID stage. A custom domain can only be associated with an account in the DID stage. See [Create a DID](../larepass/create-org-account.md#create-a-did) for instructions.
- Logged into Olares Space with your DID. See [Log in and manage accounts](manage-accounts.md).
- A domain name registered through a domain registrar. The domain should not already be associated with another account in Olares Space.
- LarePass installed on your phone.
- Access to the DNS settings of your domain for configuring TXT and NS records.

## Add your domain

The following steps use `olares.hellocoffee.online` as an example custom domain.

1. In Olares Space, go to the **Domain management** page and select **Set up domain name**.

   ![Domain management page with Set up domain name button](/images/manual/tutorials/custom-domain-set-up-domain-name.png#bordered)

2. In the pop-up dialog, enter a valid subdomain like `olares.hellocoffee.online`, and click **Confirm**.

   :::warning Do not use a primary domain
   Using a primary domain like `yourdomain.com` will move all DNS management to Olares Space and will not transfer your existing records automatically.
   Use a subdomain like `app.yourdomain.com` instead.
   :::

3. Add and verify a TXT record to prove ownership of the domain.

   a. Click **Guide** in the **Action** column.

   b. Follow the on-screen instructions to add a TXT record to your DNS provider configuration.

   Once verified, the status updates to **Awaiting NS record configuration**.
   ![Domain status updated to Awaiting NS record configuration](/images/manual/tutorials/custom-domain-add-ns.png#bordered)

4. Verify the Name Server (NS) record for your custom domain. This delegates the DNS resolution for your domain to Olares's Cloudflare.

   a. Click **Guide** in the **Action** column.

   b. Follow the on-screen instructions to add two NS records to your DNS provider configuration.

   Once verified, the domain status will update to **Awaiting the application for the Domain's Verifiable Credential**.
   ![Domain status updated to Awaiting the application for the Domain's Verifiable Credential](/images/manual/tutorials/custom-domain-wait-vc.png#bordered)

   :::warning
   Once verification is successful, do not modify the NS record. Doing so will cause the custom domain resolution to fail, making it inaccessible.
   :::

Once TXT and NS records are verified, your domain is successfully added to Olares Space.

## What's next

To create an Olares ID under your custom domain, see [Create an Olares ID with a custom domain](../larepass/create-org-account.md).