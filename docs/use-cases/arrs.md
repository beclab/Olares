---
outline: deep
description: Learn how to connect and configure the *Arrs family of applications (Sonarr, Radarr, Prowlarr, Bazarr, and qBittorrent) on Olares for automated media management.
head:
  - - meta
    - name: keywords
      content: Olares, *Arrs, Sonarr, Radarr, Prowlarr, Bazarr, qBittorrent, media server, self-hosted
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-04-21"
---

# Manage your media library with the *Arrs ecosystem

The *Arrs family is a suite of open-source, self-hosted media managers. Sonarr manages TV shows, Radarr manages movies, Lidarr handles music, and Readarr organizes books. Prowlarr manages indexers for these applications, while Bazarr handles subtitle services.

By configuring connections between these tools, they communicate with each other to automatically search, download, and organize your media.

## Learning objectives

In this guide, you will learn how to:
- Locate application provider URLs in Olares.
- Connect a download client (qBittorrent) to a media manager (Sonarr).
- Connect an indexer manager (Prowlarr) to Sonarr.
- Connect a subtitle manager (Bazarr) to Sonarr.

## Prerequisites

This guide focuses specifically on configuring the connections between the *Arrs applications. It does not cover the complete setup or general usage of each individual app. 

Ensure you have properly configured the core settings of your media managers and download clients before connecting them.

## Install the *Arrs applications

Install the *Arrs applications you need for your media stack. This tutorial uses Sonarr, Prowlarr, Bazarr, and qBittorrent.

