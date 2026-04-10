---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this works out of the box, you might want to use your own domain instead.

## Learning objectives
In this tutorial, you will learn how to:
- Add and verify your custom domain in Olares Space
- Create an organization to manage your custom domain
- Configure member access for your organization
- Create an Olares ID under your custom domain
- Install and activate Olares with your Olares ID

## How custom domains work in Olares
Custom domains in Olares are managed through organizations. Whether you're an individual user or representing a company, you'll need to set up an organization first. The required actions depend on your role:

| Step                                                                              | Organization admin | Organization member |
|-----------------------------------------------------------------------------------|--------------------|---------------------|
| Create a DID in LarePass                                                          | ✅                  | ✅                   |
| Add custom domain to Olares Space                                                 | ✅                  |                     |
| Create organization for the domain &<br> create an Olares ID as the admin in LarePass | ✅                  |                     |
| Add new user to the organization in Olares Space                                  | ✅                  |                     |
| Join the organization & create an Olares ID in LarePass                           |                    | ✅                   |
| Set up Olares                                                                     | ✅                  | ✅                   |

If you are joining an organization, you can skip to [Join an organization as a member](#join-an-organization-as-a-member).

## Prerequisites

Make sure you have:
- A valid domain name from a domain registrar.
- LarePass app installed on your phone. You will use LarePass to sign in to Olares Space and to associate your custom domain with an Olares ID.

:::info
If you have previously installed and activated an Olares instance on your device and want to reuse the same hardware with a custom domain, [perform a factory reset](../larepass/manage-olares.md#restore-olares-to-factory-settings) first, then [create a new account](../larepass/create-account.md) to follow this tutorial.
:::

## Step 1: Create a DID

A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only associate a custom domain with the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account**.
   
   ![LarePass account creation page](/images/manual/tutorials/create-a-did1.png)
   
<!-- 

   This gets you an Olares account in the DID stage, which displays as "No Olares ID bound".

   ![DID stage](/images/manual/tutorials/did-stage1.png)

3. Tap the Olares account in the DID stage that you just created.

 -->

## Step 2: Add your domain to Olares Space

The following steps use `olares.hellocoffee.online` as an example custom domain.

1. In your browser, open [Olares Space](https://space.olares.com/).
2. In the LarePass app, tap the scan icon in the top-right corner and scan the QR code on the login page to log in to Olares Space.

   ![LarePass QR code scanner](/images/manual/tutorials/scan-qr-code1.png)

3. In Olares Space, go to the **Domain management** page and select **Set up domain name**.

   ![Domain management page with Set up domain name button](/images/manual/tutorials/custom-domain-set-up-domain-name.png#bordered)

4. In the pop-up dialog, enter a valid subdomain like `olares.hellocoffee.online`, and click **Confirm**.
   :::warning Do not use a primary domain
   Using a primary domain like `yourdomain.com` will move all DNS management to Olares Space and will not transfer your existing records automatically.
   Use a subdomain like `app.yourdomain.com` instead.
   :::

5. Add and verify a TXT record to prove ownership of the domain.

   a. Click **Guide** in the **Action** column.

   b. Follow the on-screen instructions to add a TXT record to your DNS provider configuration.

   Once verified, the status updates to **Awaiting NS record configuration**.
   ![Domain status updated to Awaiting NS record configuration](/images/manual/tutorials/custom-domain-add-ns.png#bordered)


6. Verify the Name Server (NS) record for your custom domain. This delegates the DNS resolution for your domain to Olares's Cloudflare.

   a. Click **Guide** in the **Action** column. 

   b. Follow the on-screen instructions to add two NS records to your DNS provider configuration.

   Once verified, the domain status will update to **Awaiting the application for the Domain's Verifiable Credential**.
   ![Domain status updated to Awaiting the application for the Domain's Verifiable Credential](/images/manual/tutorials/custom-domain-wait-vc.png#bordered)

   :::warning
   Once verification is successful, do not modify the NS record. Doing so will cause the custom domain resolution to fail, making it inaccessible.
   :::

Once TXT and NS records are verified, your domain is successfully added to Olares Space.

## Step 3: Create an organization for the domain

This step links your domain to an organization in Olares and requests the Verifiable Credential (VC) for the domain.

::: tip Verifiable Credential
A Verifiable Credential (VC) is a digital certificate that confirms your organization's ownership of the domain.
:::

1. Create a new organization in the LarePass app.

   a. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
      ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

   b. Go to **Organization Olares ID** > **Create a new organization**. The organization for your domain will automatically show in the list. 
      ![Organization Olares ID options in LarePass](/images/manual/tutorials/custom-domain-org-olares-id.png)

      ![Organization list showing verified domain](/images/manual/tutorials/create-org1.png)

   c. Tap the organization name.

2. Create an Olares ID with the custom domain. For example, if you want an Olares ID called `admin123@olares.hellocoffee.online`, enter `admin123`.
   :::info
   The username part of your Olares ID should be 1-63 characters long, with lowercase alphanumeric characters only.
   :::

3. Tap **Confirm**. Your Olares ID now has admin privileges to manage organization users.
   ![Olares ID created with admin privileges](/images/manual/tutorials/custom-domain-admin-olares-id-created.png)

## Step 4: Install and activate Olares

Now you can install and activate Olares with your Olares ID.

The installation steps are similar to the standard process. The following example uses Linux. For other systems or a clean install with an ISO image, refer to the [installation guide](../get-started/install-olares.md).
<!--@include: ../get-started/install-and-activate-olares.md{5,7}-->

1. In the terminal, run the following script to start the installation:

   ```bash
   export PREINSTALL=1 &&
   curl -sSfL https://olares.sh | bash -
   ```
   This runs a partial installation (prepare phase only) without proceeding to full setup.

<!--@include: ../get-started/install-and-activate-olares.md{9,13}-->

<!--@include: ../get-started/install-and-activate-olares.md{20,20}-->

Once activation is complete, LarePass will display the desktop address of your Olares device with the custom domain, such as `https://desktop.admin123.olares.hellocoffee.online`.

## Step 5: Add a new user within the same organization

1. In Olares Space, go to **Domain management**, and click **View** next to your domain.

2. Click **Add New User** and enter the username (the part before your custom domain) for the member. For example, `admin456`.

   ![Add a new user dialog in Olares Space](/images/manual/tutorials/custom-domain-add-user.png#bordered)

3. Click **Submit**. Repeat steps 2 and 3 to add more users.

4. Provide the Olares ID and the corresponding password to the user.

5. (Optional) If the member will join the admin's existing Olares instance instead of installing on a separate device, go to **Settings** > **Users** in the admin's Olares and create a user account for the member. Share the activation wizard URL and one-time password with the member. See [Manage your team](../olares/settings/manage-team.md) for details.

:::tip Manage member list
As an organization admin, you can manage your organization's member list at any time from the **Domain Management** page.
:::

## Join an organization as a member

1. In the LarePass app, tap **Create an account**.
2. Tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
      ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

3. Go to **Organization Olares ID** > **Join an existing organization**.

4. Enter the Olares ID (including the domain part) and password.
   ![Joining an organization with Olares ID and password](/images/manual/tutorials/custom-domain-member-olares-id.png)

5. Tap **Continue**.

Your custom Olares ID is now created. Next, activate Olares:

- **Separate device**: Follow [Step 4](#step-4-install-and-activate-olares) to install and activate Olares on a new device.
- **Same device as the admin**: Get the activation wizard URL and one-time password from the admin, then scan the wizard QR code in LarePass to activate. See [Activate Olares](../get-started/join-olares.md) for details. 


## Learn more

- [Olares account](../../developer/concepts/account.md): How DIDs, Olares IDs, and organizations work.
- [Install Olares](../get-started/install-olares.md): Installation options for different platforms and environments.
- [Manage your team](../olares/settings/manage-team.md): Create and manage user accounts within an Olares instance.
