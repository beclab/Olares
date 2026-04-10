---
outline: [2, 3]
description: Create an Olares ID with a custom domain. Set up your domain for the first time or join as a member using credentials from the domain admin.
---

# Create an Olares ID with a custom domain

In Olares, custom domains are managed through organizations. To use a custom domain, the domain owner first creates an organization, then adds members who can create their own Olares IDs under that domain.

Before you start, make sure the custom domain has been [verified in Olares Space](../space/host-domain.md).

## Create a new organization

As the domain owner, create an organization and get an admin Olares ID under your custom domain.

1. In the LarePass app, tap **Create an account**.

   ![Create DID](/images/manual/tutorials/create-a-did1.png)

2. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   ![Select Advanced account creation](/images/manual/tutorials/custom-domain-advanced.png)

3. Go to **Organization Olares ID** > **Create a new organization**. Your verified domain will automatically show in the list.

   ![Select Organization Olares ID](/images/manual/tutorials/custom-domain-org-olares-id.png)

   ![Select Create a new organization](/images/manual/tutorials/create-org1.png)

4. Tap the domain name.
5. Enter the username for your Olares ID. For example, if you want `admin123@olares.hellocoffee.online`, enter `admin123`.

   :::info
   The username part of your Olares ID should be 1-63 characters long, with lowercase alphanumeric characters only.
   :::

6. Click **Confirm**. Your Olares ID now has admin privileges to manage users under this domain.

   ![Admin Olares ID created](/images/manual/tutorials/custom-domain-admin-olares-id-created.png)

After setting up the domain, you can [add members](../space/manage-domain.md) in Olares Space.

## Join an existing organization

If the domain admin has already created the organization and added you, use the Olares ID and password provided by the admin to join.

1. In the LarePass app, tap **Create an account**.
2. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
3. Go to **Organization Olares ID** > **Join an existing organization**.
4. Enter the Olares ID (including the domain part) and password provided by the admin.

   ![Join with custom domain](/images/manual/tutorials/custom-domain-member-olares-id.png)

5. Click **Continue**.

Your Olares ID is now created. You can proceed to [install and activate Olares](../get-started/install-olares.md).
