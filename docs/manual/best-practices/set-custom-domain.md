---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares 

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this default setup saves you from common network and domain configuration hassles, you might want to use your own domain instead, especially in these common scenarios:

- **As an organization**: Use a company domain similar to your organizational email address for all team members, for example, `employee@company.com`.
- **As an individual**: Use your personal domain for a more personalized experience.

This tutorial walks you through setting up your own domain for your Olares.

## Objectives
In this tutorial, you will learn how to:
- Add and verify your custom domain in Olares Space
- Create an organization to manage your custom domain
- Configure member access for your organization
- Create an Olares ID under your custom domain
- Install and activate Olares with your Olares ID

## How custom domains work in Olares
Custom domains in Olares are managed through organizations. This means whether you're an individual user or representing a company, you'll need to set up an organization first. The organization serves as the container for your custom domain configuration.

The table below outlines the steps involved in setting up a custom domain and who is responsible for each task. Depending on whether you're an individual user or part of an organization, the actions you need to perform will differ.

| Step                                        | Individual user | Organization admin | Organization member |
|---------------------------------------------|-----------------|--------------------|---------------------|
| Create a DID                                | ✅               | ✅                  | ✅                   |
| Add domain to Olares Space                  | ✅               | ✅                  |                     |
| Create organization for the domain          | ✅               | ✅                  |                     |
| Add email to the organization               | ✅               | ✅                  |                     |
| Join the organization & create an Olares ID | ✅               | ✅                  | ✅                   |
| Set up Olares                               | ✅               | ✅                  | ✅                   |

## Before you begin

Ensure you have:
- A registered domain name from a domain registrar.
- A Gmail or G-Suite account. Currently, only these two formats are supported for organization domain membership.
- LarePass app installed on your phone.<br>
  LarePass will be used later to sign in to Olares Space, and to bind your custom domain to Olares ID.

## Step 1: Create a DID

A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only bind a custom domain to the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account** to trigger a DID creation.
   
   ![Create DID](/images/manual/tutorials/create-a-did1.png)

   This gets you an Olares account in the DID stage, which displays as "No Olares ID bound".

   ![DID stage](/images/manual/tutorials/did-stage1.png)

3. Tap the Olares account in the DID stage that you just created.

## Step 2: Add your domain to Olares Space
Add and verify your own domain in Olares Space before binding it.

1. In your browser, access Olares Space at https://space.olares.com/.
2. In LarePass app, tap the scan icon in the top-right corner, and scan the QR code on the login page to log in to Olares Space.

   ![Scan QR code](/images/manual/tutorials/scan-qr-code1.png)

