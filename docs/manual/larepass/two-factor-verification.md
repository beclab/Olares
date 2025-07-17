---
description: Set up and use two-factor authentication in Olares' Vault to generate secure TOTP codes for your online accounts, enhancing login security across your services.
---
# Generate two-factor authentication codes
Two-factor authentication (2FA) requires both your password and an authentication code when signing in. These codes are generated using Time-Based One-Time Password (TOTP), which creates temporary codes that refresh automatically. Similar to Google Authenticator or Microsoft Authenticator, Vault can generate secure 2FA codes for your online accounts.

This guide explains how to generate two-factor authentication (2FA) codes in Vault.
## Prepare your target service
1. Log in to the website where you want to enable 2FA (e.g., GitHub or OpenAI).
2. Navigate to the security settings page and enable two-factor authentication using authenticator app.

   ![Enable GitHub 2FA](/images/manual/olares/2fa-github.png#bordered)
3. Save the provided secret key or QR code for the next steps.
:::tip
If the service provides recovery codes, store them securely. They are crucial to account recovery if you lose access to Vault.
:::

## Create an authenticator in Vault
:::tip
Visit the [official page](https://olares.com/larepass) for download options.
:::
<tabs>
<template #Olares,-LarePass-desktop,-or-browser-extension>

1. In Vault, click <i class="material-symbols-outlined">add</i> in the top right corner.
2. Select **Authenticator** as the item type, and click **Create**.
3. Fill in the required fields:
    - Item name: enter a descriptive name or the service. For example, `GitHub`.
    - One-time password: Paste the secret key.
4. Click **Save**.
</template>

<template #LarePass-mobile>

1. Open LarePass on your device, and navigate to the **Vault** page within the app.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. Select **Authenticator** as the item type, and click **Create**.
4. Fill in the required fields:
    - Item name: enter a descriptive name or the service. For example, `GitHub`.
    - One-time password: Click <i class="material-symbols-outlined">qr_code</i> in the text field to scan the QR code.
5. Click **Save**.
</template>
</tabs>
Once saved, your new authenticator will immediately begin generating codes.

## Use your 2FA generator
 To use it:
1. Sign in to the website with your username and password.
2. When prompted for an authentication code, open Vault to view the current 6-digit code.
3. Enter the code to complete login.
