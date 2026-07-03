---
search: false
---
<!--
  Reusable content blocks for custom domain setup. Include by named region.

  Used by:
  - manual/best-practices/set-custom-domain.md
  - manual/larepass/create-org-account.md
  - manual/space/host-domain.md
  - manual/space/manage-domain.md
-->

<!-- #region custom-domain-create-did -->
A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only associate a custom domain with the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account**.

   ![LarePass account creation page](/images/manual/tutorials/create-a-did1.png)

   This creates an Olares account in the DID stage. On the **Switch account** page, it displays as "No Olares ID bound" with an identifier like `did:key:xxxx`.

   ![DID stage](/images/manual/tutorials/did-stage1.png)
<!-- #endregion custom-domain-create-did -->

The following steps use `space.n1.monster` as an example custom domain.

<!-- #region custom-domain-add-domain-steps -->
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
   b. In your DNS provider's settings, add a TXT record with the name and value provided in the dialog.

      ![Add TXT record in DNS provider](/images/manual/tutorials/custom-domain-add-txt-record.png#bordered)

   Once verified, the status updates to **Awaiting NS record configuration**.
   ![Domain status updated to Awaiting NS record configuration](/images/manual/tutorials/custom-domain-add-ns.png#bordered)

4. Verify the Name Server (NS) record for your custom domain. This delegates the DNS resolution for your domain to Olares's Cloudflare.

   a. Click **Guide** in the **Action** column.

   b. In your DNS provider's settings, add two NS records for your subdomain with the values provided in the dialog.

      ![Add NS records in DNS provider](/images/manual/tutorials/custom-domain-add-ns-record.png#bordered)

   Once verified, the domain status will update to **Awaiting the application for the Domain's Verifiable Credential**.
   ![Domain status updated to Awaiting the application for the Domain's Verifiable Credential](/images/manual/tutorials/custom-domain-wait-vc.png#bordered)

   :::warning
   Once verification is successful, do not modify the NS record. Doing so will cause the custom domain resolution to fail, making it inaccessible.
   :::

Once TXT and NS records are verified, you can proceed to create an organization in LarePass.
<!-- #endregion custom-domain-add-domain-steps -->

<!-- #region custom-domain-create-organization -->
1. Open LarePass on your phone, and in the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
   ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

2. Go to **Organization Olares ID** > **Create a new organization**. Your verified domain will automatically show in the list.

   ![Organization Olares ID options in LarePass](/images/manual/tutorials/custom-domain-org-olares-id.png)

   ![Create a new organization](/images/manual/tutorials/custom-domain-create-org.png)

3. Tap the domain name.
   ![Select the domain for the organization](/images/manual/tutorials/custom-domain-select-org.png)

   :::warning
   Once an organization is created for a domain, the domain cannot be removed from Olares Space.
   :::

4. Enter the username for your Olares ID. For example, if you enter `alex`, your Olares ID will be `alex@space.n1.monster`.

   :::info
   The username part of your Olares ID should be 2-24 characters long, with lowercase alphanumeric characters only.
   :::
   ![Create an Olares ID with admin privileges](/images/manual/tutorials/custom-domain-create-olares-id-as-admin1.png)

5. Tap **Confirm**.

   ![Olares ID created with admin privileges](/images/manual/tutorials/custom-domain-admin-id-created.png)

   Your Olares ID is now created and has admin privileges to manage users under this domain.
<!-- #endregion custom-domain-create-organization -->

<!-- #region custom-domain-install-and-activate-olares -->
Now you can install and activate Olares with your Olares ID.

Use a new or factory-reset Olares device for this step. An Olares device that has already been activated with another Olares ID cannot be switched to the custom-domain Olares ID in place.

The installation steps are similar to the standard process. The following example uses Linux. For other systems, refer to the [installation guide](/manual/get-started/install-olares).

:::warning Same network required
To avoid activation failures, ensure that both your phone and the Olares device are connected to the same network.
:::

1. Open a terminal on the machine where you want to install Olares, and run the following command:

   ```bash
   export PREINSTALL=1 &&
   curl -sSfL https://olares.sh | bash -
   ```

   This runs a partial installation (prepare phase only) without proceeding to full setup.

2. Open LarePass on your phone, and on your Olares activation page, tap **Discover nearby Olares**. LarePass will list the detected Olares instances in the same network.
3. Select the target Olares instance from the list and tap **Install now**.

   ![ISO Activate](/images/manual/larepass/iso-activate1.png#bordered)

4. When the installation completes, tap **Activate now**.
5. In the **Select a reverse proxy** dialog, select a node that is closer to your geographical location. The installer will then configure the HTTPS certificate and DNS for Olares.

   :::tip Note
   - You can change this setting later on the [Change reverse proxy](/manual/olares/settings/change-frp.md) page in Olares.
   - If your Olares device is connected to a public IP network, this step will be skipped automatically.
   :::

6. Follow the on-screen instructions to set the login password for Olares, then tap **Complete**.

   ![ISO Activate-2](/images/manual/larepass/iso-activate-4.png#bordered)

Once activation is complete, LarePass will display the desktop address of your Olares device with the custom domain, such as `https://desktop.alex.space.n1.monster`.
<!-- #endregion custom-domain-install-and-activate-olares -->

<!-- #region custom-domain-add-user -->
1. In Olares Space, refresh the **Domain management** page, the domain status now updated to **Allocated**.
   ![Domain member list in Olares Space](/images/manual/tutorials/custom-domain-view-user.png#bordered)

2. Click **View** in the **Action** column.

3. Click **Add New User** and enter the username (the part before your custom domain) for the member. For example, `alice`.

   ![Add a new user dialog in Olares Space](/images/manual/tutorials/custom-domain-add-user.png#bordered)

4. Click **Submit**.

   ![User added to the organization](/images/manual/tutorials/custom-domain-user-added.png#bordered)

5. (Optional) Repeat steps 2 and 3 to add more users.
6. Provide the username and password to the member.
<!-- #endregion custom-domain-add-user -->

:::tip Manage member list
As an organization admin, you can manage your organization's member list at any time from the **Domain management** page.
:::

<!-- #region custom-domain-join-organization -->
1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   ![Advanced account creation option in LarePass](/images/manual/tutorials/custom-domain-advanced.png)

2. Go to **Organization Olares ID** > **Join an existing organization**.
   ![Join an exisitng organization](/images/manual/tutorials/custom-domain-join-org.png)

3. Enter the username with the domain part (e.g., `alice@space.n1.monster`) and the password provided by the admin.
   
   :::tip One-time password
   This password verifies your identity when you create your Olares ID, and is for one-time use only.
   :::

   ![Join an organization with Olares ID and password](/images/manual/tutorials/custom-domain-member-olares-id.png)

4. Tap **Continue**.
<!-- #endregion custom-domain-join-organization -->
