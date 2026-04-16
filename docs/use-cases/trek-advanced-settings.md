---
outline: deep
description: Configure advanced settings in TREK on Olares, including OIDC single sign-on, Google API keys, two-factor authentication, and data backups.
head:
  - - meta
    - name: keywords
      content: Olares, TREK, SSO, OIDC, Google API, 2FA, backup, restore, configuration, administration, self-hosted
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-16"
---

# Configure advanced settings in TREK

As your TREK workspace grows, you might want to integrate third-party services, strengthen your account security, and safeguard your data.

Use TREK's administrative settings to configure single sign-on (SSO), improve map features, enable two-factor authentication (2FA), and set up backups.

## Learning objectives

In this guide, you will learn how to:
- Set up third-party single sign-on to streamline user logins.
- Connect Google API keys to improve map search results.
- Secure user accounts with two-factor authentication.
- Back up and restore your workspace data.

## Set up third-party single sign-on

Simplify the login experience for your travel companions by allowing them to sign in using their existing Google, Apple, or other OIDC provider accounts. 

The following workflow demonstrates how to set up Google SSO.

### Step 1: Set up a Google Cloud project

Prepare a dedicated project in Google Cloud to manage your TREK authentication credentials.

1. Sign in to the [Google Cloud console](https://console.cloud.google.com/auth/clients) using your Google account.
2. Create a new project, or select an existing project from the top navigation bar.

### Step 2: Configure the OAuth consent screen

Define the authorization screen that users will see when they select to sign in with Google.

1. In the left sidebar, hover over **APIs & Services**, and then select **OAuth consent screen**.

   ![Select OAuth consent screen](/images/manual/use-cases/trek-oauth-consent-menu.png#bordered)

2. On the **OAuth Overview** page, click **Get started**.
3. Complete the following project configurations, and then click **Create**:
   - **App Information**: Enter an app name such as `TREK`, select a user support email, and then click **Next**.
   - **Audience**: Select **External** as your target audience type, and then click **Next**.
   - **Contact Information**: Enter an email for receiving project updates, and then click **Next**.
   - **Finish**: Select **Agree to the Google API Services: User Data Policy**, and then click **Continue**.

### Step 3: Create the OAuth client ID

Generate the credentials that TREK needs to communicate with Google.

1. In the **Metrics** section, click **Create OAuth client**.

   ![Create OAuth client](/images/manual/use-cases/trek-create-oauth-client.png#bordered)

2. Specify the following settings, and then click **Create**:
   - **Application type**: Select **Web application**.
   - **Name**: Enter a name for the OAuth 2.0 client for easy identification.
   - **Authorized JavaScript origins**: Click **ADD URI**, and then enter your TREK domain. 
   
      For example, `https://8eb06391.alexmiles.olares.com`.
   - **Authorized redirect URIs**: Click **ADD URI**, and then enter your callback URL in the format of `https://<your-trek-domain>/api/auth/oidc/callback`. 
   
      For example, `https://8eb06391.alexmiles.olares.com/api/auth/oidc/callback`.

      ![Google OAuth client](/images/manual/use-cases/trek-google-oauth1.png#bordered)

3. In the **OAuth client created** window, copy the **Client ID** and **Client secret**, and then click **OK**. The new client ID appears on the **OAuth 2.0 Client IDs** page.

   ![OAuth 2.0 Client IDs](/images/manual/use-cases/trek-google-oauth-id.png#bordered)

### Step 4: Connect Google SSO to TREK

Integrate the credentials you generated in Google Cloud with your TREK admin settings.

1. Go back to TREK, click your user avatar, and then select **Admin**.
2. Click the **Settings** tab, and then locate the **Single Sign-On (OIDC)** panel.
3. Specify the following settings, and then click **Save**:
   - **Display Name**: Enter `Google`.
   - **Issuer URL**: Enter `https://accounts.google.com`.
   - **Client ID**: Enter the client ID you copied.
   - **Client Secret**: Enter the client secret you copied.

   ![Paste OIDC credentials](/images/manual/use-cases/trek-oidc-config.png#bordered)

4. Log out of TREK. 
5. On the **Sign In** page, select **Sign in with Google**.

   ![TREK sign in with Google account](/images/manual/use-cases/trek-google-login.png#bordered)

6. Select your Google account to log in.
7. Click **Continue** when prompted to **Sign in to olares.com**. You are now logged in to TREK.

## Improve map search with Google API keys

By default, TREK uses basic maps. To display rich place details such as photos, ratings, and opening hours, connect a Google Places API key to your workspace.

1. Ensure you have created an API key in Google Cloud console. For more information, see [Create an API key](https://docs.cloud.google.com/docs/authentication/api-keys#create).

   :::tip Required Google APIs
   For the best search experience, ensure that these APIs are enabled: Directions API, Geocoding API, Geolocation API, Maps Elevation API, Maps Embed API, Maps JavaScript API, Maps SDK for Android, Places API, Places API (New), Roads API, and Time Zone API.
   :::

2. Log in to TREK, click your user avatar, and then select **Admin**.
3. Click the **Settings** tab, and then locate the **API Keys** panel.

   ![API key settings](/images/manual/use-cases/trek-api-key-settings.png#bordered)

4. Under **Google Maps API Key**, enter your API key, and then click **Test**. The **Connected** status is displayed, indicating that the key is valid.
5. Click **Save**.
6. Open your trip and add a new place. The results now display additional context, such as photos, ratings, and opening hours.

   ![Place details](/images/manual/use-cases/trek-place-details.png#bordered)
<!--
1. Sign in to the [Google Cloud Console](https://console.cloud.google.com/auth/clients) using your Google account.
2. Create a project or select an existing one.

### Step 2: Enable search-specific APIs

A standard map key only shows the map. To improve search, you must Enable the following APIs:
- Directions API
- Geocoding API
- Geolocation API
- Maps Elevation API
- Maps Embed API
- Maps JavaScript API
- Maps SDK for Android
- Places API
- Places API (New)
- Roads API
- Time Zone API

1. In the left sidebar, hover over **APIs & Services**, and then select **Library**.
2. Search for the API, click it, and then click **Enable**.

   ![Enable Places API key](/images/manual/use-cases/trek-enable-mapapi.png#bordered)

3. Repeat the same steps to enable all the above APIs.

### Step 3: Create API keys

1. In the left sidebar, hover over **APIs & Services**, and then select **Credentials**.
2. Click **Create credentials** at the top of the page, and then select **API key**.

   ![API key menu](/images/manual/use-cases/trek-create-api-key.png#bordered)

3. In the **Create API key** panel, configure the following settings:

   - **Name**: Enter a name to identify the API key. 
   - **Select API restrictions**: Select the APIs you just enabled, and then click **OK**.
   - **Authenticate API calls through a service account**: Do not select.
   - **Application restrictions**: Select **None**.

   ![Create API key settings in google cloud console](/images/manual/use-cases/trek-create-api-key-settings.png#bordered)

4. Click **Create**. 
5. In the **API key created** panel, note down your API key.
-->

## Secure your account with two-factor authentication

Protect your travel data by adding a second layer of security to your login process. TREK supports TOTP-based two-factor authentication (2FA) using apps like Google Authenticator or Authy.

### Enable 2FA

You can choose to secure only your personal account, or mandate 2FA for your entire workspace. You must enable 2FA on your own admin account before you can require it for all your workspace members.

#### Enable 2FA for your admin account

Set up an authenticator app to require a 6-digit code alongside your password.

1. Log in to TREK, click your user avatar, and then select **Settings**.
2. Click the **Account** tab, and then locate the **Two-factor authentication (2FA)** section.
3. Click **Set up authenticator**.

   ![Enable 2FA for admin account](/images/manual/use-cases/trek-2fa-admin.png#bordered)

4. Scan the QR code with your authenticator app, enter the generated 6-digit code in TREK, and then click **Enable 2FA**.
5. Save the backup codes displayed on screen, store them in a safe place, and then click **OK**.

   ![Note down backup codes](/images/manual/use-cases/trek-2fa-backup-codes.png#bordered)

6. Log out of TREK, and then log in again.
7. Enter your credentials, and then click **Sign in**.
8. Enter the verification code from your authenticator app, and then click **Verify**.

   ![2FA authentication upon login](/images/manual/use-cases/trek-2fa-login.png#bordered)

#### Enable 2FA for members

After you enable 2FA on your admin account, you can enforce a mandatory 2FA policy for all members who access your TREK workspace.

1. In TREK, click the user avatar, and then select **Admin**.
2. Click the **Settings** tab, locate the **Require two-factor authentication (2FA)** panel, and then toggle it on.

   ![Enable 2FA](/images/manual/use-cases/trek-2fa-enable.png#bordered)

   When workspace members log in, TREK directs them to the **Settings** page with the message: `Your administrator requires two-factor authentication. Set up an authenticator app below before continuing`. They must click **Set up authenticator**, scan the QR code, and configure their authenticator app before they can view or edit any trips.

   ![Member enable 2FA notification](/images/manual/use-cases/trek-2fa-member-enable.png#bordered)

### Disable 2FA

Removing 2FA is a two-part process if it was enforced globally. Turning off the workspace requirement does not automatically delete members' 2FA configurations. Each user must still manually disable it on their own account.

#### Disable the workspace-wide 2FA requirement

As an admin, if you enforced 2FA globally, you must turn off this requirement before anyone (including yourself) can disable personal 2FA.

1. Log in to TREK, select your user avatar, and then select **Admin**.
2. Select the **Settings** tab, locate the **Require two-factor authentication (2FA)** section, and then toggle it off.

      ![Disable 2FA for all users](/images/manual/use-cases/trek-2fs-disable-all.png#bordered)

#### Disable 2FA for your personal account

Remove the 2FA step from your login process.

:::info Workspace restrictions
If you are a workspace member, you cannot disable your personal 2FA until your [admin turns off the global requirement](#disable-the-workspace-wide-2fa-requirement).
:::

1. Log in to TREK using your password and current 2FA code.
2. Select your user avatar, and then select **Settings**.
3. Select the **Account** tab, and then locate the **Two-factor authentication (2FA)** section.
4. Enter your login password.
5. Enter the verification code from your authenticator app.
6. Select **Disable 2FA**.

      ![Disable 2FA](/images/manual/use-cases/trek-2fa-disable.png#bordered)

7. Log out of TREK, and then log in again. The verification step is no longer required.

## Back up and restore your data

Regularly save your itineraries, notes, and budgets to prevent data loss. TREK allows you to create manual backups or schedule automated ones.

### Back up your workspace data

Choose to trigger a backup immediately or set up a recurring schedule.

1. Create a backup:

   - **Manual backup**: Click user avatar, select **Admin**, go to the **Backup** tab, and then click **Create Backup**.

      ![Manual backup](/images/manual/use-cases/trek-backup-manual.png#bordered)

   - **Auto backup**: Configure scheduled backups under **Admin** > **Backups** > **Auto Backup** settings.

      ![Automatic backup](/images/manual/use-cases/trek-backup-auto.png#bordered)
2. Click **Download** to save your backup package locally.

### Restore data from a backup

If you need to recover lost information, you can upload a previously saved backup file.

1. On the **Backup** page, click **Upload Backup**, and then select your local package.
2. When prompted, review the **Restore Backup** warning, and then confirm your restore to overwrite the current workspace with the backup data.
