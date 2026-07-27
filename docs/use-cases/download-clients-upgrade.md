---
outline: deep
description: Upgrade notes for the qBittorrent, NZBGet, Transmission, and Deluge download clients when moving to Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, download clients, qBittorrent, NZBGet, Transmission, Deluge, upgrade, 1.12.6
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-07-02"
---

# Upgrade notes for download clients

This page describes how to update the default download paths in your download clients after upgrading Olares to V1.12.6. The new unified directory mount structure means the same paths are visible to both *Arr apps and download clients, so you might need to update your existing download locations.

If you also use *Arr apps such as Sonarr or Radarr, see [Upgrade notes for *Arrs](./arrs-upgrade.md) for the app connection and root folder changes.

## qBittorrent

1. Open qBittorrent.
2. On the toolbar, select **Tools** > **Options**.
3. In the **Options** window, go to the **Downloads** tab.
4. Update **Default Save Path** from `/downloads/home/Downloads/qBittorrent` to `/olares/userdata/home/Downloads/qBittorrent`.

   ![qBittorrent default save path](/images/manual/use-cases/update-default-save-path.png#bordered)
   
5. Scroll down and click **Save**.

### qBittorrent asks me to log in after the upgrade

If qBittorrent previously allowed you to access the WebUI without logging in, but now prompts for credentials, the migration script might not have preserved the authentication settings.

1. Open Control Hub and view the **qBittorrent** container logs to find the temporary username (usually `admin`) and password.
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

## NZBGet

1. Open NZBGet and go to **Settings** > **PATHS**.
2. Update the following paths:
   - **DestDir**: Change `/downloads/completed` to `/olares/userdata/home/Downloads/nzbget/completed`
   - **InterDir**: Change `/downloads/intermediate` to `/olares/userdata/home/Downloads/nzbget/intermediate`
3. Save the settings and restart NZBGet.

## Transmission

1. Open Transmission, click <i class="material-symbols-outlined">menu</i> in the top right, and then select **Edit preferences**.
2. Update the following paths:
   - **Download to**: Change `/downloads/complete` to `/olares/userdata/home/Downloads/transmission/complete`
   - **Use temporary folder**: Change `/downloads/incomplete` to `/olares/userdata/home/Downloads/transmission/incomplete`

## Deluge

1. Open Deluge and go to **Preferences** > **Downloads**.
2. Update the download path from `/downloads` to `/olares/userdata/home/Downloads/deluge`.
3. Click **Apply** and then click **OK**.
