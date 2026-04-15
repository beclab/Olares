---
outline: [2, 3]
description: Configure advanced settings in TREK on Olares, including OIDC single sign-on, Google API keys, two-factor authentication, and data backups.
head:
  - - meta
    - name: keywords
      content: Olares, TREK, SSO, OIDC, Google API, 2FA, backup, restore, configuration, administration, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-15"
---

# Configure advanced settings in TREK

As your TREK workspace grows, you might want to integrate third-party services, strengthen your account security, and safeguard your data. Use TREK's administrative settings to configure single sign-on, improve map features, enable two-factor authentication, and set up automated backups.

## Learning objectives

In this guide, you will learn how to:
- Set up OIDC single sign-on (SSO) for third-party logins.
- Add Google API keys to improve map search results.
- Secure your account with two-factor authentication (2FA).
- Back up and restore your travel data.

## Set up OIDC single sign-on

TREK supports third-party login through Google, Apple, Authentik, Keycloak, or any OIDC provider. The following example uses Google.

1. Go to the Google Cloud Console, and create an OAuth client. Specify the following:
   - **Authorized JavaScript origins**: `https://<your-trek-domain>`
   - **Authorized redirect URIs**: `https://<your-trek-domain>/oidc/callback`

   ![Google OAuth client](/images/manual/use-cases/trek-google-oauth.png#bordered)

2. After creating the client, copy the **Client ID** and **Client Secret**.
3. In TREK, go to **Admin** > **Configuration** > **OIDC**, and enter the Client ID and Client Secret. Select **Save**.

   ![Paste OIDC credentials](/images/manual/use-cases/trek-oidc-config.png#bordered)

4. Log out. On the login page, sign in with your Google account.

   ![Google login](/images/manual/use-cases/trek-google-login.png#bordered)

## Improve map search with Google API keys

Adding a Google API key enables place photos, ratings, and opening hours when you search for locations in your itinerary.

1. Go to the Google Cloud Console, and create an API key.
2. In TREK, go to **Admin** > **Settings** > **API Keys**, enter your Google API key, and select **Save**.

   ![API key settings](/images/manual/use-cases/trek-api-key.png#bordered)

3. Places you add display photos, ratings, and opening hours.

   ![Place details](/images/manual/use-cases/trek-place-details.png#bordered)

## Secure your account with two-factor authentication (2FA)

TREK supports TOTP-based two-factor authentication (2FA) with apps like Google Authenticator or Authy.

### Enable 2FA

1. Go to **Settings** > **Two-factor authentication (2FA)** > **Set up authentication**.
2. Scan the QR code with your authenticator app, and enter the generated code.
3. Select **Enable 2FA**.

   ![2FA setup](/images/manual/use-cases/trek-2fa-setup.png#bordered)

4. Save the backup codes displayed on screen. Store them in a safe place, and then select **OK**.

   ![Backup codes](/images/manual/use-cases/trek-2fa-backup-codes.png#bordered)

### Disable 2FA

1. Go to **Settings** > **Two-factor authentication (2FA)**.
2. Enter your current password and a 2FA code from your authenticator app.
3. Select **Disable 2FA**.

   ![Disable 2FA](/images/manual/use-cases/trek-2fa-disable.png#bordered)

## Back up and restore your data

TREK supports both manual and automatic backups to ensure your travel data is safe.

- **Manual backup**: Go to **Admin** > **Backups** to create a new backup or upload an existing backup file.

  ![Manual backup](/images/manual/use-cases/trek-backup-manual.png#bordered)

- **Automatic backup**: Configure scheduled backups under **Admin** > **Backups** > **Auto Backup** settings.

  ![Auto backup](/images/manual/use-cases/trek-backup-auto.png#bordered)