1. Open Market and search for "Sonarr".

   ![Sonarr app in Market](/images/manual/use-cases/arrs-sonarr.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.
3. Search for "Prowlarr" and install it.

   ![Prowlarr app in Market](/images/manual/use-cases/arrs-prowlarr.png#bordered)

4. Search for "Bazarr" and install it.

   ![Bazarr app in Market](/images/manual/use-cases/arrs-bazarr.png#bordered)

5. Search for "qBittorrent" and install it.

   ![qBittorrent app in Market](/images/manual/use-cases/arrs-qbittorrent.png#bordered)

## Complete the initial app settings

To secure your media server and prevent unauthorized access, you must configure administrator credentials for your applications upon their first launch if prompted. 

The following steps demonstrate the initial setup for Sonarr.

1. Open Sonarr from the Launchpad. The **Authentication Required** page appears.

   ![Sonarr initial settings upon first launch](/images/manual/use-cases/arrs-sonarr-ini-settings.png#bordered)

2. Select an authentication method:
   - **Basic (Browser Popup)**: To use your web browser's native login prompt, select this option. This method is often more compatible with automated password managers.
   - **Forms (Login Page)**: To use Sonarr's built-in, custom login interface for a more visually integrated experience, select this option.

3. Select your security preference from the **Authentication Required** list:

   - **Enabled**: This is the default selection, which requires your username and password regardless of where you access the app from. To ensure maximum security, use this option.
   - **Disabled for Local Addresses**: This option bypasses the login screen when you open Sonarr from a device within your local network. Only select this option if you fully trust all users and devices connected to your local network.

4. In the **Username** field, enter an admin username.
5. In the **Password** field, enter a secure password.
6. In the **Password Confirmation** field, type the same password again.
7. Click **Save**. You log in to Sonarr.

   ![Sonarr landing page](/images/manual/use-cases/arrs-sonarr-landing.png#bordered)

## Locate provider URLs

*Arrs applications use provider URLs to communicate securely within the Olares cluster. 

A provider URL is the HTTP protocol version of an application's entrance endpoint. For example, if the entrance address of an *Arrs application is `https://9691c178.alexmiles.olares.com`, its provider URL is `http://9691c178.alexmiles.olares.com`.

The following steps demonstrate how to locate an application's provider URL in Olares using Sonarr as an example:

1. Open Settings, go to **Applications**, and then select the target application, that is **Sonarr**.
2. Under **Entrances**, click **Sonarr**.

   ![Sonarr entrance in Settings](/images/manual/use-cases/arrs-sonarr-entrance.png#bordered){width=75%}

3. Note down the URL in the **Endpoint** field, that is `https://9691c178.alexmiles.olares.com`.

   ![Sonarr endpoint in Settings](/images/manual/use-cases/arrs-sonarr-endpoint.png#bordered){width=75%}

4. Replace `https` with `http` to construct the provider URL, that is `http://9691c178.alexmiles.olares.com`.

## Connect a download client to a media manager

To download media, your must connect a download client like qBittorrent or Transmission to your media managers (Sonarr, Radarr, Lidarr, and Readarr).

The following steps demonstrate how to connect qBittorrent to Sonarr. You can apply the same process for other media managers.

### Step 1: Locate the provider URL for qBittorrent

1. Open Settings, and then go to **Applications** > **qBittorrent**.

   ![qBittorrent entrance in Settings](/images/manual/use-cases/arrs-qbittorrent-entrance.png#bordered){width=75%}

2. Under **Entrances**, click **qBittorrent**.

   ![qBittorrent endpoint in Settings](/images/manual/use-cases/arrs-qbittorrent-endpoint.png#bordered){width=75%}

3. Note down the URL in the **Endpoint** field. In this case, it is `https://44e535c5.alexmiles.olares.com`.
4. Replace `https` with `http` to construct the provider URL for qBittorrent. In this case, it becomes `http://44e535c5.alexmiles.olares.com`.

### Step 2: Configure the connection in Sonarr

1. Open Sonarr, click **Settings** from the left sidebar, and then select **Download Clients**.

   ![Sonarr Download Clients page](/images/manual/use-cases/arrs-sonarr-download-clients.png#bordered)

2. Click <span class="material-symbols-outlined">add_2</span>, and then scroll down to select **qBittorrent** to add a new client connection.
3. Specify the connection details as follows:

   ![Sonarr Download Clients settings](/images/manual/use-cases/arrs-sonarr-download-clients-settings.png#bordered)

   - **Host**: Enter the qBittorrent provider URL, excluding the `http://` prefix and any trailing slashes. In this case, enter `44e535c5.alexmiles.olares.com`.
   - **Port**: Enter `80`.
   - **Username** and **Password**:
      - If your qBittorrent client requires authentication, enter your username and password.
      - If you use the default qBittorrent settings without a password, leave the two fields blank.
6. Click **Test**. A green checkmark appears, indicating the connection is successful.
7. Select **Save**. qBittorrent appears in the **Download Clients** section as enabled.

   ![Sonarr Download Clients enabled](/images/manual/use-cases/arrs-sonarr-download-clients-enabled.png#bordered)

## Connect an indexer manager to a media manager

To automatically search for media files across multiple indexers (search sites), you must connect an indexer manager like Prowlarr to your media managers.

The following steps demonstrate how to connect Prowlarr to Sonarr. You can apply the same process for other media managers.

### Step 1: Obtain the Sonarr API Key

1. Open Sonarr, click **Settings** from the left sidebar, and then select **General**.
2. In the **Security** section, note down the API Key. In this case, it is `e4ee9f376d754fd3b7146629d737644f`.

   ![Sonarr API key](/images/manual/use-cases/arrs-sonarr-api.png#bordered)

### Step 2: Add indexers in Prowlarr

1. Open Prowlarr and sign in.

   ![Prowlarr landing page](/images/manual/use-cases/arrs-prowlarr-landing.png#bordered)

2. Click **Add New Indexer**.
3. Add your preferred indexers and ensure they connect successfully. For example, to add Uindex:

   a. Search for Uindex, and then click it from the results list.

   ![Prowlarr add an indexer](/images/manual/use-cases/arrs-prowlarr-add-indexer.png#bordered)

   b. Click **Test**. A green checkmark appears, indicating the connection is successful.

   c. Click **Save**.

4. Close the **Add Indexer** window. Prowlarr displays the enabled indexers.

   ![Prowlarr indexers added](/images/manual/use-cases/arrs-prowlarr-indexer-added.png#bordered)

### Step 3: Sync Prowlarr with Sonarr

1. In Prowlarr, click **Settings** from the left sidebar, and then select **Apps**.
2. Click **Apps**.

   ![Prowlarr add apps](/images/manual/use-cases/arrs-prowlarr-add-apps.png#bordered)

3. Click <span class="material-symbols-outlined">add_2</span>, and then select the application you want to connect, that is Sonarr.
4. In the **Add Application - Sonarr** window, specify the following settings:

   - **Prowlarr Server**: Enter the provider URL of Prowlarr, such as `http://e5e5b409.alexmiles.olares.com`.
   - **Sonarr Server**: Enter the provider URL of Sonarr, such as `http://9691c178.alexmiles.olares.com`.
   - **API Key**: Enter the API key you noted down previously.

   ![Prowlarr add apps configuration](/images/manual/use-cases/arrs-prowlarr-add-apps-config.png#bordered)

5. Click **Test**. A green checkmark appears, indicating the connection is successful.
6. Click **Save**. Sonarr appears in the **Applications** section, and Prowlarr automatically pushes the indexers to Sonarr.

   ![Prowlarr apps added](/images/manual/use-cases/arrs-prowlarr-apps-added.png#bordered)

7. To verify the sync, open Sonarr, and then go to **Settings** > **Indexers**. You can see the data sources imported from Prowlarr.

   ![Prowlarr Sonarr sync verify](/images/manual/use-cases/arrs-prowlarr-sync-verify.png#bordered)

   Now when you add a new TV show in Sonarr, it searches these indexers for available files and triggers qBittorrent to download them.

## Connect a subtitle manager to a media manager

To automatically download missing subtitles for your media library, you must connect a subtitle manager like Bazarr to your media managers (Sonarr and Radarr).

The following steps demonstrate how to connect Bazarr to Sonarr.

1. Open Bazarr from the Launchpad, and then click **Sonarr** from the left sidebar.
2. Toggle on **Enabled**, and then specify the following settings:

   - **Address**: Enter the provider URL of Sonarr, excluding the `http://` prefix. For example, `9691c178.alexmiles.olares.com`.
   - **Port**: Enter `80`.
   - **API Key**: Enter the Sonarr API Key you noted down earlier.

3. Click **Test**. Bazarr displays the connected Sonarr version number if the connection is successful.

   ![Bazarr connection success](/images/manual/use-cases/arrs-bazarr-test-connection.png#bordered)

4. Select **Save** in the upper-left corner.

   Bazarr now monitors Sonarr. whenever Sonarr downloads a TV show, Bazarr automatically detects it and downloads the corresponding subtitles according to your language settings.

## FAQ

### Why does my connection test fail?

- **Check the URL format**: Review the requirements for the specific application you configure. Some applications (like Prowlarr connecting to Sonarr) require the `http://` prefix, while others (like Sonarr connecting to qBittorrent, or Bazarr connecting to Sonarr) require to omit the `http://` prefix. 
- **Verify the port**: Ensure you set the port to `80`.

## Learn more

- [Servarr Wiki](https://wiki.servarr.com/)