3. In Olares Space, go to **Domain Management** > **Domain Name Setup**, enter your domain and click **Confirm**.

   ![Add domain](/images/manual/tutorials/add-domain.png#bordered)

4. Add and verify a TXT record to prove ownership of the domain.

   a. Click **Guide** in the **Action** column. 

   b. Follow the on-screen instructions to add a TXT record to your DNS provider configuration.

   ![Verify TXT](/images/manual/tutorials/verify-txt.png#bordered)

   Once verified, the domain setup status will update automatically to **Await NS Record for Your Domain**.
5. Verify the Name Server (NS) Record for your domain. This delegates the DNS resolution for your domain to Olares's Cloudflare. 

   a. Click **Guide** in the **Action** column. 

   b. Follow the on-screen instructions to add the NS record to your DNS provider configuration.

   Once verified, the domain status will update to **Awaiting the application for the domain's Verifiable Credential**.

   ![Domain added](/images/manual/tutorials/domain-added.png#bordered)

:::tip 
- TXT verification typically completes within 30 minutes. NS record verification may take up to 2 hours. If the whole process exceeds 3 hours, check with your DNS provider.

- Once verification is successful, do not modify the NS record. Doing so will cause the custom domain resolution to fail, making it inaccessible.
:::

Once TXT and NS records are verified, your domain is successfully added to Olares Space.

## Step 3: Create an organization for the domain

This step creates an organization for the domain. Specifically, it binds your domain to an organization in Olares and requests the Verifiable Credential (VC) for the domain.

::: tip Verifiable Credential
A Verifiable Credential is a digital format proof that verifies certain attributes or qualifications of its holder without revealing additional personal information. 
:::

1. Create a new organization in LarePass app.

   a. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
      ![Select Advanced account creation](/images/manual/tutorials/custom-domain-advanced.png)

   b. Go to **Organization Olares ID** > **Create a new organization**. The organization for your domain will automatically show in the list. 
      ![Select Organization Olares ID](/images/manual/tutorials/custom-domain-org-olares-id.png)

      ![Select Create a new organization](/images/manual/tutorials/create-org1.png)

   c. Tap the organization name to apply for the VC. When it's done, you will see your domain name for confirmation.

   d. Click **Confirm** to finish the organization domain binding in LarePass.

    ![Bind org](/images/manual/tutorials/bind-domain-with-org.png)

2. On Olares Space, navigate to the **Domain management** page. The domain setup status should change to **Awaiting rule configuration**.

So far, you have successfully bound your custom domain to an organization, and you are set to configure the domain rules in Olares Space.

## Step 4: Add a new member

The domain rules specify how you add the members for the organization. Only members in the organization can apply for Olares ID under the organization domain (or, your custom domain). To configure domain rules:

1. In Olares Space, go to **Domain management**, and click **View** next to your domain.
2. Under **Domain Invitation Rule**, select **Specified email address**, and click **Save**.
   :::tip Invitation rules
   Two types of rules are available:
   - **Fixed email suffix**: Suitable for large teams that share the same corporate email domain (e.g., `@company.com`). Any email matching the specified suffix can apply for an Olares ID under the organization. Only a single suffix is supported and must follow the G-Suite format.
   - **Specified email address**: Allows you to add members by specifying their email addresses. This option is recommended if you do not have a corporate email domain and only need to invite a small number of members to your organization. Both Gmail and G-Suite accounts are supported.
   :::

3. Click **Add New User** and enter the Gmail address for the member. For example, `justtest1953@gmail.com`.

   ![Configure domain rules](/images/manual/tutorials/set-domain-rule.png#bordered)

4. Click **Submit** to finalize the member addition. Repeat steps 3 and 4 to add more users.

:::tip Maintain member list
As an organization admin, you can manage your organization's member list at any time from the **Domain Management** page.
:::

## Step 5: Create an Olares ID with the custom domain

To use the domain, apply for an Olares ID under the organization.

1. On the account creation page of LarePass app, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
2. Tap **Organization Olares ID** > **Join an existing organization**.
3. Type the org domain name (the verified custom domain) and click **Continue**. If you see an error, verify if the domain name is correct and the domain rules are set properly in Olares Space.
4. Add a VC for the member.

   a. When prompted, select Google as your VC credential provider.

   b. Log in with the Gmail account you added in the previous step and grant access for VC.  
 
  ![Join the organization](/images/manual/tutorials/join-org.png)

After successful authorization, an Olares ID with your custom domain (for example, `justtest1953@xxxx.cloud`) is created.

## Step 6: Install and activate Olares
Almost there! Now you are all set to install and activate Olares with your Olares ID.

<tabs>
<template #Linux-and-macOS>

1. In the terminal, run the following script to start the installation:

   ```bash
   curl -sSfL https://olares.sh | bash -
   ```

2. At the end of the installation, enter your custom domain and Olares ID when prompted. In our case, it's `xxxx.cloud` for the domain, and `justtest1953` for Olares ID.

   ![Enter the domain](/images/manual/get-started/enter-olares-id.png)

3. Wait for the installation to finish. Depending on your network, the process can take 10–30 minutes. When the installation completes, you will see the wizard URL and login credentials:

    ```bash
    2024-12-17T21:00:58.086+0800        Olares is running at:
    2024-12-17T21:00:58.086+0800        http://192.168.1.16:30180

    2024-12-17T21:00:58.086+0800        Open your browser and visit the above address
    2024-12-17T21:00:58.086+0800        with the following credentials:

    2024-12-17T21:00:58.086+0800        Username: justtest1953
    2024-12-17T21:00:58.086+0800        Password: 2uO5PZ2X
    ```

4. Open the Olares activation wizard in your browser using the given URL, and follow the on-screen instructions to complete the activation.

See [Install Olares](../get-started/install-olares.md) for more detailed steps.

</template>
<template #Windows>

:::warning System environment setup required
Before proceeding with the following steps, ensure that your Windows environment is properly set up.

If the setup is incomplete, the installation script will not work as expected. For detailed instructions, refer to the dedicated [installation guide for Windows](../get-started/install-windows-script.md).
:::

1. Click https://windows.olares.sh to download the installation script `publicInstaller.latest.ps1`.

2. Run the installation script.

   a. Launch PowerShell as an administrator and navigate to the folder where you downloaded the script. For example, if the script is in your `Downloads` folder, enter:
   ```powershell
   # Replace <YourUsername> with your actual Windows username
   cd C:\Users\<YourUsername>\Downloads
   ```

   b. Run the script:

   ```powershell
   .\publicInstaller.latest.ps1
   ```

3. When you see the security warning, type `R` and press **Enter** to run the script once. The Olares installation will start.

   ```powershell
   Security warning
   Run only scripts that you trust. While scripts from the internet can be useful, this script can potentially harm your computer. If you trust this script, use the Unblock-File cmdlet to allow the script to run without this warning message. Do you want to run
   publicInstaller.latest.ps1?
   [D] Do not run [R] Run once [S] Suspend [?] Help (default is "D"):
   ```

4. When prompted to choose the drive for the WSL Ubuntu distro, enter the letter of a drive with at least 80 GB free space.

   ```powershell{8}
   Installing Olares will create a WSL Ubuntu Distro and occupy at least 80 GB of disk space.
   Please select the drive where you want to install it.
   
   Available drives and free space:
   C:\  Free Disk: 391.07 GB
   D:\  Free Disk: 281.32 GB
   
   Please enter the drive letter (e.g., C):
   ```

5. When prompted to set up firewall rules, type `yes` to configure them automatically, or `no` to skip. If you skip, you must either [disable Windows Defender Firewall](../get-started/install-windows-script.md#how-to-disable-windows-defender-firewall) or [add the TCP inbound rules manually](../get-started/install-windows-script.md#how-to-manually-set-firewall-rules).

   ```powershell{2}
   Accessing Olares requires setting up firewall rules, specifically adding TCP inbound rules for ports 80, 443, and 30180.
   Do you want to set up the firewall rules? (yes/no):
   ```

6. When prompted to confirm the Windows (NAT gateway) IP address, type **Y** to confirm or **R** to re-enter.

   ```powershell
   The NAT gateway (the Windows host)'s IP is 192.168.50.136. Confirm [Y] or Re-enter [R]?
   ```

   ::: tip Obtain the IPv4 address of Windows
   You can get the IPv4 address in advance by running `ipconfig` in the Windows command line.
   :::

7. When prompted for domain and Olares ID, enter your custom domain and the Olares ID prefix you created earlier separately. For example, if your Olares ID is `justtest1953@xxxx.cloud`, enter `xxxx.cloud` for the domain and `justtest1953` for the Olares ID.

   ![Enter domain name and Olares ID](/images/manual/get-started/enter-olares-id.png)

8. Wait for the installation to finish. Depending on your network, this can take 10–30 minutes. When it completes, the wizard URL and login credentials will be shown:

    ```bash
    2024-12-17T21:00:58.086+0800        Olares is running at:
    2024-12-17T21:00:58.086+0800        http://192.168.1.16:30180

    2024-12-17T21:00:58.086+0800        Open your browser and visit the above address
    2024-12-17T21:00:58.086+0800        with the following credentials:

    2024-12-17T21:00:58.086+0800        Username: justtest1953
    2024-12-17T21:00:58.086+0800        Password: 2uO5PZ2X
    ```

9. Open the Olares activation wizard in your browser using the URL shown, and follow the on-screen instructions to complete the activation.

For more details and troubleshooting, see the [installation guide for Windows](../get-started/install-windows-script.md).

</template>
</tabs>

After completing these steps, your Olares installation will be accessible via your custom domain.

## Learn more

- [Olares account](../../developer/concepts/account.md)
- [Install Olares](../get-started/install-olares.md)
- [Configure domain rules](../../space/manage-domain.md)