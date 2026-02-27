---
outline: [2, 3]
description: Step-by-step guide to setting up a custom domain for your Olares environment. Learn how to add and verify domains, create organizations, configure member access, and create Olares IDs under your domain.
---

# Set up a custom domain for your Olares 

By default, when you create an account in LarePass, you get an Olares ID with the `olares.com` domain. This means you access your Olares services through URLs like `desktop.{your-username}.olares.com`. While this default setup saves you from common network and domain configuration hassles, you might want use your own domain instead, especially in these common scenarios:

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
- A Gmail or G-Suit account. Currently, only these two formats are supported for organization domain membership.
- LarePass app installed on your phone.<br>
  LarePass will be used later to sign in to Olares Space, and to bind your custom domain to Olares ID.

## Step 1: Create a DID

A DID (Decentralized Identifier) is a temporary account state before you get your final Olares ID. You can only bind a custom domain to the account when it is in the DID stage. To create one:

1. In the LarePass app, go to the account creation page.

2. Tap **Create an account** to trigger a DID creation.
   
   ![create DID](/images/manual/tutorials/create-a-did.png)

   This gets you an Olares account in the DID stage. 

   ![DID stage](/images/manual/tutorials/did-stage.png)

## Step 2: Add your domain to Olares Space
Add and verify your own domain in Olares Space before binding it.

1. In your browser, access Olares Space at https://space.olares.com/.
2. In LarePass app, tap the scan button in the top-right corner, and scan the QR code on the login page to log in to Olares Space.

   ![scan QR](/images/manual/tutorials/scan-qr-code.png)

