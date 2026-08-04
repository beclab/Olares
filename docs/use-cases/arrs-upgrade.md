---
outline: deep
description: Upgrade notes for the *Arr media managers when moving to Olares 1.12.6.
head:
  - - meta
    - name: keywords
      content: Olares, *Arrs, Sonarr, Radarr, Prowlarr, upgrade, 1.12.6
app_version: "1.0.x"
doc_version: "1.0"
doc_updated: "2026-07-02"
---

# Upgrade notes for *Arrs

This page describes the configuration changes to make to the *Arr apps after upgrading Olares to V1.12.6.

## Update *Arr app connections

Starting with Olares V1.12.6, the *Arr apps communicate with each other through internal entrance URLs instead of provider URLs. This change simplifies the networking model and makes the media stack more consistent. 

After upgrading to V1.12.6, your existing connections might stop working until you update them to use the internal entrance URLs.

1. In Olares, update each *Arr app to the latest version available for V1.12.6.
2. Open **Settings**, and then go to **Applications** > **[Arr-App-Name]** > **Entrances**.
3. Make sure the **Authentication level** is set to **Internal**.
4. Copy the **Endpoint** URL. This is the internal entrance URL.
5. Use this URL when updating the connection settings in the following sections.

:::tip
The following sections use Sonarr, Prowlarr, and qBittorrent as examples. If you use other *Arr apps (such as Radarr, Lidarr, Readarr, or Bazarr) or other download clients (such as NZBGet, Transmission, or Deluge), follow the same steps to update their connection settings.
:::

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

## Update paths for the unified mount structure

Olares V1.12.6 uses a unified directory mount structure. Update the paths used by your *Arr apps and download clients so they all see the same locations.

### Update Sonarr root folders

It is strongly recommended to update your root folders to the new mount paths.

1. In Sonarr, go to **Settings** > **Media Management** > **Root Folders**.
2. Replace the old root folder path with the new one. For example:
   - Old: `/home/Movies`
   - New: `/olares/userdata/home/Movies/`
3. For each existing series, open the series editor and update the **Path** to the new root folder. When prompted whether to move files, choose **No**.

### Remove remote path mappings

If you previously configured **Remote Path Mappings** between *Arr apps and download clients, you can remove them. The new unified mount structure means the same paths are visible to both *Arr apps and download clients.

## (Optional) Update download clients

If you use qBittorrent, NZBGet, Transmission, or Deluge with your *Arr apps, you may also need to update their default download paths after upgrading to Olares V1.12.6. The new unified directory mount structure means the same paths are visible to both *Arr apps and download clients.

See [Upgrade notes for download clients](./download-clients-upgrade.md) for the detailed steps.
