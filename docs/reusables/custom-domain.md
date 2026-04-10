---
search: false
---
<!--
  Reusable content blocks for custom domain setup.

  Block 1 – Create a DID:                  lines 21-27
  Block 2 – Add domain (TXT + NS):         lines 30-65
  Block 3 – Create a new organization:      lines 68-94
  Block 4 – Add a new user:                 lines 95-110
  Block 5 – Join an existing organization:  lines 111-122

  Used by:
  - manual/best-practices/set-custom-domain.md (Blocks 1-5)
  - manual/larepass/create-org-account.md (Blocks 1, 3, 5)
  - manual/space/host-domain.md (Block 2)
  - manual/space/manage-domain.md (Block 4)
-->

<!-- Block 1: Create a DID -->
A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only associate a custom domain with the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account**.

   ![LarePass account creation page](/images/manual/tutorials/create-a-did1.png)

<!-- Block 2: Add domain (TXT + NS) -->
The following steps use `space.n1.monster` as an example custom domain.

1. In Olares Space, go to the **Domain management** page and select **Set up domain name**.

   ![Domain management page with Set up domain name button](/images/manual/tutorials/custom-domain-set-up-domain-name.png#bordered)

2. In the pop-up dialog, enter a valid subdomain, and click **Confirm**.

   :::warning Do not use a primary domain
   Using a primary domain like `yourdomain.com` will move all DNS management to Olares Space and will not transfer your existing records automatically.
   Use a subdomain like `app.yourdomain.com` instead.
   :::

3. Add and verify a TXT record to prove ownership of the domain.

   a. Click **Guide** in the **Action** column.
   ![Verify TXT](/images/manual/tutorials/custom-domain-verify-txt.png#bordered)
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

<!-- Block 3: Create a new organization -->
:::warning
Once an organization is created for a domain, the domain cannot be removed from Olares Space.
:::

1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

2. Go to **Organization Olares ID** > **Create a new organization**. Your verified domain will automatically show in the list.

   ![Organization Olares ID options in LarePass](/images/manual/tutorials/custom-domain-org-olares-id.png)

   ![Create a new organization](/images/manual/tutorials/custom-domain-create-org.png)

3. Tap the domain name.
   ![Select the domain for the organization](/images/manual/tutorials/custom-domain-select-org.png)

4. Enter the username for your Olares ID. For example, if you enter `alex`, your Olares ID will be `alex@space.n1.monster`.

   :::info
   The username part of your Olares ID should be 1-63 characters long, with lowercase alphanumeric characters only.
   :::
   ![Create an Olares ID with admin privileges](/images/manual/tutorials/custom-domain-create-olares-id-as-admin.png)

5. Tap **Confirm**.

Your Olares ID is now created and has admin privileges to manage users under this domain.

<!-- Block 4: Add a new user -->
1. In Olares Space, refresh the **Domain management** page, and click **View** next to your domain.
   ![Domain member list in Olares Space](/images/manual/tutorials/custom-domain-view-user.png#bordered)

2. Click **Add New User** and enter the username (the part before your custom domain) for the member. For example, `alice`.

   ![Add a new user dialog in Olares Space](/images/manual/tutorials/custom-domain-add-user.png#bordered)

3. Click **Submit**.
4. Repeat steps 2 and 3 to add more users.
5. Provide the full Olares ID (e.g., `alice@space.n1.monster`) and password to the user.

:::tip Manage member list
As an organization admin, you can manage your organization's member list at any time from the **Domain management** page.
:::

<!-- Block 5: Join an existing organization -->
1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

2. Go to **Organization Olares ID** > **Join an existing organization**.
   ![Join an exisitng organization](/images/manual/tutorials/custom-domain-join-org.png)

3. Enter the Olares ID (including the domain part) and password provided by the admin.

   ![Join an organization with Olares ID and password](/images/manual/tutorials/custom-domain-member-olares-id.png)

4. Tap **Continue**.