3. In Olares Space, go to **Domain Management** > **Domain Name Setup**, enter your domain and click **Confirm**.

   ![add domain](/images/manual/tutorials/add-domain.png#bordered)

4. Verify your TXT record for your domain. This verifies your ownership of the domain.

   a. Click **Guide** in the **Action** column. 

   b. Follow the on-screen instructions to add a TXT record to your DNS provider configuration.

   ![verify TXT](/images/manual/tutorials/verify-txt.png#bordered)

   Once verified, the domain setup status will update automatically to **Await NS Record for Your Domain**.
5. Verify the Name Server (NS) Record for your domain. This delegates the DNS resolution for your domain to Olares's Cloudflare. 

   a. Click **Guide** in the **Action** column. 

   b. Follow the on-screen instructions to add the NS record to your DNS provider configuration.

   Once verified, the domain status will update to **Awaiting the application for the domain's Verifiable Credential**.

   ![domain added](/images/manual/tutorials/domain-added.png#bordered)

:::tip 
- TXT verification typically completes within 30 minutes. NS record verification may take up to 2 hours. If the whole process exceeds 3 hours, check with your DNS provider.

- Once verification is successful, do not modify the NS record. Doing so will cause the custom domain resolution to fail, making it inaccessible.
:::

Once TXT and NS records are verified, your domain is successfully added to Olares Space.

## Step 3: Create an org for the domain

This step creates an organization for the domain. Specifically, it binds your domain to an organization in Olares and requests the Verifiable Credential (VC) for the domain.

::: tip Verifiable Credential
A Verifiable Credential is a digital format proof that verifies certain attributes or qualifications of its holder without revealing additional personal information. 
:::

1. Create a new organization in LarePass app.

   a. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.

   b. Go to **Organization Olares ID** > **Create a new organization**. The organization for your domain will automatically show in the list. 

      ![Create org](/images/manual/tutorials/create-org.png)

   c. Tap the organization name to apply for the VC. When it's done, you will see your domain name for confirmation.

   d. Click **Confirm** to finish the organization domain binding in LarePass.

    ![Bind org](/images/manual/tutorials/bind-domain-with-org.png)

2. On Olares Space, navigate to the **Domain management** page. The domain setup status should change to **Awaiting rule configuration**.

So far, you have successfully bound your custom domain with an organization, and is set for configuring the domain rules in Olares Space.

## Step 4: Add new member

The domain rules specify how you add the members for the organization. Only members in the organization can apply for Olares ID under the organization domain (or, your custom domain). To configure domain rules:

1. In Olares Space, go to **Domain management**, and click **View** next to your domain.
2. Under **Domain Invitation Rule**, select **Specified email address**, and click **Save**.
   :::tip Invitation rules
   Two types of rules are available:
   - **Fixed email suffix**: Suitable for large teams who share the same corporation email domain (e.g., `@company.com`). Any email matching the specified suffix is valid to apply for Olares ID under the organization. Currently, only single suffix is supported. Must follow G-Suite format.
   - **Specified email address**: Allows you to add members by specifying their email addresses. This option is recommended if you do not have a corporate email domain and only need to invite a small number of members to your organization. Both Gmail and G-Suite accounts are supported.
   :::

3. Click **Add New User** and enter the Gmail address for the member. For example, `justtest1953@gmail.com`.

   ![Configure domain rules](/images/manual/tutorials/set-domain-rule.png#bordered)

4. Click **Submit** to finalize the member addition. Repeat step 3 and step 4 if you want to add multiple users

:::tip Maintain member list
For organization admin, you can manage your organization's member list anytime through the **Domain management** page.
:::

## Step 5: Create an Olares ID with the custom domain

To use the domain, apply for an Olares ID under the organization.

1. On the account creation page of LarePass app, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner to go to the **Advanced account creation** page.
2. Tap **Organization Olares ID** > **Join an existing organization**.
3. Type the org domain name (the verified custom domain) and click **Continue**. If you see an error, verify if the domain name is correct and the domain rules are set properly in Olares Space.
4. Add a VC for the member.

   a. When prompted, select Google as your VC credential provider.

   b. Log in with the Gmail account you added in the previous step and grant access for VC.  
 
  ![Join the org](/images/manual/tutorials/join-org.png)

 After successful authorization, an Olares ID with the custom domain, `justtest1953@xxxx.cloud`, is successfully created.

## Step 6: Install and activate Olares
Almost there! Now you are all set to install and activate Olares with your Olares ID.

<tabs>
<template #Linux-and-macOS>

1. In the terminal, run the following script to start the installation:

   ```bash {1,2}
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

3. Open the Olares activation wizard in your browser using the given URL, and follow the on-screen instructions to complete the activation.

See [Install Olares](../get-started/install-olares.md) for more detailed steps.

</template>
<template #Windows>

:::warning System environment setup required
Before proceeding with the following steps, ensure that your Windows environment is properly set up.

If the setup is incomplete, the installation script will not work as expected. For detailed instructions, refer to the dedicated [installation guide for Windows](../get-started/install-windows-script.md).
:::

1. Click https://windows.olares.sh to download the installation script `publicInstall.latest.ps1`.

2. Open `publicInstall.latest.ps1` with Notepad, and add the environment variables to the beginning:

   ```bash {1,2}
   $env:TERMINUS_OS_DOMAINNAME = "xxxx.cloud"
   $env:TERMINUS_OS_USERNAME= "justtest1953"
   $env:WSL_UTF8 = 1
   $OutputEncoding = [System.Text.Encoding]::UTF8
   $currentPath = Get-Location
   $architecture = $env:PROCESSOR_ARCHITECTURE
   $downloadCdnUrlFromEnv = $env:DOWNLOAD_CDN_URL
   $version = "1.11.1"
   $downloadUrl = "https://dc3p1870nn3cj.cloudfront.net"
   ```
   - `$env:TERMINUS_OS_DOMAINNAME=xxxx.cloud`: Specify your custom domain. Replace `xxxx.cloud` with the actual one.
   - `$env:TERMINUS_OS_USERNAME=justtest1953`: Specify the prefix of your Olares ID. Replace `justtest1953` with the actual one.

3. Execute the script.

   a. Open PowerShell as administrator, then navigate to the folder where the script is located. For example, if the script is in the `Downloads` folder, run the following command:
   ```powershell
   cd C:\Users\<YourUsername>\Downloads
   ```

   b. Once in the correct folder, run the following command:
   ```powershell
   .\publicInstall.latest.ps1
   ```

4. When prompted with security warning, type `R` and press **Enter** to run the script once. The installation process for Olares will start.

   ```powershell
   Security warning
   Run only scripts that you trust. While scripts from the internet can be useful, this script can potentially harm your computer. If you trust this script, use the Unblock-File cmdlet to allow the script to run without this warning message. Do you want to run
   publicInstall.latest.ps1?
   [D] Do not run [R] Run once [S] Suspend [?] Help (default is "D"):
   ```

5. Wait for the installation to finish. Depending on your network, the process can take 20-30 minutes. When the installation completes, you will see the wizard URL and login credentials:

    ```bash
    2024-12-17T21:00:58.086+0800        Olares is running at:
    2024-12-17T21:00:58.086+0800        http://192.168.1.16:30180

    2024-12-17T21:00:58.086+0800        Open your browser and visit the above address
    2024-12-17T21:00:58.086+0800        with the following credentials:

    2024-12-17T21:00:58.086+0800        Username: justtest1953
    2024-12-17T21:00:58.086+0800        Password: 2uO5PZ2X
    ```

6. Open the Olares activation wizard in your browser using the given URL, and follow the on-screen instructions to complete the activation.

</template>
</tabs>

After completing these steps, your Olares installation will be accessible via your custom domain.

## Learn more

- [Olares account](../../developer/concepts/account.md)
- [Install Olares](../get-started/install-olares.md)
- [Configure domain rules](../../space/manage-domain.md)