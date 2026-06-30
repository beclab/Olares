---
outline: deep
description: Upgrade notes for the *Arr media managers and download clients when moving to Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, *Arrs, Sonarr, Radarr, Prowlarr, qBittorrent, NZBGet, Transmission, Deluge, upgrade, 1.12.6
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-06-30"
---

# Upgrade notes for *Arrs

This page describes the configuration changes to make to the *Arr apps and download clients after upgrading Olares to V1.12.6.

## Update *Arr app connections

Starting with Olares V1.12.6, the *Arr apps communicate with each other through internal entrance URLs instead of provider URLs. This change simplifies the networking model and makes the media stack more consistent. After upgrading to V1.12.6, your existing connections might stop working until you update them to use the internal entrance URLs.

1. In Olares, update each *Arr app to the latest version available for V1.12.6.
2. Open **Settings**, and then go to **Applications** > **App Name** > **Entrances**. Replace **App Name** with the app you are configuring, such as Sonarr or Prowlarr.
3. Make sure the **Authentication level** is set to **Internal**.
4. Copy the **Endpoint** URL. This is the internal entrance URL.
5. Use this URL when updating the connection settings in the following sections.

### Update the server URLs for Prowlarr-Sonarr sync

1. In Prowlarr, go to **Settings** > **Apps**, and then select **Sonarr**.
2. Change both **Prowlarr Server** and **Sonarr Server** from `http://` to `https://`. For example:
   - **Prowlarr Server**: `https://e5e5b409.alexmiles.olares.com`
   - **Sonarr Server**: `https://9691c178.alexmiles.olares.com`
3. Click **Test**. A green checkmark indicates the connection is successful.
4. Click **Save**.

### Update the download client connection in Sonarr

1. In Sonarr, go to **Settings** > **Download Clients**, and then select **qBittorrent**.
2. Specify the connection details as follows:
   - **Host**: Enter the internal entrance URL of qBittorrent.
   - **Port**: Enter `443`.
   - **Use SSL**: Enable this option.
3. Click **Test**, and then click **Save**.

### Update Sonarr root folders

Olares V1.12.6 uses a unified directory mount structure. It is strongly recommended to update your root folders to the new mount paths.

1. In Sonarr, go to **Settings** > **Media Management** > **Root Folders**.
2. Replace the old root folder path with the new one. For example:
   - Old: `/home/Movies`
   - New: `/olares/userdata/home/Movies/`
3. For each existing series, open the series editor and update the **Path** to the new root folder. When prompted whether to move files, choose **No**.

### Remove remote path mappings

If you previously configured **Remote Path Mappings** between *Arr apps and download clients, you can remove them. The new unified mount structure means the same paths are visible to both *Arr apps and download clients.

## (Optional) Update download clients

Olares V1.12.6 uses a unified directory mount structure. Update the default download paths in your download clients so that *Arr apps can still access the downloaded files.

### qBittorrent

1. Open qBittorrent, on the toolbar, select **Tools** > **Options**.
2. In the **Options** window, go to the **Downloads** tab.
3. Update **Default Save Path** from `/downloads/home/Downloads/qBittorrent` to `/olares/userdata/home/Downloads/qBittorrent`.

   ![qBittorrent default save path](/images/manual/use-cases/update-default-save-path.png#bordered)
   
4. Scroll down and click **Save**.

#### FAQ: qBittorrent asks me to log in after the upgrade

If qBittorrent previously allowed you to access the WebUI without logging in, but now prompts for credentials, the migration script might not have preserved the authentication settings.

1. Open the Control Hub and view the **qBittorrent** container logs to find the temporary username (usually `admin`) and password.
2. Log in to qBittorrent and verify it works.
3. Set a new username and password as needed.
4. To restore passwordless access:

   a. Open Olares Files, and then edit `/Data/qbittorrent/qBittorrent.conf`.

   b. In the `[Preferences]` section, add or update the following values:

      ```ini
      WebUI\Address=*
      WebUI\ServerDomains=*
      WebUI\Port=8080
      WebUI\CSRFProtection=false
      WebUI\HostHeaderValidation=false
      WebUI\LocalHostAuth=false
      WebUI\AuthSubnetWhitelistEnabled=true
      WebUI\AuthSubnetWhitelist=0.0.0.0/0, ::/0
      ```

   c. Save the file and restart qBittorrent.

If the issue persists, back up the existing configuration, delete `qBittorrent.conf`, and restart qBittorrent. The initialization process will recreate a default configuration with passwordless access enabled. You can then re-apply your previous settings line by line to identify any conflicting option.

### NZBGet

1. Open NZBGet and go to **Settings** > **PATHS**.
2. Update the following paths:
   - **DestDir**: Change `/downloads/completed` to `/olares/userdata/home/Downloads/nzbget/completed`
   - **InterDir**: Change `/downloads/intermediate` to `/olares/userdata/home/Downloads/nzbget/intermediate`
3. Save the settings and restart NZBGet.

### Transmission

1. Open Transmission, click <i class="material-symbols-outlined">menu</i> in the top right, and then select **Edit preferences**.
2. Update the following paths:
   - **Download to**: Change `/downloads/complete` to `/olares/userdata/home/Downloads/transmission/complete`
   - **Use temporary folder**: Change `/downloads/incomplete` to `/olares/userdata/home/Downloads/transmission/incomplete`

### Deluge

1. Open Deluge and go to **Preferences** > **Downloads**.
2. Update the download path from `/downloads` to `/olares/userdata/home/Downloads/deluge`.
3. Click **Apply** and then click **OK**.
