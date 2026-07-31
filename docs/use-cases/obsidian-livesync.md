---
outline: [2, 3]
description: Use Obsidian LiveSync on Olares to host a private CouchDB sync backend and keep an Obsidian vault synced across your computer and phone.
head:
  - - meta
    - name: keywords
      content: Olares, Obsidian LiveSync, Obsidian, Self-hosted LiveSync, CouchDB, Markdown notes, knowledge management, vault sync, cross-device sync
app_version: "1.0.16"
doc_version: "1.0"
doc_updated: "2026-07-24"
---

# Sync Obsidian notes with Obsidian LiveSync

Obsidian is a flexible Markdown note-taking app for private notes, linked knowledge bases, journals, and project work. Obsidian LiveSync on Olares provides a self-hosted CouchDB backend for the Self-hosted LiveSync community plugin, so your vault can sync across devices without relying on Obsidian Sync or another cloud service.

This guide uses a computer as the primary device and a phone as the second device.

## Learning objectives

In this guide, you will learn how to:

- Install Obsidian LiveSync on Olares.
- Create a CouchDB database for your Obsidian vault.
- Connect the Self-hosted LiveSync plugin on your primary Obsidian device.
- Import the same sync setup on another device.
- Enable LiveSync mode and check cross-device sync.

## Prerequisites

