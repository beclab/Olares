---
outline: [2, 3]
description: Create an Olares ID with a custom domain. Set up your domain for the first time or join as a member using credentials from the domain admin.
---

# Create an Olares ID with a custom domain

:::tip
For a complete end-to-end walkthrough including domain setup, Olares installation, and member onboarding, see [Set up a custom domain for your Olares](../best-practices/set-custom-domain.md).
:::

In Olares, custom domains are managed through organizations. To use a custom domain, the domain owner first creates an organization, then adds members who can create their own Olares IDs under that domain.

## Create a DID

A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only associate a custom domain with the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account**.

   ![LarePass account creation page](/images/manual/tutorials/create-a-did1.png)

Once you have a DID:
- If you are the domain owner setting up the organization, continue to [Create a new organization](#create-a-new-organization).
- If the admin has already created the organization and added you, skip to [Join an existing organization](#join-an-existing-organization).

## Create a new organization

Before you start, make sure the custom domain has been [set up in Olares Space](../space/host-domain.md).

As the domain owner, create an organization and get an admin Olares ID under your custom domain.

1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

2. Go to **Organization Olares ID** > **Create a new organization**. Your verified domain will automatically show in the list.

   ![Organization Olares ID options in LarePass](/images/manual/tutorials/custom-domain-org-olares-id.png)

   ![Organization list showing verified domain](/images/manual/tutorials/create-org1.png)

3. Tap the domain name.
4. Enter the username for your Olares ID. For example, if you want `admin123@olares.hellocoffee.online`, enter `admin123`.

   :::info
   The username part of your Olares ID should be 1-63 characters long, with lowercase alphanumeric characters only.
   :::

5. Tap **Confirm**. Your Olares ID now has admin privileges to manage users under this domain.

   ![Olares ID created with admin privileges](/images/manual/tutorials/custom-domain-admin-olares-id-created.png)

After setting up the domain, you can [add members](../space/manage-domain.md) in Olares Space.

## Join an existing organization

If the domain admin has already created the organization and added you, use the Olares ID and password provided by the admin to join.

1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
2. Go to **Organization Olares ID** > **Join an existing organization**.
3. Enter the Olares ID (including the domain part) and password provided by the admin.

   ![Joining an organization with Olares ID and password](/images/manual/tutorials/custom-domain-member-olares-id.png)

4. Tap **Continue**.

Your Olares ID is now created. You can proceed to [install and activate Olares](../get-started/install-olares.md).
