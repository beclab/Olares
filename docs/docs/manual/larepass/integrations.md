---
outline: [2, 3]
description: Connect Olares with third-party services to enhance functionality. Learn how to integrate, authorize, and manage connected services for seamless data synchronization.
---

# Manage integrations in LarePass

LarePass is your central hub for connecting Olares with third-party services like Google Drive, Dropbox, AWS S3, and Tencent COS. These integrations extend the capabilities of your Olares environment like file sync, secure backup, and more.

:::info
We're working on adding support for more third-party integrations that will connect you with more external services to your Olares account.
:::

## Connect to Olares Space

Olares Space is a cloud hosting service for Olares that shares the same account system with LarePass and Olares.

### Step 1. Log in to Olares Space

1. Open https://space.olares.com/login in your browser.
2. Open LarePass on your mobile device.
3. On the Settings page, tap the "Scan" icon in the top-right corner.
4. Scan the QR code on the Olares Space login page.
5. Confirm the risk prompt and proceed with the login.

### Step 2. Authorize Olares Space

1. In the LarePass app, go to **Settings** > **Integration**.
2. Tap <i class="material-symbols-outlined">add</i> in the top-right corner and select **Space** to add your Olares Space account.

### Step 3. Associate Olares ID
Associating your Olares ID allows you to import a blockchain wallet, which is necessary for using NFT images as unique avatars in your profile.

1. Open the Settings app from the Dock or Launchpad.
2. Select **Integration** from the left sidebar.
3. Click on the Olares Space card on the right to view details.
4. Click **Bind**. This will trigger a confirmation prompt in LarePass app.
5. Open the LarePass app. You should see a confirmation prompt. If not:

   a. Go to **Settings** > **Integration**.

   b. Tap the Olares Space card.

   c. In the confirmation prompt, tap **Confirm** to authorize.
6. Return to Olares, and click **Confirm** to complete the association to your Olares ID.

## Add a cloud drive via OAuth

OAuth-based integrations like Google Drive and Dropbox require initial setup via the LarePass mobile app:

1. Open LarePass on your mobile device.

2. Tap **Settings** > **Integration**, then tap <i class="material-symbols-outlined">add</i> in the top-right corner.

3. Select either Google Drive or Dropbox.

4. Follow the login prompts to authorize your account.

Once authorized, you'll see the connected account in the integration list. You can now access the storage in Files.

## Add a cloud storage using API-keys

Services like AWS S3 and Tencent Cloud COS require setup using API keys (Access Key & Secret Key). You can do this directly from the LarePass app or from the **Integration** settings within Olares:

1. Open LarePass on your mobile device.
2. Tap **Settings** > **Integration**, then tap <i class="material-symbols-outlined">add</i> in the top-right corner.
3. Select AWS S3 or Tencent COS.
4. Enter your Access Key, Secret Key, and other required credentials, then tap **Confirm**.

Once configured, you'll see the connected service in the integration list. And you can access the cloud storage through Files.

Alternatively, you can configure these integrations directly within [Olares Settings](/manual/olares/settings/integrations.md). 


## Disconnect integrations
::: warning
Disconnecting Olares Space may affect your ability to manage devices, and access cloud backups through the Olares Space interface.
You can always reconnect later if needed.
:::

To disconnect an integration from LarePass:

1. Open LarePass app, and go to **Settings** > **Integration**.
2. Tap on the integration you wish to remove.
3. Tap <i class="material-symbols-outlined">more_horiz</i> in the top-right corner, and tap **Delete**.
