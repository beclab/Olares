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

<!--@include: ../../reusables/custom-domain.md{30,65}-->

## What's next

In LarePass, create your organization and Olares ID on the domain you added here. For details, see [Create an Olares ID with a custom domain](../larepass/create-org-account.md).