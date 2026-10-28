---
outline: [2, 3]
description: Store photos and videos uploaded through Immich directly on a NAS while keeping the database in Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Immich, NAS uploads, UPLOAD_LOCATION, NFS, SMB, photo backup
app_version: "1.0.37"
doc_version: "1.0"
doc_updated: "2026-10-28"
---

# Store Immich uploads on a NAS

Immich backs up and organizes photos and videos uploaded from its web and mobile apps. You can save these uploads directly to a NAS while keeping the database in Olares. The selected media folder also holds generated files such as thumbnails and transcoded videos.

For photos already stored and managed in existing NAS folders, see [Import photos from a NAS into Immich](immich-import-from-nas.md). That workflow adds an external library without changing the destination of Immich uploads.

## Learning objectives

In this guide, you will learn how to:

- Set a NAS folder as the destination for Immich uploads.
- Move an existing media library between NAS and local storage.
- Recover access when NAS storage is unavailable.

## Prerequisites

- Use Immich Olares package 1.0.37 (Immich 3.3.5) or a later package that provides `UPLOAD_LOCATION`. If the setting is missing, the installed package does not support this feature yet.
- Connect Olares and the NAS to the same network.
- Prepare an NFS or SMB share with read/write access. NFS permissions must allow UID/GID `1000:1000` to read and write. For SMB, mount the share with a NAS account that has read/write access.

Set up the destination before uploading your first photos. For an existing Immich library, complete [Move existing media](#move-existing-media) before changing the setting.

## Mount the NAS share

1. Open Files, go to **Drive**, then **External**, and select **Connect to server**.
2. Select **NFS** and enter your share address. For example:

   ```text
   192.168.1.100:/volume1/photos
   ```

   Replace the IP address and share path with your NAS configuration. For SMB, select **SMB**, enter `//<NAS-IP>/<share>`, and provide an account with read/write access when prompted. See [Mount NFS shares](../manual/olares/files/mount-nfs.md) for the full NFS procedure.
3. Select **Confirm**. Check the folder name shown under **External**. The following steps use `photos` as the mounted folder name.
4. Create an empty folder named `immich` inside the mounted share and check that you can create a file in it. You can choose a different name, but use that name in the upload path too.

Use a dedicated folder for Immich-managed media. Keep unrelated NAS files outside it.

## Set the upload location

1. Open Settings, select **Applications**, then **Immich**.
2. Under **Environment variables**, open **Manage environment variables**.
3. Select the edit icon for `UPLOAD_LOCATION` and enter the media-folder path:

   ```text
   /external_storage/photos/immich
   ```

   Replace `photos` with the mounted folder name shown in Files and `immich` with the dedicated folder you created. Use this path format, not the NAS IP address, NFS export address, or SMB address.
4. Select **Confirm**, then **Apply**. If Immich is running, it restarts to apply the change. If you stopped it for migration, return to its application settings and select **Resume** after the change is applied.

| `UPLOAD_LOCATION` value | Upload destination |
| --- | --- |
| Empty | Default local media folder in Olares |
| An existing folder under `/external_storage/` | The selected NFS or SMB media folder |

Only this setting is needed. Mounting and unmounting the share itself is managed separately in Files.

## Check an upload

1. Open Immich and upload a new photo.
2. Check that its thumbnail appears and that you can open and download the original.
3. In Files, open the dedicated NAS folder. Check that new media files appear inside Immich's subfolders, such as `upload` or `library`.

Immich manages the folder layout and stored filenames. A stored filename can differ from the name shown in the photo details. The live database continues to use Olares storage.

## Move existing media

Changing `UPLOAD_LOCATION` selects a different media folder. It does not copy or synchronize files.

1. Pause uploads from all connected devices and wait for active jobs to finish. Create a database backup from Immich’s administration interface while the app is running. Follow the [database backup instructions](https://docs.immich.app/administration/backup-and-restore/#creating-a-backup). Keep a copy outside the media folder you are moving.
2. In Settings, select **Applications**, then **Immich**, and select **Stop**. Wait until **Resume** is available before copying files.
3. Back up the complete media folder. For this package, the default local folder is in Files under **Drive**, then **Data**, at `immich/data/upload`.
4. Copy the complete media folder contents to the new destination, including hidden `.immich` files and all subfolders. Check that the destination contains all originals and generated media by comparing file counts, sizes, and checksums.
5. [Set the upload location](#set-the-upload-location) to the destination, apply the change, and resume Immich.
6. Open existing photos, download an original, and test a new upload before removing any old copy.

Keep the same media-root layout at the destination. Do not add an extra directory level between the configured path and Immich's `upload`, `library`, `thumbs`, `encoded-video`, `profile`, and `backups` folders.

To return to default local storage, follow the same backup and stop procedure, then copy the complete media folder to the default local folder. Clear `UPLOAD_LOCATION`, apply the change, and resume Immich. Before switching back to NAS, repeat the procedure to copy the current media folder, including uploads added while using local storage.

## Recover when NAS storage is unavailable

If a configured NAS folder is missing at startup, Immich cannot start. It does not silently switch uploads to local storage.

- To keep using NAS, restore the mount in Files and check that the configured folder exists and is writable. Leave `UPLOAD_LOCATION` unchanged and allow startup to retry. If Immich remains unavailable, contact support.
- To use local storage, stop Immich if it is still running, then restore a complete media backup to the default local folder. Clear `UPLOAD_LOCATION` and apply the change. If Immich is stopped, select **Resume**. Local mode does not need a NAS mount, but photos whose files have not been restored remain unavailable.

:::info Market status after recovery
Olares Market might still show a failure after access to Immich is restored, including after switching to local storage. Check whether you can open Immich and access your photos. If the status remains incorrect, contact support. Do not delete the library or database to clear it.
:::

## Learn more

- [Self-host your photos with Immich](immich.md): Set up Immich and use its photo-management features.
- [Import photos from a NAS into Immich](immich-import-from-nas.md): Scan existing NAS folders as an external library.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change application configuration in Olares Settings.