- Obsidian installed on your computer and phone. Download it from the [official website](https://obsidian.md/download).
- A local folder where you want to store the primary Obsidian vault.
- LarePass installed on your computer and phone.

## Install Obsidian LiveSync

1. Open Market and search for "Obsidian LiveSync".

   ![Obsidian LiveSync in Market](/images/manual/use-cases/obsidian-livesync.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure Obsidian LiveSync on Olares

Before configuring your Obsidian clients, create a database in Obsidian LiveSync and allow the Self-hosted LiveSync plugin to reach it.

### Create a database

1. Open Obsidian LiveSync from Launchpad.

2. Sign in with the default credentials:

   - **Username**: `admin`
   - **Password**: `password`

   ![Sign in to Obsidian LiveSync](/images/manual/use-cases/obsidian-livesync-sign-in.png#bordered)

3. (Optional) Change the default password:

   a. In the left sidebar, click the user management icon at the bottom.

   ![Change password in Obsidian LiveSync](/images/manual/use-cases/obsidian-change-password.png#bordered)

   b. On the **Change Password** tab, enter and verify a new password, then click **Change**.

4. In the left sidebar, click <i class="material-symbols-outlined">database</i> icon.

5. Click **Create Database** in the upper-right corner.

6. Enter a name to create the database.

   :::warning Database naming rules
   Use only lowercase letters, numbers, and hyphens. Self-hosted LiveSync relies on CouchDB, which strictly requires lowercase names. Using uppercase letters will cause the sync to fail silently later on.
   :::

   In this example, we create a database named `olares`.

   ![Create Obsidian LiveSync database](/images/manual/use-cases/obsidian-livesync-create-database.png#bordered)

After the database is created, Obsidian LiveSync redirects you to the database configuration page.

### Allow plugin access

The Self-hosted LiveSync plugin connects directly to the CouchDB endpoint hosted by Obsidian LiveSync. Update the app entrance so the plugin can connect through LarePass VPN.

1. Open Olares Settings, and then go to **Applications** > **Obsidian LiveSync** > **Entrances**.

2. Set **Authentication level** to **Internal**, and then click **Submit**.

   ![Set Obsidian LiveSync authentication level](/images/manual/use-cases/obsidian-livesync-authentication-level.png#bordered)

   :::warning Public access risk
   Setting **Authentication level** to **Public** can also make the connection work, but anyone who gets the endpoint can try to connect to CouchDB. Keep it set to **Internal** unless you have a specific reason to expose the endpoint.
   :::

3. Copy the **Endpoint** URL from the same entrance page. You will use it as the CouchDB URL in Obsidian.

   The endpoint usually looks like this:

   ```text
   https://8591294e.{username}.olares.com
   ```

## Configure the primary device

Use your computer as the primary device. This device initializes the remote database, so start from the vault that you want to use as the source of truth.

### Prepare Obsidian

:::info LarePass VPN required
Before configuring desktop Obsidian, enable LarePass VPN in the LarePass desktop app.
:::

1. Open Obsidian on your computer.
2. Create or open the vault you want to sync. This guide creates a new vault named `Olares` from the vault selection screen:

   a. Next to **Create new vault**, click **Create**.

   ![Create a vault](/images/manual/use-cases/obsidian-vault-selection.png#bordered)

   b. In **Vault name**, enter `Olares`.

   c. Click **Browse**, and then choose the local folder where you want to store the vault.

   d. Click **Create**.

3. In Obsidian, go to **Settings** > **Community plugins**, then click **Browse**.

   ![Browse plugin](/images/manual/use-cases/obsidian-livesync-plugin-browse.png#bordered)

4. Search for "Self-hosted LiveSync", then click **Install**.
5. Click **Enable** to activate the plugin.

   ![Install Self-hosted LiveSync on desktop](/images/manual/use-cases/obsidian-livesync-enable-plugin.png#bordered)

6. When prompted, click **No, please take me back** to leave the plugin setup guide.

### Connect to CouchDB

1. Go to **Settings** > **Self-hosted LiveSync**.
2. Click the remote configuration icon. It is the fourth icon from the left at the top of the plugin settings page.
3. Under **E2EE Configuration**, click **Configure And Change Remote**.

   ![Configure and change remote](/images/manual/use-cases/obsidian-livesync-configure-change-remote.png#bordered)

4. In **End-to-End Encryption**, choose whether to enable encryption:

   - To enable it, select **End-to-End Encryption**, enter a passphrase, and keep the passphrase somewhere safe.
   - To skip it, leave **End-to-End Encryption** cleared.
   - Optional: Select **Obfuscate Properties** if you also want to hide file metadata, such as paths, sizes, and timestamps, from the remote server.

5. Click **Proceed**.
6. In **Enter Server Information**, select **CouchDB**, and then click **Continue to CouchDB setup**.

   ![Select CouchDB as the sync server](/images/manual/use-cases/obsidian-livesync-select-couchdb.png#bordered){width=70%}

7. In **CouchDB Configuration**, enter the database connection details:

   - **URL**: Enter the Obsidian LiveSync endpoint URL from **Settings** > **Applications** > **Obsidian LiveSync** > **Entrances**. For example:
     ```text
     https://8591294e.{username}.olares.com
     ```
   - **Username**: Enter the Obsidian LiveSync username. If you kept the default, enter `admin`.
   - **Password**: Enter the Obsidian LiveSync password.
   - **Database Name**: Enter the database name you created earlier. In this example, it's `olares`.

      ![Configure CouchDB in Self-hosted LiveSync](/images/manual/use-cases/obsidian-livesync-configure-couchdb.png#bordered){width=70%}

8. Click **Test Settings and Continue**. If the connection is configured correctly, Obsidian continues to the next setup step.

:::tip Configure encryption later
If you skipped end-to-end encryption during the remote setup wizard, you can configure it later. Open **Settings** > **Self-hosted LiveSync**, click the remote configuration icon, and then click **Configure** under **E2EE Configuration**.

In the **End-to-End Encryption** dialog, select **End-to-End Encryption**, enter a passphrase, optionally select **Obfuscate Properties**, and then save the configuration.

Use the same end-to-end encryption setting and passphrase on every device connected to the same sync target. A mismatched passphrase can make synced data unreadable.
:::

### Initialize the server

:::warning Data overwrite
The initialization steps overwrite existing data on the remote server and rebuild the server from the current device. Use them only for a new Obsidian LiveSync database or when you intentionally want the current device to become the source of truth. Before proceeding, copy your Obsidian vault folder to a safe location.
:::

1. When **Mostly Complete: Decision Required** appears, select **I am setting up a new server for the first time / I want to reset my existing server**, and then click **Proceed to the next step**.

   In this guide, the remote database is new, so the primary device should initialize the server with the current vault data.

   ![Initialize the remote server from the primary device](/images/manual/use-cases/obsidian-livesync-initialize-server.png#bordered){width=70%}

2. When **Setup Complete: Preparing to Initialise Server** appears, click **Restart and Initialise Server**.

   Obsidian restarts and uploads the current vault data from this device to the server as the master copy.

   ![Restart and initialize the LiveSync server](/images/manual/use-cases/obsidian-livesync-restart-initialise-server.png#bordered){width=70%}

3. When **Final Confirmation: Overwrite Server Data with This Device's Files** appears, confirm that the current device should overwrite the server:

   a. Read the warning carefully.

   b. Select all three confirmation checkboxes.

   c. Under **Have you created a backup before proceeding?**, select **I have created a backup of my Vault**.

   d. Click **I Understand, Overwrite Server**.

   ![Confirm overwriting the LiveSync server](/images/manual/use-cases/obsidian-livesync-overwrite-server-confirmation.png#bordered){width=70%}

### Enable LiveSync mode

After Obsidian restarts and finishes the server initialization, return to the Self-hosted LiveSync settings page.

1. Open **Settings** > **Self-hosted LiveSync**.
2. Click the sync configuration icon. It is the fifth icon from the left at the top of the plugin settings page.
3. Set **Sync Mode** to **LiveSync**.

   ![Enable LiveSync](/images/manual/use-cases/obsidian-livesync-enable-livesync.png#bordered){width=70%}

:::info Keep your encryption passphrase
If you enable end-to-end encryption, save the passphrase somewhere safe. You need the same passphrase when adding another device.
:::

## Configure another device

Use the setup URI from the primary device to import the same LiveSync configuration on your phone.

### Prepare the phone

:::info LarePass VPN required
Before configuring mobile Obsidian, enable LarePass VPN in the LarePass mobile app.
:::

1. Open Obsidian on your phone, and create a new local vault for synced notes.
2. Install and enable Self-hosted LiveSync in **Settings** > **Community plugins**. The process is the same as on the primary device.

### Copy the setup URI from the primary device

1. On your computer, open **Settings** > **Self-hosted LiveSync**.
2. Click the quick setup icon. It is the second icon from the left at the top of the plugin settings page.
3. In **To setup other devices**, click **Copy**.
4. Enter a password to encrypt the setup URI, and then click **Ok**. Obsidian copies the setup URI to your clipboard.

   ![Copy setup URI from primary device](/images/manual/use-cases/obsidian-livesync-copy-setup-uri.png#bordered)

5. Send the setup URI to your phone through a secure channel.

### Join the existing server

1. On your phone, open **Settings** > **Self-hosted LiveSync**.
2. Click the quick setup icon.
3. Next to **Connect with Setup URI**, tap **Use**.
4. Paste the setup URI, enter the password you used to encrypt it, and then confirm.
5. If Obsidian asks how to handle the remote server, select **My remote server is already set up. I want to join this device**.

   This lets the second device fetch the existing synchronization data from the server instead of overwriting it.

6. Open **Settings** > **Self-hosted LiveSync**.
7. Click the sync configuration icon.
8. Set **Sync Mode** to **LiveSync**.

When both devices use LiveSync mode, changes made in one vault sync to the other device in real time.

## Check sync results

1. Create or edit a note on the primary device.
2. Wait a few moments.
3. Open the same vault on your phone. The note should appear with the latest changes.
4. Edit the note on your phone, and then check that the change appears on your computer. You can view the sync status in the upper-right status bar.

   ![Check Obsidian LiveSync result](/images/manual/use-cases/obsidian-livesync-sync-result.png#bordered)

:::tip First sync
The first sync can take longer if the vault contains many notes or attachments. Keep both devices online until the first sync finishes.
:::

## Learn more

- [Obsidian Help](https://help.obsidian.md): Official guides for Obsidian features and settings.
- [Self-hosted LiveSync repository](https://github.com/vrtmrz/obsidian-livesync): Plugin documentation and troubleshooting notes.
