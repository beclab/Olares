---
outline: [2, 3]
description: Store photos and videos uploaded through Immich directly on a NAS while keeping the database in Olares.
head:
  - - meta
    - name: keywords
      content: Olares, Immich, NAS uploads, UPLOAD_LOCATION, NFS, SMB, photo backup
app_version: "1.0.37"
doc_version: "1.0"
doc_updated: "2026-10-27"
---

# Store Immich uploads on a NAS

Immich backs up and organizes photos and videos uploaded from its web and mobile apps. You can save these uploads directly to a NAS while keeping the database in Olares. The selected media folder also holds generated files such as thumbnails and transcoded videos.

For photos already stored and managed in existing NAS folders, see [Import photos from a NAS into Immich](immich-import-from-nas.md). That workflow adds an external library; it does not change the destination of Immich uploads.

## Before you begin

- Use Immich Olares package 1.0.37 (Immich 3.3.5) or a later package that provides `UPLOAD_LOCATION`. If the setting is missing, the installed package does not support this feature yet.
- Connect Olares and the NAS to the same network.
- Prepare an NFS or SMB share with read/write access. NFS permissions must allow UID/GID `1000:1000` to read and write. For SMB, mount the share with a NAS account that has read/write access.

This guide uses the actual demonstration share `nfs_share2` and its dedicated media folder `immich-upgrade-20261027-Kk3SeD`. This is a test folder created for the demonstration, not a fixed name required by Immich. The date identifies the test and the random suffix avoids name conflicts. The addresses and folder names are examples only; replace them with your own configuration.

Set up the destination before uploading your first photos. For an existing Immich library, complete [Move existing media](#move-existing-media) before changing the setting.

## Mount the NAS share

1. Open Files, go to **Drive** > **External**, and click **Connect to server**.
2. Select **NFS** and enter your share address. The actual demonstration environment uses:

   ```text
   192.168.50.124:/volume1/nfs_share2
   ```

   Replace this example with your NAS address and NFS share path. For SMB, select **SMB**, enter `//<NAS-IP>/<share>`, and provide an account with read/write access when prompted. See [Mount NFS shares](../manual/olares/files/mount-nfs.md) for the full NFS procedure.
3. Confirm the connection. Check the folder name shown under **External**. The demonstration uses `nfs_share2`.
4. First create an empty, dedicated media folder inside the mounted share and check that you can create a file in it. Choose any suitable name, such as `immich`. The demonstration's test folder is named `immich-upgrade-20261027-Kk3SeD`.

Use a dedicated folder for Immich-managed media. Keep unrelated NAS files outside it.

## Set the upload location

1. Open **Settings** > **Applications** > **Immich**.
2. Under **Environment variables**, open **Manage environment variables**.
3. Edit `UPLOAD_LOCATION` and enter the full media-folder path. The actual demonstration environment uses:

   ```text
   /external_storage/nfs_share2/immich-upgrade-20261027-Kk3SeD
   ```

   Here, `nfs_share2` is the mounted folder name in the demonstration, and `immich-upgrade-20261027-Kk3SeD` is the dedicated media folder created beforehand. Enter the full path for your own mounted share and media folder. Do not copy the demonstration value unchanged or enter the NAS IP address here.

   For example, if you create a folder named `immich` inside the `nfs_share2` share, enter:

   ```text
   /external_storage/nfs_share2/immich
   ```
4. Click **Confirm**, then **Apply**. Wait for Immich to restart.

Enter the path inside Immich, starting with `/external_storage/`. Do not enter an SMB address such as `//<NAS-IP>/<share>` or an NFS export address here.

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

1. Stop Immich from its application settings.
2. Back up both the PostgreSQL database and the complete media folder. See Immich's [backup and restore guide](https://docs.immich.app/administration/backup-and-restore/).
3. Copy the complete media folder contents to the new destination, including hidden `.immich` files and all subfolders. For the default local storage in this package, the media folder is **Files** > **Drive** > **Data** > **immich** > **data** > **upload**.
4. Check that the destination contains all files, including originals and generated media. Compare file counts, sizes and checksums before switching.
5. Set `UPLOAD_LOCATION` to the destination and apply the change. Open existing photos, download an original, and test a new upload before removing any old copy.

Keep the same media-root layout at the destination. Do not add an extra directory level between the configured path and Immich's `upload`, `library`, `thumbs`, `encoded-video`, `profile`, and `backups` folders.

To return to default local storage, restore the complete media folder there before clearing `UPLOAD_LOCATION`. Copy any uploads created while using local storage before switching back to NAS.

## Recover when NAS storage is unavailable

If a configured NAS folder is missing at startup, Immich cannot start. It does not silently switch uploads to local storage.

- To keep using NAS, restore the mount in Files and check the configured folder. Apply the upload-location setting after the share is available.
- To use local storage, restore the media from a backup to the default local folder, clear `UPLOAD_LOCATION`, and apply the change. The NAS does not need to be mounted for local mode to start. NAS-only photos remain unavailable until their files are restored.

:::info Market status after recovery
Olares Market can retain a failed status after Immich itself becomes reachable. This was observed after a missing-mount failure, including after switching to local storage. The service can work while this status is stale. If it persists, contact support; do not delete the library or database to clear it.
:::

## Learn more

- [Self-host your photos with Immich](immich.md): Set up Immich and use its photo-management features.
- [Import photos from a NAS into Immich](immich-import-from-nas.md): Scan existing NAS folders as an external library.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change application configuration in Olares Settings.
