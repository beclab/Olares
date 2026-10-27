---
outline: [2, 3]
description: Use Vault as an authenticator app to generate Time-Based One-Time Password (TOTP) verification codes and strengthen account security for your online services.
head:
  - - meta
    - name: keywords
      content: Olares, Vault, two-factor authentication, TOTP, authenticator app, 2FA
---
# Generate two-factor authentication codes

Two-factor authentication (2FA) adds an extra layer of account security. After you enter your password, the service asks for a verification code.

Vault can work as an authenticator app. It generates Time-Based One-Time Password (TOTP) codes that refresh automatically, just like Google Authenticator or Microsoft Authenticator.

This guide shows you how to set up Vault as your authentication method and generate 2FA codes for websites such as GitHub or OpenAI.

## Prerequisites

- You have installed LarePass on your device. Visit the [official page](https://www.olares.com/larepass) to download it.
- You have access to the security settings of the online service where you want to enable 2FA.

## Prepare your target service

1. Sign in to the website where you want to enable 2FA, such as GitHub or OpenAI.
2. Go to the security settings page and enable two-factor authentication with an authenticator app.

   ![Enable GitHub 2FA](/images/manual/olares/2fa-github.png#bordered)

3. Save the secret key or QR code that the service provides. You will need it in the next section.

:::tip
If the service gives you recovery codes, store them in a safe place. You need them to recover your account if you lose access to Vault.
:::

## Create an authenticator in Vault

<tabs>
<template #Olares,-LarePass-desktop>

1. Open **Vault**.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. Select **Authenticator** as the item type, then click **Create**.
4. Fill in the required fields:
    - **Item name**: Enter a name that helps you identify the service. For example, `GitHub`.
    - **One-time password**: Paste the secret key.
5. Click **Save**.

</template>

<template #LarePass-mobile>

1. Open LarePass on your device, and go to the **Vault** page.
2. Click <i class="material-symbols-outlined">add</i> in the top right corner.
3. Select **Authenticator** as the item type, then click **Create**.
4. Fill in the required fields:
    - **Item name**: Enter a name that helps you identify the service. For example, `GitHub`.
    - **One-time password**: Click <i class="material-symbols-outlined">qr_code</i> in the text field to scan the QR code.
5. Click **Save**.

</template>
</tabs>

Once saved, Vault starts generating verification codes for the account.

## Use your 2FA codes

1. Sign in to the website with your username and password.
2. When the site asks for a verification code, open Vault and find the current 6-digit code.
3. Enter the code to complete sign-in.
