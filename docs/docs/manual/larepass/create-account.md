---
outline: [2, 4]
---

# Create and manage accounts

Managing Olares accounts is a core function of LarePass. If you are new to Olares, you will need to start by creating an Olares ID. This guide walks you through the process, and other commonly used account operations as well.

::: tip Note
Olares ID creation is only available on LarePass mobile.
:::

## Create an Olares ID

Before you start, ensure you have downloaded the [LarePass](https://olares.com/larepass) app on your phone. Depending on your personal preference, you can use one of the following options to create your Olares ID:

- **Quick creation**: Create an Olares ID by entering an available name that meets the requirements. This is the default mode.
- **Advanced creation**: Link an existing trusted identity (such as social accounts) with Olares ID using Verification Credentials (VC). This is for individual or organizational users who require enhanced security and more distinctive IDs or domains.

### Quick creation

To create an individual Olares ID quickly:

1. In LarePass app, tap **Create an account**. 

2. Enter your desired Olares ID. It must be at least 8 characters long and contain only lowercase letters and numbers.
3. Click **Continue** to finish the creation process.

   ![Fast creation](/images/manual/larepass/create-olares-id.png)

After you get your Olares ID, wait for [Olares installation](../get-started/install-olares.md) to complete, then proceed with [activation](../get-started/activate-olares.md).

### Advanced creation

::: tip VC support
Olares currently supports VC via Google Gmail. For details, refer to the [Gmail Issuer Service](/developer/contribute/olares-id/verifiable-credential/olares.md#gmail-issuer-service).
:::

<Tabs>
<template #Individual-Olares-ID>

1. In the LarePass app, tap **Create an account**.
2. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner.
3. In the **Advanced account creation** page, tap **Individual Olares ID**.
   ![Advanced account creation](/images/manual/larepass/advanced_creation.png)
4. Tap the Gmail VC option. Authenticate using your Gmail account as promoted, and then click **Continue**.
5. Wait for the binding to complete, then click **Continue** to view your Olares ID information.
   ![Olares ID with VC](/images/manual/larepass/individual_olares_id_vc.png)
</template>
<template #Organization-Olares-ID>

:::tip Note
You must have already [set up a custom domain in Olares Space](/space/host-domain.md#add-your-domain) and created the organization for it on LarePass. 
:::
1. In the LarePass app, tap **Create an account**.
1. On the account creation page, tap <i class="material-symbols-outlined">display_settings</i> in the top-right corner.
2. In the **Advanced account creation** page, tap **Organization Olares ID** > **Join an existing organization**.
    ![Advanced account creation](/images/manual/larepass/advanced_creation_org.png)
3. Enter your organization's domain name and click **Continue**.
4. Bind the VC via your email accounts. Currently, only Gmail and Google Workspace email are supported.

   ![Org ID VC](/images/manual/larepass/organization_olares_id.png)

Upon completion, you will receive an Organization Olares ID.
</template>
</Tabs>

## Import an existing account

You can also set up an account by importing an existing Olares ID:

:::tip Back up mnemonic phrase
You must have already [backed up the mnemonic phrase](back-up-mnemonics.md) for the Olares ID to import.
:::

1. In LarePass app, tap **Import an account**.
2. Enter the 12-word mnemonic phrase to import your Olares ID.
